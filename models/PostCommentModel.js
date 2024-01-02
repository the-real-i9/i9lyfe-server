import { dbQuery } from "./db.js"

/**
 * @param {object} post
 * @param {number} post.user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const createNewPost = async (
  { user_id, media_urls, type, description },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Post" (user_id, media_urls, type, description) 
      VALUES ($1, $2, $3, $4) 
      RETURNING id, media_urls, type, description, reactions_count, comments_count, reposts_count user_id`,
    values: [user_id, media_urls, type, description],
  }

  const result = await dbClient.query(query)

  return result
}

/**
 * @param {string[]} usernames
 * @param {import("pg").PoolClient} dbClient
 * @returns {Promise<number[]>}
 */
export const mapUsernamesToUserIds = async (usernames, dbClient) => {
  return await Promise.all(
    usernames.map(async (username) => {
      console.log(username)
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
 * @param {"Post" | "Comment"} param0.post_or_comment_table
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {import("pg").PoolClient} dbClient
 */
export const incrementReactionsCount = async (
  { post_or_comment_table, post_or_comment_id },
  dbClient
) => {
  const query = {
    text: `UPDATE "${post_or_comment_table}" 
    SET reactions_count = reactions_count + 1 
    WHERE id = $1`,
    values: [post_or_comment_id],
  }

  await dbClient.query(query)
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
    VALUES ($1, $2, $3, $4) RETURNING id, commenter_user_id, comment_text, attachment_url, reactions_count, comments_count AS replies_count`,
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
 * @param {"Post" | "Comment"} param0.post_or_comment_table
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {import("pg").PoolClient} dbClient
 */
export const incrementCommentsCount = async (
  { post_or_comment_table, post_or_comment_id },
  dbClient
) => {
  const query = {
    text: `UPDATE "${post_or_comment_table}" 
    SET comments_count = comments_count + 1 
    WHERE id = $1`,
    values: [post_or_comment_id],
  }

  await dbClient.query(query)
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

// GET a post
/** 
 * @param {object} param0
 * @param {number} param0.post_id 
 * @param {number} param0.client_user_id
*/
export const getPost = async ({post_id, client_user_id}) => {
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
      reactions_count,
      comments_count, 
      reposts_count,
      reaction_code_point AS client_reaction,
      CASE
        WHEN "repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "reaction"
      ON "reaction".post_id = "post".id
      AND "reaction".reactor_user_id = $2
    LEFT JOIN "Repost" "repost"
      ON "repost".post_id = "post".id
      AND "repost".reposter_user_id = $2
    WHERE "post".id = $1`,
    values: [post_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

// GET all comments on a post
export const getAllPostComments = async ({post_id, client_user_id}) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "comment".id AS comment_id,
      comment_text,
      attachment_url,
      reactions_count,
      comments_count AS replies_count, 
      reaction_code_point AS client_reaction
    FROM "Comment" "comment"
    INNER JOIN "User" "user" ON "user".id = "comment".commenter_user_id
    LEFT JOIN "PostCommentReaction" "reaction"
      ON "reaction".comment_id = "comment".id
      AND "reaction".reactor_user_id = $2
    WHERE "comment".post_id = $1`,
    values: [post_id, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET a single comment on a post
export const getComment = async (post_id, comment_id) => {}

// GET all reactions to a post: returning all users that reacted to the post
export const getAllPostReactors = async (post_id) => {}

// GET a single reaction to a post: limiting returned users to the ones with that reaction
export const getPostReactorsForReaction = ({
  post_id,
  reaction_code_point,
}) => {}

// GET all replies to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
export const getAllCommentReplies = (post_id, comment_id) => {}

// GET a single reply to a comment/reply
// the :comment_id either selects a comment or reply, since all replies are comments
// the :reply_id is a single reply to the comment/reply with the that id
export const getCommentReply = ({ post_id, comment_id, reply_id }) => {}

// GET all reactions to a comment/reply: returning all users that reacted to the comment
// the :comment_id either selects a comment or reply, since all replies are comments
export const getAllCommentReactors = (post_id, comment_id) => {}

// GET a specific reaction to a comment/reply: limiting returned users to the ones with that reaction
// the :comment_id either selects a comment or reply, since all replies are comments
export const getCommentReactorsForReaction = async ({
  post_id,
  comment_id,
  reaction_code_point,
}) => {}
