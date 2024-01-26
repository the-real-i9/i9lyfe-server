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
 * @param {string} post.type
 * @param {string} post.description
 * @param {PgPoolClient} dbClient
 */
export const createNewPost = async (
  { client_user_id, media_urls, type, description },
  dbClient
) => {
  const query = {
    text: `
    WITH post AS (
      INSERT INTO "Post" (user_id, media_urls, type, description) 
      VALUES ($1, $2, $3, $4) 
      RETURNING id, user_id, media_urls, type, description
    )
    SELECT post.id AS post_id,
      media_urls,
      type,
      description,
      json_build_object(
        'user_id', "user".id,
        'username', "user".username,
        'name', "user".name,
        'profile_pic_url', "user".profile_pic_url,
        'connection_status', "user".connection_status
      ) AS owner_user
    FROM post
    INNER JOIN "User" "user" ON "user".id = post.user_id`,
    values: [client_user_id, media_urls, type, description],
  }

  return (await dbClient.query(query)).rows[0]
}

export const createRepost = async (original_post_id, reposter_user_id) => {
  const query = {
    text: `
    INSERT INTO "Repost" (post_id, reposter_user_id) 
    VALUES ($1, $2)`,
    values: [original_post_id, reposter_user_id],
  }

  await dbQuery(query)
}

export const savePost = async (post_id, client_user_id) => {
  const query = {
    text: `
    WITH iisp AS (
      INSERT INTO "SavedPost" (saver_user_id, post_id) 
      VALUES ($1, $2)
    )
    SELECT saves_count FROM "AllPostsView" WHERE post_id = $2`,
    values: [client_user_id, post_id],
  }

  return (await dbQuery(query)).rows[0].saves_count
}

/**
 * @param {string[]} usernames
 * @param {PgPoolClient} dbClient
 * @returns {Promise<number[]>}
 */
export const mapUsernamesToUserIds = async (usernames, dbClient) => {
  const query = {
    text: 'SELECT id, username FROM "User" WHERE username = ANY($1)',
    values: [[...usernames]],
  }

  const usernameToIdDict = (await dbClient.query(query)).rows.reduce(
    (acc, { id, username }) => {
      acc[username] = id
      return acc
    },
    {}
  )

  return usernames.map((username) => usernameToIdDict[username])
}

/**
 * @param {object} param0
 * @param {number} param0.post_or_comment_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {number} param0.content_owner_user_id
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {PgPoolClient} dbClient
 */
export const createMentions = async (
  {
    post_or_comment,
    post_or_comment_id,
    mentioned_user_ids,
    content_owner_user_id,
  },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH pc_mention AS (
      INSERT INTO "PostCommentMention" (${post_or_comment}_id, user_id) 
      VALUES ${generateMultiRowInsertValuesParameters({
        rowsCount: mentioned_user_ids.length,
        columnsCount: 2,
      })}
    ), mention_notification AS (
      INSERT INTO "Notification" (type, sender_user_id, receiver_user_id, ${post_or_comment}_id) 
      VALUES ${generateMultiRowInsertValuesParameters({
        rowsCount: mentioned_user_ids.length,
        columnsCount: 4,
        paramNumFrom: mentioned_user_ids.length * 2 + 1,
      })} 
      RETURNING type, sender_user_id, receiver_user_id, ${post_or_comment}_id
    )
    SELECT sender.id AS sender_user_id,
      mention_notification.receiver_user_id,
      sender.username AS sender_username,
      sender.profile_pic_url AS sender_profile_pic_url,
      mention_notification.type,
      mention_notification.${post_or_comment}_id
    FROM mention_notification
    INNER JOIN "User" sender ON sender.id = mention_notification.sender_user_id`,
    values: [
      ...mentioned_user_ids.map((mentioned_user_id) => [
        post_or_comment_id,
        mentioned_user_id,
      ]),
      ...mentioned_user_ids.map((receiver_user_id) => [
        "mention",
        content_owner_user_id,
        receiver_user_id,
        post_or_comment_id,
      ]),
    ].flat(),
  }

  return (await dbClient.query(query)).rows
}

/**
 * @param {object} param0
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {string[]} param0.hashtag_names
 * @param {PgPoolClient} dbClient
 */
export const createHashtags = async (
  { post_or_comment, post_or_comment_id, hashtag_names },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "PostCommentHashtag" (${post_or_comment}_id, hashtag_name) 
    VALUES ${generateMultiRowInsertValuesParameters({
      rowsCount: hashtag_names.length,
      columnsCount: 2,
    })}`,
    values: hashtag_names
      .map((hashtag_name) => [post_or_comment_id, hashtag_name])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.reactor_user_id
 * @param {number} param0.content_owner_user_id
 * @param {"post" | "comment"} param0.post_or_comment Post `id` or Comment `id`
 * @param {number} param0.post_or_comment_id
 * @param {number} param0.reaction_code_point
 */
export const createReaction = async ({
  reactor_user_id,
  content_owner_user_id,
  post_or_comment,
  post_or_comment_id,
  reaction_code_point,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH pc_reaction AS (
      INSERT INTO "PostCommentReaction" (reactor_user_id, ${post_or_comment}_id, reaction_code_point) 
      VALUES ($1, $2, $3) 
      RETURNING id AS new_reaction_id
    ), reaction_notification AS (
      INSERT INTO "Notification" (sender_user_id, ${post_or_comment}_id, type, receiver_user_id, reaction_created_id)
      VALUES ($1, $2, $4, $5, (SELECT new_reaction_id FROM pc_reaction)) 
      RETURNING type, sender_user_id, receiver_user_id, ${post_or_comment}_id
    )
    SELECT json_build_object(
      'notifData', (SELECT json_build_object(
          'sender_user_id', sender.id,
          'sender_username', sender.username,
          'sender_profile_pic_url', sender.profile_pic_url,
          'reciver_user_id', reaction_notification.receiver_user_id,
          'type', reaction_notification.type,
          '.${post_or_comment}_id', reaction_notification.${post_or_comment}_id
        )
        FROM reaction_notification
        INNER JOIN "User" sender ON sender.id = reaction_notification.sender_user_id),
      'latestReactionsCount',  ${
        post_or_comment === "post"
          ? `(SELECT reactions_count 
              FROM "AllPostsView" 
              WHERE post_id = $2)`
          : `(SELECT reactions_count 
              FROM "CommentsOnPost_RepliesToCommentView" 
              WHERE main_comment_id = $2)`
      } 
    ) AS data
    `,
    values: [
      reactor_user_id,
      post_or_comment_id,
      reaction_code_point,
      "reaction",
      content_owner_user_id,
    ],
  }

  return (await dbQuery(query)).rows[0].data
}

/**
 *
 * @param {object} param0
 * @param {number} param0.commenter_user_id
 * @param {number} param0.content_owner_user_id
 * @param {string} param0.comment_text
 * @param {string} param0.attachment_url
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {PgPoolClient} dbClient
 */
export const createComment = async (
  {
    commenter_user_id,
    content_owner_user_id,
    comment_text,
    attachment_url,
    post_or_comment,
    post_or_comment_id,
  },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH comment_cte AS (
      INSERT INTO "Comment" (commenter_user_id, comment_text, attachment_url, ${post_or_comment}_id)
      VALUES ($1, $2, $3, $4) 
      RETURNING id AS new_comment_id, commenter_user_id${
        post_or_comment === "comment" ? " AS replier_user_id" : ""
      }, comment_text${
      post_or_comment === "comment" ? " AS reply_text" : ""
    }, attachment_url
    ), comment_notification AS (
      INSERT INTO "Notification" (sender_user_id, ${post_or_comment}_id, type, receiver_user_id, comment_created_id)
      VALUES ($1, $4, $5, $6, (SELECT new_comment_id FROM comment_cte)) 
      RETURNING type, sender_user_id, receiver_user_id, ${post_or_comment}_id
    )
    SELECT json_build_object(
      'commentData', (SELECT json_build_object(
          'id', new_comment_id,
          ${
            post_or_comment === "comment"
              ? `'replier_user_id', replier_user_id, 'reply_text', reply_text,`
              : `'commenter_user_id', commenter_user_id, 'comment_text', comment_text,`
          }
          'attachment_url', attachment_url)
        FROM comment_cte),
      'notifData', (SELECT json_build_object(
          'sender_user_id', sender.id,
          'receiver_user_id', comment_notification.receiver_user_id,
          'sender_username', sender.username,
          'sender_profile_pic_url', sender.profile_pic_url,
          'type', comment_notification.type,
          '${post_or_comment}_id', comment_notification.${post_or_comment}_id)
        FROM comment_notification
        INNER JOIN "User" sender ON sender.id = comment_notification.sender_user_id),
      'latestCommentsRepliesCount', ${
        post_or_comment === "post"
          ? `(SELECT comments_count 
              FROM "AllPostsView" 
              WHERE post_id = $4)`
          : `(SELECT replies_count 
              FROM "CommentsOnPost_RepliesToCommentView" 
              WHERE main_comment_id = $4)`
      }
    ) AS data`,
    values: [
      commenter_user_id,
      comment_text,
      attachment_url,
      post_or_comment_id,
      "comment",
      content_owner_user_id,
    ],
  }

  return (await dbClient.query(query)).rows[0].data
}

/* ************* */

export const getFeedPosts = async ({ client_user_id, limit, offset }) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $1 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $1 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $1 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    INNER JOIN "Follow" follow ON follow.followee_user_id = owner_user_id
    WHERE follow.follower_user_id = $1
    ORDER BY created_at DESC
    LIMIT $2 OFFSET $3`,
    values: [client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

/**
 * @param {object} param0
 * @param {number} param0.post_id
 * @param {number} param0.client_user_id
 */
export const getPost = async (post_id, client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $2 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $2 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $2 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    WHERE post_id = $1
    `,
    values: [post_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

/**
 * @param {object} param0
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {number} param0.client_user_id
 * @param {number} param0.limit
 * @param {number} param0.offset
 * @returns
 */
export const getAllCommentsOnPost_OR_RepliesToComment = async ({
  post_or_comment,
  post_or_comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      main_comment_id AS ${post_or_comment === "post" ? "comment" : "reply"}_id,
      comment_text AS ${post_or_comment === "post" ? "comment" : "reply"}_text,
      attachment_url,
      reactions_count,
      replies_count,
      CASE 
        WHEN reactor_user_id = $2 THEN reaction_code_point 
        ELSE NULL
      END AS client_reaction
    FROM "CommentsOnPost_RepliesToCommentView"
    WHERE owner_${post_or_comment}_id = $1 
    ORDER BY created_at DESC
    LIMIT $3 OFFSET $4
    `,
    values: [post_or_comment_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getCommentOnPost_OR_ReplyToComment = async ({
  post_or_comment,
  comment_or_reply_id,
  client_user_id,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
      'user_id', owner_user_id,
      'username', owner_username,
      'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      main_comment_id AS ${post_or_comment === "post" ? "comment" : "reply"}_id,
      comment_text AS ${post_or_comment === "post" ? "comment" : "reply"}_text,
      attachment_url,
      reactions_count,
      replies_count,
      CASE 
        WHEN reactor_user_id = $2 THEN reaction_code_point 
        ELSE null
      END AS client_reaction
    FROM "CommentsOnPost_RepliesToCommentView"
    WHERE main_comment_id = $1
    `,
    values: [comment_or_reply_id, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

export const getAllReactorsToPost_OR_Comment = async ({
  post_or_comment,
  post_or_comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS user_id, 
      "user".profile_pic_url, 
      "user".username, 
      "user".name,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "PostCommentReaction" "reaction" 
    INNER JOIN "User" "user" ON "reaction".reactor_user_id = "user".id 
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "user".id AND "client_follows".follower_user_id = $2
    WHERE "reaction".${post_or_comment}_id = $1
    ORDER BY "reaction".created_at DESC
    LIMIT $3 OFFSET $4`,
    values: [post_or_comment_id, client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

export const getAllReactorsWithReactionToPost_OR_Comment = async ({
  post_or_comment,
  post_or_comment_id,
  reaction_code_point,
  client_user_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS user_id, 
      "user".profile_pic_url, 
      "user".username, 
      "user".name,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "PostCommentReaction" "reaction" 
    INNER JOIN "User" "user" ON "reaction".reactor_user_id = "user".id 
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "user".id AND "client_follows".follower_user_id = $3
    WHERE "reaction".${post_or_comment}_id = $1 AND "reaction".reaction_code_point = $2
    ORDER BY "reaction".created_at DESC
    LIMIT $4 OFFSET $5`,
    values: [
      post_or_comment_id,
      reaction_code_point,
      client_user_id,
      limit,
      offset,
    ],
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
    text: `
    WITH pc_reaction AS (
      DELETE FROM "PostCommentReaction" WHERE ${post_or_comment}_id = $1 AND reactor_user_id = $2
    )
    ${
      post_or_comment === "post"
        ? `
      SELECT reactions_count 
      FROM "AllPostsView" 
      WHERE post_id = $1`
        : `
      SELECT reactions_count 
      FROM "CommentsOnPost_RepliesToCommentView" 
      WHERE main_comment_id = $1`
    }`,
    values: [post_or_comment_id, reactor_user_id],
  }

  return (await dbQuery(query)).rows[0].reactions_count
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
    text: `
    WITH dsp AS (
      DELETE  FROM "SavedPost" WHERE post_id = $1 AND saver_user_id = $2
    )
    SELECT saves_count FROM "AllPostsView" WHERE post_id = $1`,
    values: [post_id, saver_user_id],
  }

  return (await dbQuery(query)).rows[0].saves_count
}
