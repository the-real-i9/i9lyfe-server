import { dbQuery } from "./db.js"

/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

export class Comment {
  static async reactTo({
    client_user_id,
    target_comment_id,
    target_comment_owner_user_id,
    reaction_code_point,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `
    SELECT reaction_notif, latest_reactions_count 
    FROM create_reaction_to_comment($1, $2, $3, $4)`,
      values: [
        client_user_id,
        target_comment_id,
        target_comment_owner_user_id,
        reaction_code_point,
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async commentOn({
    target_comment_id,
    target_comment_owner_user_id,
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
    FROM create_comment_on_comment($1, $2, $3, $4, $5, $6, $7)`,
      values: [
        target_comment_id,
        target_comment_owner_user_id,
        client_user_id,
        comment_text,
        attachment_url,
        mentions,
        hashtags,
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async getComments({
    comment_id,
    client_user_id,
    limit,
    offset,
  }) {
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

  static async getReactors({
    comment_id,
    client_user_id,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_to_comment($1, $2, $3, $4)",
      values: [comment_id, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getReactorsWithReaction({
    comment_id,
    reaction_code_point,
    client_user_id,
    limit,
    offset,
  }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reactors_with_reaction_to_comment($1, $2, $3, $4, $5)",
      values: [comment_id, reaction_code_point, client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async removeReaction(target_comment_id, client_user_id) {
    const query = {
      text: `
    WITH pc_reaction AS (
      DELETE FROM pc_reaction WHERE target_comment_id = $1 AND reactor_user_id = $2
    )
    SELECT reactions_count - 1 AS latest_reactions_count 
    FROM "CommentView" 
    WHERE comment_id = $1`,
      values: [target_comment_id, client_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async removeChildComment({
    parent_comment_id,
    comment_id,
    client_user_id,
  }) {
    const query = {
      text: `
    WITH comment_cte AS (
      DELETE FROM comment_ WHERE id = $1 AND commenter_user_id = $2
    )
    SELECT comments_count - 1 AS latest_comments_count
    FROM "CommentView" 
    WHERE comment_id = $3
    `,
      values: [comment_id, client_user_id, parent_comment_id],
    }

    return (await dbQuery(query)).rows[0]
  }
}
