import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class Comment {
  static async reactTo({ client_user_id, comment_id, reaction }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let latest_reactions_count = 0
      let reaction_notif = null

      const { records: reactionRecords } = await tx.run(
        `
            MATCH (comment:Comment{ id: $comment_id }), (clientUser:User{ id: $client_user_id })
            CREATE (clientUser)-[:REACTS_TO_COMMENT { user_to_comment: $user_to_comment, reaction: $reaction }]->(comment)

            SET comment.reactions_count = comment.reactions_count + 1

            RETURN comment.reactions_count AS latest_reactions_count
            `,
        {
          comment_id,
          client_user_id,
          user_to_comment: `user-${client_user_id}_to_comment-${comment_id}`,
          reaction,
        }
      )

      latest_reactions_count = reactionRecords[0].get("latest_reactions_count")

      const { records: reactionNotifRecords } = await tx.run(
        `
            MATCH (comment:Comment{ id: $comment_id }), (clientUser:User{ id: $client_user_id })
            WITH comment, clientUser
            MATCH (commentOwner:User WHERE commentOwner.id <> $client_user_id)-[:WRITES_COMMENT]->(comment)
            CREATE (commentOwner)-[:RECEIVES_NOTIFICATION]->(reactNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_comment", reaction: $reaction, is_read: false, created_at: datetime() })-[:REACTOR_USER]->(clientUser)
            WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, commentOwner.id AS receiver_user_id, clientUser {.id, .username, .profile_pic_url} AS reactor_user
            RETURN reactionNotif { .*, created_at, receiver_user_id, reactor_user } AS reaction_notif
            `,
        { comment_id, client_user_id, reaction }
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
        CREATE (clientUser)-[:WRITES_COMMENT]->(childComment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url,  reactions_count: 0, comments_count: 0, created_at: datetime() })-[:COMMENT_ON]->(comment)

        WITH parentComment, childComment, toString(childComment.created_at) AS created_at, clientUser { .id, .username, .profile_pic_url } AS owner_user

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
          CREATE (childComment)-[:MENTIONS]->(mentionUser)
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
            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", in_comment_id: childComment.id })-[:MENTIONING_USER]->(clientUser)
            WITH mentionUser, mentionNotif, clientUser { .id, .username, .profile_pic_url } AS clientUserView
            RETURN [notif IN collect(mentionNotif) | notif { .*, receiver_user_id: mentionUser.id, mentioning_user: clientUserView }] AS mention_notifs
            `,
            {
              mentionsExcClient,
              childCommentId: new_comment_data.id,
              client_username,
            }
          )

          mention_notifs = records[0].get("mention_notifs")
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
          CREATE (parentCommentOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_comment", child_comment_id: $childCommentId, is_read: false, created_at: datetime() })-[:COMMENTER_USER]->(clientUser)
          WITH parentCommentOwner, clientUser {.id, .username, .proifle_pic_url} clientUserView
          RETURN commentNotif { .*, created_at: (.created_at), receiver_user_id: parentCommentOwner.id, commenter_user: clientUserView } AS comment_notif
          `,
        { client_username, comment_id, childCommentId: new_comment_data.id }
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

  static async getComments({ comment_id, client_user_id, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_comments_on_comment($1, $2, $3, $4)",
      values: [comment_id, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async find(comment_id, client_user_id) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_comment($1, $2)",
      values: [comment_id, client_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async getReactors({ comment_id, client_user_id, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_to_comment($1, $2, $3, $4)",
      values: [comment_id, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactorsWithReaction({
    comment_id,
    reaction,
    client_user_id,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_with_reaction_to_comment($1, $2, $3, $4, $5)",
      values: [comment_id, reaction, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async removeReaction(comment_id, client_user_id) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH ()-[rxn:REACTS_TO_COMMENT { user_to_comment: $user_to_comment }]->(comment)
      DELETE rxn

      SET comment.reactions_count = comment.reactions_count - 1

      RETURN comment.reactions_count AS latest_reactions_count
      `,
      {
        comment_id,
        client_user_id,
        user_to_comment: `user-${client_user_id}_to_comment-${comment_id}`,
      }
    )

    return records[0].toObject()
  }

  static async removeChildComment({
    parent_comment_id,
    comment_id,
    client_user_id,
  }) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:WRITES_COMMENT]->(childComment:Comment{ id: $comment_id })-[:COMMENT_ON]->(parentComment:Comment{ id: $parent_comment_id })
      DETACH DELETE childComment

      SET parentComment.comments_count = parentComment.comments_count - 1

      RETURN parentComment.comments_count AS latest_comments_count
      `,
      { parent_comment_id, comment_id, client_user_id }
    )

    return records[0].toObject()
  }
}
