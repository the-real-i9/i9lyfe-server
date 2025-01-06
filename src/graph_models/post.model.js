import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

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
        WITH post, toString(post.created_at) AS created_at, clientUser { .id, .username, .profile_pic_url } AS owner_user
        RETURN post { .*, created_at, owner_user, client_reaction: "", client_reposted: false, client_saved: false } AS new_post_data
        `,
        { client_username, media_urls, type, description }
      )

      new_post_data = postRecords[0].toObject().new_post_data

      if (mentions.length) {
        const { records: mentionRecords } = await tx.run(
          `
          MATCH (user:User WHERE user.username IN $mentions)
          RETURN collect(user.username) AS valid_mentions
          `,
          { mentions }
        )

        mentions = mentionRecords[0].toObject().valid_mentions

        await tx.run(
          `
          UNWIND $mentions AS mentionUsername
          MATCH (mentionUser:User{ username: mentionUsername }), (post:Post{ id: $postId })
          CREATE (post)-[:MENTIONS]->(mentionUser)
          `,
          { mentions, postId: new_post_data.id }
        )

        const mentionsExcClient = mentions.filter(
          (uname) => uname != client_username
        )

        if (mentionsExcClient.length) {
          const { records } = await tx.run(
            `
            UNWIND $mentionsExcClient AS mentionUsername
            MATCH (mentionUser:User{ username: mentionUsername }), (post:Post{ id: $postId }), (clientUser:User{ username: $client_username })
            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_post", in_post_id: post.id })-[:MENTIONING_USER]->(clientUser)
            WITH mentionUser, mentionNotif, clientUser { .id, .username, .profile_pic_url } AS clientUserView
            RETURN [notif IN collect(mentionNotif) | notif { .*, receiver_user_id: mentionUser.id, mentioning_user: clientUserView }] AS mention_notifs
            `,
            { mentionsExcClient, postId: new_post_data.id, client_username }
          )

          mention_notifs = records[0].toObject().mention_notifs
        }
      }

      await tx.run(
        `
        UNWIND $hashtags AS hashtagName
        MATCH (post:Post{ id: $postId })
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

  static async repost(original_post_id, client_user_id) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let repost_data = null
      let latest_reposts_count = 0
      let repost_notif = null

      const { records: repostRecords } = await tx.run(
        `
        MATCH (post:Post{ id: $original_post_id}), (clientUser:User{ id: $client_user_id })
        CREATE (clientUser)-[:CREATES_REPOST { user_to_post: $user_to_post }]->(repost:Repost:Post{ id: randomUUID(), type: post.type, media_urls: post.media_urls, description: post.description, created_at: datetime(), reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0 })-[:REPOST_OF]->(post)

        WITH post, repost, toString(repost.created_at) AS created_at, clientUser { .id, username, .profile_pic_url } owner_user

        SET post.reposts_count = post.reposts_count + 1

        RETURN post.reposts_count AS latest_reposts_count,
          repost { .*, created_at, owner_user, client_reaction: "", client_reposted: false, client_saved: false } AS repost_data
        `,
        {
          original_post_id,
          client_user_id,
          user_to_post: `user-${client_user_id}_to_post-${original_post_id}`,
        }
      )

      const rco = repostRecords[0].toObject()

      latest_reposts_count = rco.latest_reposts_count
      repost_data = rco.repost_data

      const { records: repostNotifRecords } = await tx.run(
        `
        MATCH (post:Post{ id: $original_post_id}), (clientUser:User{ id: $client_user_id })
        MATCH (postOwner:User WHERE postOwner.id <> $client_user_id)-[:CREATES_POST]->(post)
        CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(repostNotif:Notification:RepostNotification{ id: randomUUID(), type: "repost", repost_id: $repostId, is_read: false, created_at: datetime() })-[:REPOSTER_USER]->(clientUser)
        WITH repostNotif, toString(repostNotif.created_at) AS created_at, postOwner.id AS receiver_user_id, clientUser { .id, .username, .profile_pic_url } AS reposter_user
        RETURN repostNotif { .*, created_at, receiver_user_id, reposter_user } AS repost_notif
        `,
        { original_post_id, client_user_id, repostId: repost_data.id }
      )

      repost_notif = repostNotifRecords[0]?.get("repost_notif")

      return { repost_data, latest_reposts_count, repost_notif }
    })

    session.close()

    return res
  }

  static async save(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (post:Post{ id: $post_id }), (clientUser:User{ id: $client_user_id })
      CREATE (clientUser)-[:SAVES_POST { user_to_post: $user_to_post }]->(post)

      SET post.saves_count = post.saves_count + 1

      RETURN post.saves_count AS latest_saves_count
      `,
      {
        post_id,
        client_user_id,
        user_to_post: `user-${client_user_id}_to_post-${post_id}`,
      }
    )

    return records[0].toObject()
  }

  static async reactTo({ client_user_id, post_id, reaction }) {
    const session = neo4jDriver.session()

    const res = await session.executeWrite(async (tx) => {
      let latest_reactions_count = 0
      let reaction_notif = null

      const { records: reactionRecords } = await tx.run(
        `
        MATCH (post:Post{ id: $post_id }), (clientUser:User{ id: $client_user_id })
        CREATE (clientUser)-[:REACTS_TO_POST { user_to_post: $user_to_post, reaction: $reaction }]->(post)

        SET post.reactions_count = post.reactions_count + 1

        RETURN post.reactions_count AS latest_reactions_count
        `,
        {
          post_id,
          client_user_id,
          user_to_post: `user-${client_user_id}_to_post-${post_id}`,
          reaction
        }
      )

      latest_reactions_count = reactionRecords[0].get("latest_reactions_count")

      const { records: reactionNotifRecords } = await tx.run(
        `
        MATCH (post:Post{ id: $post_id }), (clientUser:User{ id: $client_user_id })
        WITH post, clientUser
        MATCH (postOwner:User WHERE postOwner.id <> $client_user_id)-[:CREATES_POST]->(post)
        CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_post", reaction: $reaction, is_read: false, created_at: datetime() })-[:REACTOR_USER]->(clientUser)
        WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, postOwner.id AS receiver_user_id, clientUser {.id, .username, .profile_pic_url} AS reactor_user
        RETURN reactionNotif { .*, created_at, receiver_user_id, reactor_user } AS reaction_notif
        `,
        { post_id, client_user_id, reaction }
      )

      reaction_notif = reactionNotifRecords[0]?.get("reaction_notif")

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
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })
        CREATE (clientUser)-[:WRITES_COMMENT]->(comment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url, reactions_count: 0, comments_count: 0, created_at: datetime() })-[:COMMENT_ON]->(post)

        WITH post, comment, toString(comment.created_at) AS created_at, clientUser { .id, .username, .profile_pic_url } AS owner_user
        
        SET post.comments_count = post.comments_count + 1

        RETURN post.comments_count AS latest_comments_count,
        comment { .*, created_at, owner_user, client_reaction: "" } AS new_comment_data
        `,
        { client_username, attachment_url, comment_text, post_id }
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
          MATCH (mentionUser:User{ username: mentionUsername }), (comment:Comment{ id: $commentId })
          CREATE (comment)-[:MENTIONS]->(mentionUser)
          `,
          { mentions, commentId: new_comment_data.id }
        )

        const mentionsExcClient = mentions.filter(
          (uname) => uname != client_username
        )

        if (mentionsExcClient.length) {
          const { records } = await tx.run(
            `
            UNWIND $mentionsExcClient AS mentionUsername
            MATCH (mentionUser:User{ username: mentionUsername }), (comment:Comment{ id: $commentId }), (clientUser:User{ username: $client_username })
            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", in_comment_id: comment.id })-[:MENTIONING_USER]->(clientUser)
            WITH mentionNotif, mentionUser.id AS receiver_user_id, clientUser { .id, .username, .profile_pic_url } AS mentioning_user
            RETURN [notif IN collect(mentionNotif) | notif { .*, receiver_user_id, mentioning_user }] AS mention_notifs
            `,
            {
              mentionsExcClient,
              commentId: new_comment_data.id,
              client_username,
            }
          )

          mention_notifs = records[0].get("mention_notifs")
        }
      }

      await tx.run(
        `
        UNWIND $hashtags AS hashtagName
        MATCH (comment:Comment{ id: $commentId })
        MERGE (ht:Hashtag{name: hashtagName})
        CREATE (comment)-[:INCLUDES_HASHTAG]->(ht)
        `,
        { hashtags, commentId: new_comment_data.id }
      )

      // comment notif
      const { records: commentNotifRecords } = await tx.run(
        `
          MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })
          MATCH (postOwner:User WHERE postOwner.username <> $client_username)-[:CREATES_POST]->(post)
          CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_post", comment_id: $commentId, is_read: false, created_at: datetime() })-[:COMMENTER_USER]->(clientUser)
          WITH commentNotif, toString(commentNotif.created_at) AS created_at, postOwner.id AS receiver_user_id, clientUser {.id, .username, .proifle_pic_url} commenter_user
          RETURN commentNotif { .*, created_at, receiver_user_id, commenter_user } AS comment_notif
          `,
        { client_username, post_id, commentId: new_comment_data.id }
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

  static async findOne(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeRead(
      `

      `,
      { post_id, client_user_id },
    )

    return records[0].get("found_post")
  }

  static async getComments({ post_id, client_username, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_comments_on_post($1, $2, $3, $4)",
      values: [post_id, client_username, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactors({ post_id, client_username, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_to_post($1, $2, $3, $4)",
      values: [post_id, client_username, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactorsWithReaction({
    post_id,
    reaction,
    client_username,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_with_reaction_to_post($1, $2, $3, $4, $5)",
      values: [post_id, reaction, client_username, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async delete(post_id, client_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:CREATES_POST]->(post:Post{ id: $post_id })
      DETACH DELETE post
      `,
      { post_id, client_user_id }
    )
  }

  static async removeReaction(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH ()-[rxn:REACTS_TO_POST { user_to_post: $user_to_post }]->(post)
      DELETE rxn

      SET post.reactions_count = post.reactions_count - 1

      RETURN post.reactions_count AS latest_reactions_count
      `,
      {
        post_id,
        client_user_id,
        user_to_post: `user-${client_user_id}_to_post-${post_id}`,
      }
    )

    return records[0].toObject()
  }

  static async removeComment({ post_id, comment_id, client_user_id }) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:WRITES_COMMENT]->(comment:Comment{ id: $comment_id })-[:COMMENT_ON]->(post:Post{ id: $post_id })
      DETACH DELETE comment

      SET post.comments_count = post.comments_count - 1

      RETURN post.comments_count AS latest_comments_count
      `,
      { post_id, comment_id, client_user_id }
    )

    return records[0].toObject()
  }

  static async unrepost(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH ()-[:CREATES_REPOST { user_to_post: $user_to_post }]->(repost)-[:REPOST_OF]->(post)
      DETACH DELETE repost

      SET post.reposts_count = post.reposts_count - 1

      RETURN post.reposts_count AS latest_reposts_count
      `,
      { user_to_post: `user-${client_user_id}_to_post-${post_id}` }
    )

    return records[0].toObject()
  }

  static async unsave(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH ()-[save:SAVES_POST { user_to_post: $user_to_post }]->(post)
      DELETE save

      SET post.saves_count = post.saves_count - 1

      RETURN post.saves_count AS latest_saves_count
      `,
      {
        post_id,
        client_user_id,
        user_to_post: `user-${client_user_id}_to_post-${post_id}`,
      }
    )

    return records[0].toObject()
  }
}
