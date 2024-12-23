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
        RETURN { id: post.id, ownerUser: { id: clientUser.id, username: clientUser.username, profile_pic_url: clientUser.profile_pic_url }, type: $type, media_urls: $media_urls, description: $description, reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0, client_reaction: "", client_reposted: false, client_saved: false } AS new_post_data
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
            CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_post", mentioned_user_id: mentionUser.id, in_post_id: post.id })-[:MENTIONING_USER]->(clientUser)
            RETURN [notif IN collect(mentionNotif) | { id: notif.id, type: notif.type, mentioned_user_id: notif.mentioned_user_id, in_post_id: notif.in_post_id, mentioning_user: { id: clientUser.id, username: clientUser.username, profile_pic_url: clientUser.profile_pic_url } }] AS mention_notifs
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

  static async repost(original_post_id, client_username) {
    const query = {
      text: `
      INSERT INTO repost (post_id, reposter_user_id) 
      VALUES ($1, $2)`,
      values: [original_post_id, client_username],
    }

    await dbQuery(query)
  }

  static async save(post_id, client_username) {
    const query = {
      text: `
      WITH new_sp AS (
        INSERT INTO saved_post (saver_user_id, post_id) 
        VALUES ($1, $2)
      )
      SELECT saves_count + 1 AS latest_saves_count FROM "PostView" WHERE post_id = $2`,
      values: [client_username, post_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async reactTo({
    client_username,
    post_id,
    post_owner_user_id,
    reaction_code_point,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `
      SELECT reaction_notif, latest_reactions_count 
      FROM create_reaction_to_post($1, $2, $3, $4)`,
      values: [
        client_username,
        post_id,
        post_owner_user_id,
        reaction_code_point,
      ],
    }

    return (await dbQuery(query)).rows[0]
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
    /** @type {PgQueryConfig} */
    const query = {
      text: `
      SELECT new_comment_data, 
        comment_notif, 
        mention_notifs, 
        latest_comments_count 
      FROM create_comment_on_post($1, $2, $3, $4, $5, $6, $7)`,
      values: [
        post_id,
        post_owner_user_id,
        client_username,
        comment_text,
        attachment_url,
        mentions,
        hashtags,
      ],
    }

    return (await dbQuery(query)).rows[0]
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
