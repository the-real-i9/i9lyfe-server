import { dbQuery } from "../configs/db.js"

/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

export class Post {
  /**
   * @param {object} post
   * @param {number} post.client_user_id
   * @param {string[]} post.media_urls
   * @param {string[]} post.mentions
   * @param {string[]} post.hashtags
   * @param {"photo" | "video" | "reel" | "story"} post.type
   * @param {string} post.description
   */
  static async create({
    client_user_id,
    media_urls,
    type,
    description,
    mentions,
    hashtags,
  }) {
    const query = {
      text: "SELECT new_post_id, mention_notifs FROM create_post($1, $2, $3, $4, $5, $6)",
      values: [
        client_user_id,
        [...media_urls],
        type,
        description,
        [...mentions],
        [...hashtags],
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async repost(original_post_id, client_user_id) {
    const query = {
      text: `
      INSERT INTO repost (post_id, reposter_user_id) 
      VALUES ($1, $2)`,
      values: [original_post_id, client_user_id],
    }

    await dbQuery(query)
  }

  static async save(post_id, client_user_id) {
    const query = {
      text: `
      WITH new_sp AS (
        INSERT INTO saved_post (saver_user_id, post_id) 
        VALUES ($1, $2)
      )
      SELECT saves_count + 1 AS latest_saves_count FROM "PostView" WHERE post_id = $2`,
      values: [client_user_id, post_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async reactTo({
    client_user_id,
    target_post_id,
    target_post_owner_user_id,
    reaction_code_point,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `
      SELECT reaction_notif, latest_reactions_count 
      FROM create_reaction_to_post($1, $2, $3, $4)`,
      values: [
        client_user_id,
        target_post_id,
        target_post_owner_user_id,
        reaction_code_point,
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async commentOn({
    target_post_id,
    target_post_owner_user_id,
    client_user_id,
    comment_text,
    attachment_url,
    mentions,
    hashtags,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `
      SELECT new_comment_id, 
        comment_notif, 
        mention_notifs, 
        latest_comments_count 
      FROM create_comment_on_post($1, $2, $3, $4, $5, $6, $7)`,
      values: [
        target_post_id,
        target_post_owner_user_id,
        client_user_id,
        comment_text,
        attachment_url,
        mentions,
        hashtags,
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async find({ post_id, client_user_id, if_recommended }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_post($1, $2, $3)",
      values: [post_id, client_user_id, if_recommended],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async getComments({ post_id, client_user_id, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_comments_on_post($1, $2, $3, $4)",
      values: [post_id, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactors({ post_id, client_user_id, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_to_post($1, $2, $3, $4)",
      values: [post_id, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactorsWithReaction({
    post_id,
    reaction_code_point,
    client_user_id,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_with_reaction_to_post($1, $2, $3, $4, $5)",
      values: [post_id, reaction_code_point, client_user_id, limit, offset],
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

  static async removeReaction(target_post_id, client_user_id) {
    const query = {
      text: `
      WITH pc_reaction AS (
        DELETE FROM pc_reaction WHERE target_post_id = $1 AND reactor_user_id = $2
      )
      SELECT reactions_count - 1 AS latest_reactions_count 
      FROM "PostView" 
      WHERE post_id = $1`,
      values: [target_post_id, client_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async removeComment({ post_id, comment_id, client_user_id }) {
    const query = {
      text: `
      WITH comment_cte AS (
        DELETE FROM comment_ WHERE id = $1 AND commenter_user_id = $2
      )
      SELECT comments_count - 1 AS latest_comments_count
      FROM "PostView" 
      WHERE post_id = $3
      `,
      values: [comment_id, client_user_id, post_id],
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
