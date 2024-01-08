import { dbQuery } from "./db.js"

/**
 * @param {object} post
 * @param {number} post.client_user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const createNewPost = async (
  { client_user_id, media_urls, type, description },
  dbClient
) => {
  const query = {
    text: `
    INSERT INTO "Post" (user_id, media_urls, type, description) 
    VALUES ($1, $2, $3, $4) 
    RETURNING id, user_id, media_urls, type, description`,
    values: [client_user_id, media_urls, type, description],
  }

  const result = await dbClient.query(query)

  return result
}

export const createRepost = async (reposted_post_id, reposter_user_id) => {
  const query = {
    text: `
    INSERT INTO "Repost" (post_id, reposter_user_id) 
    VALUES ($1, $2)`,
    values: [reposted_post_id, reposter_user_id],
  }

  await dbQuery(query)
}

export const savePost = async (post_id, client_user_id) => {
  const query = {
    text: `
    INSERT INTO "SavedPost" (saver_user_id, post_id) 
    VALUES ($1, $2)
    `,
    values: [client_user_id, post_id],
  }

  await dbQuery(query)
}

/**
 * @param {string[]} usernames
 * @param {import("pg").PoolClient} dbClient
 * @returns {Promise<number[]>}
 */
export const mapUsernamesToUserIds = async (usernames, dbClient) => {
  return await Promise.all(
    usernames.map(async (username) => {
      const query = {
        text: 'SELECT id FROM "User" WHERE username = $1',
        values: [username],
      }
      return (await dbClient.query(query)).rows[0].id
    })
  )
}

/**
 * @param {object} param0
 * @param {number} param0.post_or_comment_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {import("pg").PoolClient} dbClient
 */
export const createMentions = async (
  { post_or_comment, post_or_comment_id, mentioned_user_ids },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostCommentMention" (${post_or_comment}_id, user_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 2)}`,
    values: mentioned_user_ids
      .map((mentioned_user_id) => [post_or_comment_id, mentioned_user_id])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.sender_user_id
 * @param {number[]} param0.receiver_user_ids
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {import("pg").PoolClient} dbClient
 */
export const createMentionsNotifications = async (
  { sender_user_id, receiver_user_ids, post_or_comment, post_or_comment_id },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostCommentNotification" (type, sender_user_id, receiver_user_id, ${post_or_comment}_id) 
    VALUES ${multipleRowsParameters(receiver_user_ids.length, 4)}`,
    values: receiver_user_ids
      .map((receiver_user_id) => [
        "mention",
        sender_user_id,
        receiver_user_id,
        post_or_comment_id,
      ])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {string[]} param0.hashtag_names
 * @param {import("pg").PoolClient} dbClient
 */
export const createHashtags = async (
  { post_or_comment, post_or_comment_id, hashtag_names },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "PostCommentHashtag" (${post_or_comment}_id, hashtag_name) 
    VALUES ${multipleRowsParameters(hashtag_names.length, 2)}`,
    values: hashtag_names
      .map((hashtag_name) => [post_or_comment_id, hashtag_name])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {number} rowsCount
 * @param {number} fieldsCountPerRow
 */
const multipleRowsParameters = (rowsCount, fieldsCountPerRow) =>
  Array(rowsCount)
    .fill()
    .map(
      (r, ri) =>
        `(${Array(fieldsCountPerRow)
          .fill()
          .map((f, fi) => `$${ri * fieldsCountPerRow + (fi + 1)}`)
          .join(", ")})`
    )
    .join(", ")

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.post_or_comment_user_id
 * @param {"post" | "comment"} param0.post_or_comment Post `id` or Comment `id`
 * @param {number} param0.post_or_comment_id
 * @param {number} param0.reaction_code_point
 * @param {import("pg").PoolClient} dbClient
 */
export const createReaction = async (
  { reactor_user_id, post_or_comment, post_or_comment_id, reaction_code_point },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostCommentReaction" (reactor_user_id, ${post_or_comment}_id, reaction_code_point) 
      VALUES ($1, $2, $3) RETURNING id`,
    values: [reactor_user_id, post_or_comment_id, reaction_code_point],
  }

  const result = await dbClient.query(query)

  return result
}

/**
 * @param {object} param0
 * @param {number} param0.sender_user_id
 * @param {number} param0.receiver_user_id
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {number} param0.post_or_comment_user_id
 * @param {number} param0.reaction_id
 * @param {import("pg").PoolClient} dbClient
 */
export const createReactionNotification = async (
  {
    sender_user_id,
    receiver_user_id,
    post_or_comment,
    post_or_comment_id,
    reaction_id,
  },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostCommentNotification" (type, sender_user_id, receiver_user_id, ${post_or_comment}_id, type_created_id)
    VALUES ($1, $2, $3, $4, $5)`,
    values: [
      "reaction",
      sender_user_id,
      receiver_user_id,
      post_or_comment_id,
      reaction_id,
    ],
  }

  await dbClient.query(query)
}

/**
 *
 * @param {object} param0
 * @param {number} param0.commenter_user_id
 * @param {string} param0.comment_text
 * @param {string} param0.attachment_url
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {import("pg").PoolClient} dbClient
 */
export const createComment = async (
  {
    commenter_user_id,
    comment_text,
    attachment_url,
    post_or_comment,
    post_or_comment_id,
  },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "Comment" (commenter_user_id, comment_text, attachment_url, ${post_or_comment}_id)
    VALUES ($1, $2, $3, $4) RETURNING id, commenter_user_id${
      post_or_comment === "comment" ? " AS replier_user_id" : null
    }, comment_text${
      post_or_comment === "comment" ? " AS reply_text" : null
    }, attachment_url`,
    values: [
      commenter_user_id,
      comment_text,
      attachment_url,
      post_or_comment_id,
    ],
  }

  const result = await dbClient.query(query)
  return result
}

/**
 * @param {object} param0
 * @param {number} param0.sender_user_id
 * @param {number} param0.receiver_user_id
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {number} param0.new_comment_id
 * @param {import("pg").PoolClient} dbClient
 */
export const createCommentNotification = async (
  {
    sender_user_id,
    receiver_user_id,
    post_or_comment,
    post_or_comment_id,
    new_comment_id,
  },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostCommentNotification" (type, sender_user_id, receiver_user_id, ${post_or_comment}_id, type_created_id)
    VALUES ($1, $2, $3, $4, $5)`,
    values: [
      "comment",
      sender_user_id,
      receiver_user_id,
      post_or_comment_id,
      new_comment_id,
    ],
  }

  await dbClient.query(query)
}

/* ************* */

/**
 * @param {object} param0
 * @param {number} param0.post_id
 * @param {number} param0.client_user_id
 */
export const getPost = async (post_id, client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "post".id AS post_id,
      type,
      media_urls,
      description,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "any_comment".id)::INTEGER AS comments_count, 
      COUNT(DISTINCT "any_repost".id)::INTEGER AS reposts_count,
      COUNT(DISTINCT "any_saved_post".id)::INTEGER AS saves_count,
      "client_reaction".reaction_code_point AS client_reaction,
      CASE
        WHEN "client_repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted,
      CASE
        WHEN "client_saved_post".id IS NULL THEN false
        ELSE true
      END AS client_saved
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".post_id = "post".id 
    LEFT JOIN "Comment" "any_comment" ON "any_comment".post_id = "post".id
    LEFT JOIN "Repost" "any_repost" ON "any_repost".post_id = "post".id
    LEFT JOIN "SavedPost" "any_saved_post" ON "any_saved_post".post_id = "post".id
    LEFT JOIN "PostCommentReaction" "client_reaction" 
      ON "client_reaction".post_id = "post".id AND "client_reaction".reactor_user_id = $2
    LEFT JOIN "Repost" "client_repost" 
      ON "client_repost".post_id = "post".id AND "client_repost".reposter_user_id = $2
    LEFT JOIN "SavedPost" "client_saved_post" 
      ON "client_saved_post".post_id = "post".id AND "client_saved_post".saver_user_id = $2
    WHERE "post".id = $1
    GROUP BY owner_user_id, 
      owner_username, 
      owner_profile_pic_url, 
      "post".id, 
      type, 
      media_urls, 
      description, 
      client_reaction, 
      client_reposted,
      client_saved`,
    values: [post_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

export const getAllCommentsOnPost_OR_RepliesToComment = async ({
  post_or_comment,
  post_or_comment_id,
  client_user_id,
}) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "comment".id AS ${post_or_comment === "post" ? "comment" : "reply"}_id,
      "comment".comment_text AS ${
        post_or_comment === "post" ? "comment" : "reply"
      }_text,
      "comment".attachment_url AS attachment_url,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "reply".id)::INTEGER AS replies_count, 
      "client_reaction".reaction_code_point AS client_reaction
    FROM "Comment" "comment"
    INNER JOIN "User" "user" ON "user".id = "comment".commenter_user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".comment_id = "comment".id
    LEFT JOIN "Comment" "reply" ON "reply".comment_id = "comment".id
    LEFT JOIN "PostCommentReaction" "client_reaction"
      ON "client_reaction".comment_id = "comment".id AND "client_reaction".reactor_user_id = $2
    WHERE "comment".${post_or_comment}_id = $1
    GROUP BY owner_user_id,
      owner_username,
      owner_profile_pic_url,
      "comment".id,
      "comment".comment_text,
      "comment".attachment_url,
      client_reaction`,
    values: [post_or_comment_id, client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const getCommentOnPost_OR_ReplyToComment = async ({
  post_or_comment,
  comment_or_reply_id,
  client_user_id,
}) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
  SELECT "user".id AS owner_user_id,
    "user".username AS owner_username,
    "user".profile_pic_url AS owner_profile_pic_url,
    "comment".id AS ${post_or_comment === "post" ? "comment" : "reply"}_id,
    "comment".comment_text AS ${
      post_or_comment === "post" ? "comment" : "reply"
    }_text,
    "comment".attachment_url AS attachment_url,
    COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
    COUNT(DISTINCT "reply".id)::INTEGER AS replies_count, 
    "client_reaction".reaction_code_point AS client_reaction
  FROM "Comment" "comment"
  INNER JOIN "User" "user" ON "user".id = "comment".commenter_user_id
  LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".comment_id = "comment".id
  LEFT JOIN "Comment" "reply" ON "reply".comment_id = "comment".id
  LEFT JOIN "PostCommentReaction" "client_reaction"
    ON "client_reaction".comment_id = "comment".id AND "client_reaction".reactor_user_id = $2
  WHERE "comment".id = $1
  GROUP BY owner_user_id,
    owner_username,
    owner_profile_pic_url,
    "comment".id,
    "comment".comment_text,
    "comment".attachment_url,
    client_reaction`,
    values: [comment_or_reply_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

export const getAllReactorsToPost_OR_Comment = async ({
  post_or_comment,
  post_or_comment_id,
  client_user_id,
}) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id, 
      "user".profile_pic_url, 
      "user".username, 
      "user".name,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "User" "user" 
    INNER JOIN "PostCommentReaction" "reaction" 
      ON "reaction".reactor_user_id = "user".id 
      AND "reaction".${post_or_comment}_id = $1
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "user".id AND "client_follows".follower_user_id = $2`,
    values: [post_or_comment_id, client_user_id],
  }

  return (await dbQuery(query)).rows
}


export const getAllReactorsWithReactionToPost_OR_Comment = async ({
  post_or_comment,
  post_or_comment_id,
  reaction_code_point,
  client_user_id,
}) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id, 
      "user".profile_pic_url, 
      "user".username, 
      "user".name,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "User" "user" 
    INNER JOIN "PostCommentReaction" "reaction" 
      ON "reaction".reactor_user_id = "user".id 
      AND "reaction".${post_or_comment}_id = $1
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "user".id AND "client_follows".follower_user_id = $3
    WHERE "reaction".reaction_code_point = $2`,
    values: [post_or_comment_id, reaction_code_point, client_user_id],
  }

  return (await dbQuery(query)).rows
}

/** DELETs */
export const deletePost = async (post_id, user_id) => {
  const query = {
    text: `DELETE FROM "Post" WHERE id = $1 AND user_id = $2`,
    values: [post_id, user_id],
  }

  await dbQuery(query)
}

/** 
 * @param {object} param0
 * @param {"post" | "comment"} post_or_comment
 * @param {number} post_or_comment_id
 * @param {number} reactor_user_id
 */
export const removeReactionToPost_OR_Comment = async ({
  post_or_comment,
  post_or_comment_id,
  reactor_user_id,
}) => {
  const query = {
    text: `DELETE FROM "PostCommentReaction" WHERE ${post_or_comment}_id = $1 AND reactor_user_id = $2`,
    values: [post_or_comment_id, reactor_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} comment_or_reply_id 
 * @param {number} commenter_or_replier_user_id 
 */
export const deleteCommentOnPost_OR_ReplyToComment = async (
  comment_or_reply_id,
  commenter_or_replier_user_id
) => {
  const query = {
    text: `DELETE FROM "Comment" WHERE id = $1 AND commenter_user_id = $2`,
    values: [comment_or_reply_id, commenter_or_replier_user_id],
  }

  await dbQuery(query)
}

export const deleteRepost = async (reposted_post_id, reposter_user_id) => {
  const query = {
    text: `DELETE FROM "Repost" WHERE post_id = $1 AND reposter_user_id = $2`,
    values: [reposted_post_id, reposter_user_id],
  }

  await dbQuery(query)
}

export const unsavePost = async (post_id, saver_user_id) => {
  const query = {
    text: `DELETE  FROM "SavedPost" WHERE post_id = $1 AND saver_user_id = $2`,
    values: [post_id, saver_user_id],
  }

  await dbQuery(query)
}
