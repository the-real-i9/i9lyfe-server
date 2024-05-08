import { generateMultiRowInsertValuesParameters } from "../utils/helpers.js"
import { dbQuery } from "./db.js"

/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

/**
 * @param {object} post
 * @param {number} post.client_user_id
 * @param {string[]} post.media_urls
 * @param {string[]} post.mentions
 * @param {string[]} post.hashtags
 * @param {"photo" | "video" | "reel" | "story"} post.type
 * @param {string} post.description
 */
export const createNewPost = async ({
  client_user_id,
  media_urls,
  type,
  description,
  mentions,
  hashtags,
}) => {
  const query = {
    text: "SELECT new_post_id, mention_notifs FROM create_post($1, $2, $3, $4, $5, $6)",
    values: [client_user_id, [...media_urls], type, description, [...mentions], [...hashtags]],
  }

  return (await dbQuery(query)).rows[0]
}

export const createRepost = async (original_post_id, client_user_id) => {
  const query = {
    text: `
    INSERT INTO repost (post_id, reposter_user_id) 
    VALUES ($1, $2)`,
    values: [original_post_id, client_user_id],
  }

  await dbQuery(query)
}

export const savePost = async (post_id, client_user_id) => {
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

export const createReactionToPost = async ({
  client_user_id,
  target_post_id,
  target_post_owner_user_id,
  reaction_code_point,
}) => {
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

export const createReactionToComment = async ({
  client_user_id,
  target_comment_id,
  target_comment_owner_user_id,
  reaction_code_point,
}) => {
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

export const createCommentOnPost = async ({
  target_post_id,
  target_post_owner_user_id,
  client_user_id,
  comment_text,
  attachment_url,
  mentions,
  hashtags,
}) => {
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

export const createCommentOnComment = async ({
  target_comment_id,
  target_comment_owner_user_id,
  client_user_id,
  comment_text,
  attachment_url,
  mentions,
  hashtags,
}) => {
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

/* ************* */

/**
 * @param {object} param0
 * @param {number} param0.post_id
 * @param {number} param0.client_user_id
 */
export const getPost = async (post_id, client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_post($1, $2)",
    values: [post_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

export const getCommentsOnPost = async ({
  post_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_comments_on_post($1, $2, $3, $4)",
    values: [post_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getCommentsOnComment = async ({
  comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_comments_on_comment($1, $2, $3, $4)",
    values: [comment_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getComment = async (comment_id, client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_comment($1, $2)",
    values: [comment_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

export const getReactorsToPost = async ({
  post_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_reactors_to_post($1, $2, $3, $4)",
    values: [post_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getReactorsToComment = async ({
  comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_reactors_to_comment($1, $2, $3, $4)",
    values: [comment_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getReactorsWithReactionToPost = async ({
  post_id,
  reaction_code_point,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_reactors_with_reaction_to_post($1, $2, $3, $4, $5)",
    values: [post_id, reaction_code_point, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getReactorsWithReactionToComment = async ({
  comment_id,
  reaction_code_point,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_reactors_with_reaction_to_comment($1, $2, $3, $4, $5)",
    values: [comment_id, reaction_code_point, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

/** DELETs */
export const deletePost = async (post_id, user_id) => {
  const query = {
    text: `DELETE FROM post WHERE id = $1 AND user_id = $2`,
    values: [post_id, user_id],
  }

  await dbQuery(query)
}

export const removeReactionToPost = async (target_post_id, client_user_id) => {
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

export const removeReactionToComment = async (target_comment_id, client_user_id) => {
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


export const deleteCommentOnPost = async ({
  post_id,
  comment_id,
  client_user_id,
}) => {
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

export const deleteCommentOnComment = async ({
  parent_comment_id,
  comment_id,
  client_user_id,
}) => {
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

export const deleteRepost = async (post_id, reposter_user_id) => {
  const query = {
    text: `DELETE FROM repost WHERE post_id = $1 AND reposter_user_id = $2`,
    values: [post_id, reposter_user_id],
  }

  await dbQuery(query)
}

export const unsavePost = async (post_id, saver_user_id) => {
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
