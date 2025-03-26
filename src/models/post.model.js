import { neo4jDriver } from "../initializers/db.js"

export class Post {
  /**
   * @param {object} post
   * @param {string} post.client_username
   * @param {string} post.client_username
   * @param {string[]} post.media_urls
   * @param {string[]} post.mentions
   * @param {string[]} post.hashtags
   * @param {"photo" | "video" | "reel" | "story"} post.type
   * @param {string} post.description
   */
  static async create({
    client_username,
    media_urls,
    type,
    description,
    mentions,
    hashtags,
  }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let mention_notifs = []
      let new_post_data = null

      const { records: postRecords } = await tx.run(
        `
        MATCH (clientUser:User{ username: $client_username })
        CREATE (clientUser)-[:CREATES_POST]->(post:Post{ id: randomUUID(), type: $type, media_urls: $media_urls, description: $description, created_at: datetime(), reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0 })
        WITH post, toString(post.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user
        RETURN post { .*, created_at, owner_user, client_reaction: "", client_reposted: false, client_saved: false } AS new_post_data
        `,
        { client_username, media_urls, type, description }
      )

      new_post_data = postRecords[0]?.get("new_post_data")

      if (mentions.length) {
        const { records: mentionRecords } = await tx.run(
          `
          MATCH (user:User WHERE user.username IN $mentions)
          RETURN collect(user.username) AS valid_mentions
          `,
          { mentions }
        )

        mentions = mentionRecords[0]?.get("valid_mentions")

        await tx.run(
          `
          MATCH (mentionUser:User WHERE mentionUser.username IN $mentions), (post:Post{ id: $postId })
          CREATE (post)-[:MENTIONS_USER]->(mentionUser)
          `,
          { mentions, postId: new_post_data.id }
        )

        const mentionsExcClient = mentions.filter(
          (uname) => uname != client_username
        )

        if (mentionsExcClient.length) {
          const { records } = await tx.run(
            `
            MATCH (mentionUser:User WHERE mentionUser.username IN $mentionsExcClient), (clientUser:User{ username: $client_username })

            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_post", is_read: false, created_at: datetime(), details: ["in_post_id", $postId], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

            WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
            RETURN collect(mentionNotif { .*, created_at, receiver_username }) AS mention_notifs
            `,
            { mentionsExcClient, postId: new_post_data.id, client_username }
          )

          mention_notifs = records[0]?.get("mention_notifs")
        }
      }

      await tx.run(
        `
        MATCH (post:Post{ id: $postId })

        UNWIND $hashtags AS hashtagName
        MERGE (ht:Hashtag{name: hashtagName})
        CREATE (post)-[:INCLUDES_HASHTAG]->(ht)
        `,
        { hashtags, postId: new_post_data.id }
      )

      return { mention_notifs, new_post_data }
    })

    session.close()

    return res
  }

  static async repost(origin_post_id, client_username) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let repost_data = null
      let latest_reposts_count = 0
      let repost_notif = null

      const { records: repostRecords } = await tx.run(
        `
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $origin_post_id })<-[:CREATES_POST]-(originPostOwner)

        MERGE (clientUser)-[:CREATES_REPOST]->(repost:Repost:Post{ reposter_username: $client_username, origin_post_id: $origin_post_id })-[:REPOST_OF]->(post)
        ON CREATE
          SET repost += { id: randomUUID(), type: post.type, media_urls: post.media_urls, description: post.description, created_at: datetime(), reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0 },
            post.reposts_count = post.reposts_count + 1

        WITH post, repost, toString(repost.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user, originPostOwner.username AS origin_post_owner_username

        RETURN post.reposts_count AS latest_reposts_count,
          repost { .*, created_at, owner_user, client_reaction: "", client_reposted: false, client_saved: false } AS repost_data,
          origin_post_owner_username
        `,
        {
          origin_post_id,
          client_username,
        }
      )

      latest_reposts_count = repostRecords[0]?.get("latest_reposts_count")
      repost_data = repostRecords[0]?.get("repost_data")
      const origin_post_owner_username = repostRecords[0]?.get("origin_post_owner_username")

      if (origin_post_owner_username !== client_username) {
        const { records: repostNotifRecords } = await tx.run(
          `
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $origin_post_id })<-[:CREATES_POST]-(postOwner)

        CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(repostNotif:Notification:RepostNotification{ id: randomUUID(), type: "repost", is_read: false, created_at: datetime(), details: ["repost_id", $repostId, "origin_post_id", $origin_post_id], reposter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
        WITH repostNotif, toString(repostNotif.created_at) AS created_at, postOwner.username AS receiver_username
        RETURN repostNotif { .*, created_at, receiver_username } AS repost_notif
        `,
          { origin_post_id, client_username, repostId: repost_data.id }
        )

        repost_notif = repostNotifRecords[0]?.get("repost_notif")
      }

      return { repost_data, latest_reposts_count, repost_notif }
    })

    session.close()

    return res
  }

  static async save(post_id, client_username) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })
      MERGE (clientUser)-[:SAVES_POST]->(post)
      ON CREATE
        SET post.saves_count = post.saves_count + 1

      RETURN post.saves_count AS latest_saves_count
      `,
      {
        post_id,
        client_username,
      }
    )

    return records[0]?.toObject() ?? { latest_saves_count: null }
  }

  static async reactTo({ client_username, post_id, reaction }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let latest_reactions_count = 0
      let reaction_notif = null

      const { records: reactionRecords } = await tx.run(
        `
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)

        MERGE (clientUser)-[crxn:REACTS_TO_POST]->(post)
        ON CREATE
          SET crxn.reaction = $reaction,
            crxn.at = datetime(),
            post.reactions_count = post.reactions_count + 1

        RETURN post.reactions_count AS latest_reactions_count, postOwner.username AS post_owner_username
        `,
        {
          post_id,
          client_username,
          reaction,
        }
      )

      latest_reactions_count = reactionRecords[0]?.get("latest_reactions_count")
      const post_owner_username = reactionRecords[0]?.get("post_owner_username")

      if (post_owner_username !== client_username) {
        const { records: reactionNotifRecords } = await tx.run(
          `
          MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
  
          CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_post", is_read: false, created_at: datetime(), details: ["reaction", $reaction, "to_post_id", $post_id], reactor_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
  
          WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, postOwner.username AS receiver_username
          RETURN reactionNotif { .*, created_at, receiver_username } AS reaction_notif
          `,
          { post_id, client_username, reaction }
        )

        reaction_notif = reactionNotifRecords[0]?.get("reaction_notif")
      }

      return { reaction_notif, latest_reactions_count }
    })

    session.close()

    return res
  }

  static async commentOn({
    post_id,
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
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
        CREATE (clientUser)-[:WRITES_COMMENT]->(comment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url, reactions_count: 0, comments_count: 0, created_at: datetime() })-[:COMMENT_ON_POST]->(post)

        WITH post, comment, toString(comment.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user, postOwner.username AS post_owner_username
        
        SET post.comments_count = post.comments_count + 1

        RETURN post.comments_count AS latest_comments_count,
          comment { .*, created_at, owner_user, client_reaction: "" } AS new_comment_data,
          post_owner_username
        `,
        { client_username, attachment_url, comment_text, post_id }
      )

      new_comment_data = commentRecords[0]?.get("new_comment_data")
      latest_comments_count = commentRecords[0]?.get("latest_comments_count")
      const post_owner_username = commentRecords[0]?.get("post_owner_username")

      if (mentions.length) {
        await tx.run(
          `
          MATCH (mentionUser:User WHERE mentionUser.username IN $mentions), (comment:Comment{ id: $commentId })

          CREATE (comment)-[:MENTIONS_USER]->(mentionUser)
          `,
          { mentions, commentId: new_comment_data.id }
        )

        const mentionsExcClient = mentions.filter(
          (uname) => uname != client_username
        )

        if (mentionsExcClient.length) {
          const { records } = await tx.run(
            `
            MATCH (mentionUser:User WHERE mentionUser.username IN $mentionsExcClient), (clientUser:User{ username: $client_username })

            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", is_read: false, created_at: datetime(), details: ["in_comment_id", $commentId], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

            WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
            RETURN collect(mentionNotif { .*, receiver_username, created_at }) AS mention_notifs
            `,
            {
              mentionsExcClient,
              commentId: new_comment_data.id,
              client_username,
            }
          )

          mention_notifs = records[0]?.get("mention_notifs")
        }
      }

      await tx.run(
        `
        MATCH (comment:Comment{ id: $commentId })

        UNWIND $hashtags AS hashtagName
        MERGE (ht:Hashtag{name: hashtagName})
        CREATE (comment)-[:INCLUDES_HASHTAG]->(ht)
        `,
        { hashtags, commentId: new_comment_data.id }
      )

      // comment notif
      if (post_owner_username !== client_username) {
        const { records: commentNotifRecords } = await tx.run(
          `
          MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
          
          CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_post", is_read: false, created_at: datetime(), details: ["on_post_id", $post_id, "comment_id", $commentId, "comment_text", $comment_text, "attachment_url", $attachment_url], commenter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

          WITH commentNotif, 
            toString(commentNotif.created_at) AS created_at, 
            postOwner.username AS receiver_username
          RETURN commentNotif { .*, created_at, receiver_username } AS comment_notif
          `,
          {
            client_username,
            post_id,
            commentId: new_comment_data.id,
            comment_text,
            attachment_url,
          }
        )

        comment_notif = commentNotifRecords[0]?.get("comment_notif")
      }

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

  static async findOne(post_id, client_username) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser:User), (clientUser:User{ username: $client_username })

      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      
      WITH post, 
        toString(post.created_at) AS created_at, 
        ownerUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        CASE csaves 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_reposted
      RETURN post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted } AS found_post
      `,
      { post_id, client_username }
    )

    return records[0]?.get("found_post")
  }

  static async getComments({ post_id, client_username, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (post:Post{ id: $post_id })<-[:COMMENT_ON_POST]-(comment:Comment)<-[:WRITES_COMMENT]-(ownerUser:User)

      OPTIONAL MATCH (comment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })

      WITH comment, 
        toString(comment.created_at) AS created_at, 
        ownerUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction
      ORDER BY comment.created_at DESC, comment.reactions_count DESC, comment.comments_count DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(comment {.*, owner_user, created_at, client_reaction }) AS res_comments
      `,
      { post_id, client_username, limit, offset }
    )

    return records[0]?.get("res_comments")
  }

  static async getReactors({ post_id, client_username, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (:Post{ id: $post_id })<-[rxn:REACTS_TO_POST]-(reactor:User)
      OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
      WITH reactor, 
        rxn, 
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY rxn.at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_rxn
      `,
      { post_id, client_username, limit, offset }
    )

    return records[0]?.get("reactors_rxn")
  }

  static async getReactorsWithReaction({
    post_id,
    reaction,
    client_username,
    limit,
    offset,
  }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (post:Post{ id: $post_id })<-[rxn:REACTS_TO_POST { reaction: $reaction }]-(reactor:User)
      OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
      WITH reactor, 
        rxn, 
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY rxn.at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_rxn
      `,
      { post_id, client_username, reaction, limit, offset }
    )

    return records[0]?.get("reactors_rxn")
  }

  static async delete(post_id, client_username) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ username: $client_username })

      OPTIONAL MATCH (clientUser)-[:CREATES_POST]->(post:Post{ id: $post_id }),
        (clientUser)-[:CREATES_REPOST]->(repost:Repost{ id: $post_id })

      DETACH DELETE post, repost
      `,
      { post_id, client_username }
    )
  }

  static async removeReaction(post_id, client_username) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (:User{ username: $client_username })-[rxn:REACTS_TO_POST]->(post:Post{ id: $post_id })
      DELETE rxn

      SET post.reactions_count = post.reactions_count - 1

      RETURN post.reactions_count AS latest_reactions_count
      `,
      {
        post_id,
        client_username,
      }
    )

    return records[0]?.toObject() ?? { latest_reactions_count: null }
  }

  static async removeComment({ post_id, comment_id, client_username }) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ username: $client_username })-[:WRITES_COMMENT]->(comment:Comment{ id: $comment_id })-[:COMMENT_ON_POST]->(post:Post{ id: $post_id })
      DETACH DELETE comment

      SET post.comments_count = post.comments_count - 1

      RETURN post.comments_count AS latest_comments_count
      `,
      { post_id, comment_id, client_username }
    )

    return records[0]?.toObject() ?? { latest_comments_count: null }
  }

  static async unrepost(post_id, client_username) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (:User{ username: $client_username })-[:CREATES_REPOST]->(repost)-[:REPOST_OF]->(post:Post{ id: $post_id })
      DETACH DELETE repost

      SET post.reposts_count = post.reposts_count - 1

      RETURN post.reposts_count AS latest_reposts_count
      `,
      { client_username, post_id }
    )

    return records[0]?.toObject() ?? { latest_reposts_count: null }
  }

  static async unsave(post_id, client_username) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (:User{ username: $client_username })-[csave:SAVES_POST]->(post:Post{ id: $post_id })

      DELETE csave

      SET post.saves_count = post.saves_count - 1

      RETURN post.saves_count AS latest_saves_count
      `,
      {
        post_id,
        client_username,
      }
    )

    return records[0]?.toObject() ?? { latest_saves_count: null }
  }
}
