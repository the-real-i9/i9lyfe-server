import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

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
        CREATE (clientUser)-[:CREATES_POST]->(post:Post{ id: randomUUID(), type: $type, media_urls: $media_urls, description: $description, created_at: datetime() })
        WITH post, clientUser { .id, .username, .profile_pic_url } AS clientUserView
        RETURN post { .*, ownerUser: clientUserView, reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0, client_reaction: "", client_reposted: false, client_saved: false } AS new_post_data
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
            RETURN [notif IN collect(mentionNotif) | notif { .*, mentioned_user_id: mentionUser.id, mentioning_user: clientUserView }] AS mention_notifs
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

    await session.close()

    return res
  }

  static async repost(original_post_id, client_user_id) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (post:Post{ id: $original_post_id}), (clientUser:User{ id: $client_user_id })
      CREATE (clientUser)-[:CREATES_REPOST]->(repost:Repost:Post{ id: randomUUID(), type: post.type, media_urls: post.media_urls, description: post.description, created_at: datetime() })-[:REPOST_OF]->(post)
      WITH post, clientUser, repost, clientUser {.id, .username, .profile_pic_url} AS clientUserView
      MATCH (postOwner:User WHERE postOwner.id <> $client_user_id)-[:CREATES_POST]->(post)
      CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(repostNotif:Notification:RepostNotification{ id: randomUUID(), type: "repost", reposted_post_id: post.id, is_read: false, created_at: datetime() })-[:REPOSTER_USER]->(clientUser)
      RETURN repost { .*, owner_user: clientUserView, reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0, client_reaction: "", client_reposted: false, client_saved: false } AS repost_data,
        repostNotif { .*, reposted_post_owner_user_id: postOwner.id, reposter_user: clientUserView } AS repost_notif
      `,
      { original_post_id, client_user_id }
    )

    return records[0].toObject()
  }

  static async save(post_id, client_user_id) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (post:Post{ id: $post_id }), (clientUser:User{ id: $client_user_id })
      CREATE (clientUser)-[:SAVES_POST]->(post)
      WITH post
      MATCH (saver:User)-[:SAVES_POST]->(post)
      RETURN count(saver) + 1 AS latest_saves_count
      `,
      { post_id, client_user_id }
    )
    
    return records[0].toObject()
  }

  static async reactTo({
    client_user_id,
    post_id,
    reaction_code_point,
  }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (post:Post{ id: $post_id }), (clientUser:User{ id: $client_user_id })
      CREATE (clientUser)-[rt:REACTS_TO { reaction_code_point: $reaction_code_point }]->(post)
      WITH post, clientUser, rtr
      MATCH (reactor:User)-[:REACTS_TO]->(post)
      WITH post, clientUser, rtr, clientUser {.id, .username, .profile_pic_url} AS clientUserView
      MATCH (postOwner:User WHERE postOwner.id <> $client_user_id)-[:CREATES_POST]->(post)
      CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(reactNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction", reaction_code_point: rtr.reaction_code_point, is_read: false, created_at: datetime() })-[:REACTOR_USER]->(clientUser)
      RETURN count(reactor) + 1 AS latest_reactions_count,
        reactionNotif { .*, post_owner_user_id: postOwner.id, reactor_user: clientUserView } AS reaction_notif
      `,
      { client_user_id, post_id, reaction_code_point }
    )
      
    return records[0].toObject()
  }

  static async commentOn({
    post_id,
    post_owner_user_id,
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

      const { records: commentRecords } = await tx.run(
        `
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })
        CREATE (clientUser)-[:WRITES_COMMENT]->(comment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url, created_at: datetime() })-[:COMMENT_ON]->(post)
        WITH comment, clientUser { .id, .username, .profile_pic_url } AS clientUserView
        RETURN comment { .*, ownerUser: clientUserView, reactions_count: 0, comments_count: 0, client_reaction: "" } AS new_comment_data
        `,
        { client_username, attachment_url, comment_text, post_id }
      )

      new_comment_data = commentRecords[0].toObject().new_comment_data

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
          MATCH (mentionUser:User{ username: mentionUsername }), (comment:Comment{ id: $commentId })
          CREATE (post)-[:MENTIONS]->(mentionUser)
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
            WITH mentionUser, mentionNotif, clientUser { .id, .username, .profile_pic_url } AS clientUserView
            RETURN [notif IN collect(mentionNotif) | notif { .*, mentioned_user_id: mentionUser.id, mentioning_user: clientUserView }] AS mention_notifs
            `,
            { mentionsExcClient, commentId: new_comment_data.id, client_username }
          )

          mention_notifs = records[0].toObject().mention_notifs
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
      await tx.run(
        ``,
        
      )

      return { mention_notifs, new_comment_data }
    })

    await session.close()

    return res
  }

  static async find({ post_id, client_username, if_recommended }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_post($1, $2, $3)",
      values: [post_id, client_username, if_recommended],
    }

    return (await dbQuery(query)).rows[0]
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
    reaction_code_point,
    client_username,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_with_reaction_to_post($1, $2, $3, $4, $5)",
      values: [post_id, reaction_code_point, client_username, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async delete(post_id, user_id) {
    const query = {
      text: `DELETE FROM post WHERE id = $1 AND user_id = $2`,
      values: [post_id, user_id],
    }

    await dbQuery(query)
  }

  static async removeReaction(post_id, client_username) {
    const query = {
      text: `
      WITH pc_reaction AS (
        DELETE FROM pc_reaction WHERE post_id = $1 AND reactor_user_id = $2
      )
      SELECT reactions_count - 1 AS latest_reactions_count 
      FROM "PostView" 
      WHERE post_id = $1`,
      values: [post_id, client_username],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async removeComment({ post_id, comment_id, client_username }) {
    const query = {
      text: `
      WITH comment_cte AS (
        DELETE FROM comment_ WHERE id = $1 AND commenter_user_id = $2
      )
      SELECT comments_count - 1 AS latest_comments_count
      FROM "PostView" 
      WHERE post_id = $3
      `,
      values: [comment_id, client_username, post_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async unrepost(post_id, reposter_user_id) {
    const query = {
      text: `DELETE FROM repost WHERE post_id = $1 AND reposter_user_id = $2`,
      values: [post_id, reposter_user_id],
    }

    await dbQuery(query)
  }

  static async unsave(post_id, saver_user_id) {
    const query = {
      text: `
      WITH dsp AS (
        DELETE  FROM saved_post WHERE post_id = $1 AND saver_user_id = $2
      )
      SELECT saves_count - 1 AS latest_saves_count FROM "PostView" WHERE post_id = $1`,
      values: [post_id, saver_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }
}
