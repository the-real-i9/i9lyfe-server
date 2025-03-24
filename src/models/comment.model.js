import { neo4jDriver } from "../initializers/db.js"

export class Comment {
  static async reactTo({ client_username, comment_id, reaction }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let latest_reactions_count = 0
      let reaction_notif = null

      const { records: reactionRecords } = await tx.run(
        `
            MATCH (clientUser:User{ username: $client_username }), (comment:Comment{ id: $comment_id })
            MERGE (clientUser)-[crxn:REACTS_TO_COMMENT]->(comment)
            ON CREATE
              SET crxn.reaction = $reaction,
                crxn.at = datetime(),
                comment.reactions_count = comment.reactions_count + 1

            RETURN comment.reactions_count AS latest_reactions_count
            `,
        {
          comment_id,
          client_username,
          reaction,
        }
      )

      latest_reactions_count = reactionRecords[0].get("latest_reactions_count")

      const { records: reactionNotifRecords } = await tx.run(
        `
            MATCH (comment:Comment{ id: $comment_id }), (clientUser:User{ username: $client_username })
            WITH comment, clientUser
            MATCH (commentOwner:User WHERE commentOwner.id <> $client_username)-[:WRITES_COMMENT]->(comment)
            CREATE (commentOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_comment", is_read: false, created_at: datetime(), details: ["reaction", $reaction, "to_comment_id", $comment_id], reactor_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
            WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, commentOwner.id AS receiver_username
            RETURN reactionNotif { .*, created_at, receiver_username } AS reaction_notif
            `,
        { comment_id, client_username, reaction }
      )

      reaction_notif = reactionNotifRecords[0]?.get("reaction_notif")

      return { reaction_notif, latest_reactions_count }
    })

    session.close()

    return res
  }

  static async commentOn({
    comment_id,
    client_username,
    comment_text,
    attachment_url,
    mentions,
    hashtags,
  }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let mention_notifs = []
      let new_comment_data = null
      let comment_notif = null
      let latest_comments_count = 0

      const { records: commentRecords } = await tx.run(
        `
        MATCH (clientUser:User{ username: $client_username }), (parentComment:Comment{ id: $comment_id })
        CREATE (clientUser)-[:WRITES_COMMENT]->(childComment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url,  reactions_count: 0, comments_count: 0, created_at: datetime() })-[:COMMENT_ON_COMMENT]->(parentComment)

        WITH parentComment, childComment, toString(childComment.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user

        SET parentComment.comments_count = parentComment.comments_count + 1

        RETURN parentComment.comments_count AS latest_comments_count,
        childComment { .*, created_at, owner_user, client_reaction: "" } AS new_comment_data
        `,
        { client_username, attachment_url, comment_text, comment_id }
      )

      const cro = commentRecords[0].toObject()

      new_comment_data = cro.new_comment_data
      latest_comments_count = cro.latest_comments_count

      if (mentions.length) {
        const { records: mentionRecords } = await tx.run(
          `
          MATCH (user:User WHERE user.username IN $mentions)
          RETURN collect(user.username) AS valid_mentions
          `,
          { mentions }
        )

        mentions = mentionRecords[0].get("valid_mentions")

        await tx.run(
          `
          UNWIND $mentions AS mentionUsername
          MATCH (mentionUser:User{ username: mentionUsername }), (childComment:Comment{ id: $childCommentId })
          CREATE (childComment)-[:MENTIONS_USER]->(mentionUser)
          `,
          { mentions, childCommentId: new_comment_data.id }
        )

        const mentionsExcClient = mentions.filter(
          (uname) => uname != client_username
        )

        if (mentionsExcClient.length) {
          const { records } = await tx.run(
            `
            UNWIND $mentionsExcClient AS mentionUsername
            MATCH (mentionUser:User{ username: mentionUsername }), (childComment:Comment{ id: $childCommentId }), (clientUser:User{ username: $client_username })
            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", is_read: false, created_at: datetime(), details: ["in_comment_id", childComment.id], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
            WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
            RETURN collect(mentionNotif { .*, created_at, receiver_username }) AS mention_notifs
            `,
            {
              mentionsExcClient,
              childCommentId: new_comment_data.id,
              client_username,
            }
          )

          mention_notifs = records[0]?.get("mention_notifs")
        }
      }

      await tx.run(
        `
        UNWIND $hashtags AS hashtagName
        MATCH (childComment:Comment{ id: $childCommentId })
        MERGE (ht:Hashtag{name: hashtagName})
        CREATE (childComment)-[:INCLUDES_HASHTAG]->(ht)
        `,
        { hashtags, childCommentId: new_comment_data.id }
      )

      // comment notif
      const { records: commentNotifRecords } = await tx.run(
        `
          MATCH (clientUser:User{ username: $client_username }), (parentComment:Comment{ id: $comment_id })
          MATCH (parentCommentOwner:User WHERE parentCommentOwner.username <> $client_username)-[:WRITES_COMMENT]->(parentComment)
          CREATE (parentCommentOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_comment", is_read: false, created_at: datetime(), details: ["on_comment_id", $comment_id, "child_comment_id", $childCommentId, "comment_text", $comment_text, "attachment_url", $attachment_url], commenter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
          WITH commentNotif, toString(commentNotif.created_at) AS created_at, parentCommentOwner.id AS receiver_username
          RETURN commentNotif { .*, created_at, receiver_username } AS comment_notif
          `,
        { client_username, comment_id, childCommentId: new_comment_data.id, comment_text, attachment_url }
      )

      comment_notif = commentNotifRecords[0]?.get("comment_notif")

      return {
        mention_notifs,
        new_comment_data,
        comment_notif,
        latest_comments_count,
      }
    })

    session.close()

    return res
  }

  static async findOne(comment_id, client_username) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (comment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(ownerUser:User), (clientUser:User{ username: $client_username })
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_COMMENT]->(comment)
      WITH comment, 
        toString(comment.created_at) AS created_at, 
        ownerUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction
      RETURN comment { .*, owner_user, created_at, client_reaction } AS found_comment
      `,
      { comment_id, client_username },
    )

    return records[0]?.get("found_comment")
  }

  static async getComments({ comment_id, client_username, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (parentComment:Comment{ id: $comment_id })<-[:COMMENT_ON_COMMENT]-(childComment:Comment)<-[:WRITES_COMMENT]-(ownerUser:User)
      OPTIONAL MATCH (childComment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })
      WITH childComment, 
        toString(childComment.created_at) AS created_at, 
        ownerUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction
      ORDER BY childComment.created_at DESC, childComment.reactions_count DESC, childComment.comments_count DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(childComment {.*, created_at, owner_user, client_reaction }) AS res_comments
      `,
      { comment_id, client_username, limit, offset }
    )
    
    return records[0]?.get("res_comments")
  }


  static async getReactors({ comment_id, client_username, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT]-(reactor:User)
      OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
      WITH reactor, 
        rxn, 
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY rxn.at DESC
      SKIP toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(reactor { .id, .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_rxn
      `,
      { comment_id, client_username, limit, offset }
    )

    return records[0]?.get("reactors_rxn")
  }

  static async getReactorsWithReaction({
    comment_id,
    reaction,
    client_username,
    limit,
    offset,
  }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT { reaction: $reaction }]-(reactor:User)
      OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
      WITH reactor, 
        rxn, 
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY rxn.at DESC
      SKIP toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(reactor { .id, .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_rxn
      `,
      { comment_id, client_username, reaction, limit, offset }
    )

    return records[0]?.get("reactors_rxn")
  }

  static async removeReaction(comment_id, client_username) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (:User{ username: $client_username })-[crxn:REACTS_TO_COMMENT]->(comment:Comment{ id: $comment_id })
      DELETE crxn

      SET comment.reactions_count = comment.reactions_count - 1

      RETURN comment.reactions_count AS latest_reactions_count
      `,
      {
        comment_id,
        client_username,
      }
    )

    return records[0].toObject()
  }

  static async removeChildComment({
    parent_comment_id,
    comment_id,
    client_username,
  }) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ username: $client_username })-[:WRITES_COMMENT]->(childComment:Comment{ id: $comment_id })-[:COMMENT_ON_COMMENT]->(parentComment:Comment{ id: $parent_comment_id })
      DETACH DELETE childComment

      SET parentComment.comments_count = parentComment.comments_count - 1

      RETURN parentComment.comments_count AS latest_comments_count
      `,
      { parent_comment_id, comment_id, client_username }
    )

    return records[0].toObject()
  }
}
