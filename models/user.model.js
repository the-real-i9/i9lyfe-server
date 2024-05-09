import { dbQuery } from "./db.js"

/** @typedef {import("pg").QueryConfig} PgQueryConfig */

/**
 * @param {Object} fields
 * @param {string} fields.email
 * @param {string} fields.username
 * @param {string} fields.password
 * @param {string} fields.name
 * @param {Date} fields.birthday
 * @param {string} fields.bio
 */
export const createUser = async (fields) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `SELECT new_user FROM create_user($1, $2, $3, $4, $5, $6)`,
    values: [
      fields.email,
      fields.username,
      fields.password,
      fields.name,
      fields.birthday,
      fields.bio,
    ],
  }

  return (await dbQuery(query)).rows[0].new_user
}

/**
 * @param {number | string} uniqueIdentifier
 * @param {boolean} forAuth
 */
export const getUser = async (uniqueIdentifier, forAuth) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `SELECT res_user FROM get_user($1)`,
    values: [uniqueIdentifier],
  }

  const { res_user } = (await dbQuery(query)).rows[0]

  if (res_user && !forAuth) {
    delete res_user.password
  }

  return res_user
}

/**
 * @param {string | number} uniqueIdentifier
 * @returns {Promise<boolean>}
 */
export const userExists = async (uniqueIdentifier) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `SELECT check_res FROM user_exists($1)`,
    values: [uniqueIdentifier],
  }

  return (await dbQuery(query)).rows[0].check_res
}

export const changeUserPassword = async (email, newPassword) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "UPDATE i9l_user SET password = $2 WHERE email = $1;",
    values: [email, newPassword],
  }

  await dbQuery(query)
}

/**
 * @param {number} client_user_id
 * @param {number} to_follow_user_id
 */
export const followUser = async (client_user_id, to_follow_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT follow_notif FROM follow_user($1, $2)",
    values: [client_user_id, to_follow_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

/**
 * @param {number} client_user_id
 * @param {number} followee_user_id
 */
export const unfollowUser = async (client_user_id, followee_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "DELETE FROM follow WHERE follower_user_id = $1 AND followee_user_id = $2;",
    values: [client_user_id, followee_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} client_user_id
 * @param {[string, any][]} updateKVPairs
 */
export const editUser = async (client_user_id, updateKVPairs) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT edit_user($1, $2)",
    values: [client_user_id, updateKVPairs],
  }

  await dbQuery(query)
}

export const uploadProfilePicture = async (client_user_id, profile_pic_url) => {
  await editUser(client_user_id, [["profile_pic_url", profile_pic_url]])
}

/* ************* */
export const getFeedPosts = async ({ client_user_id, limit, offset }) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_feed_posts($1, $2, $3)",
    values: [client_user_id, limit, offset],
  }

  return (await dbQuery(query)).rows
}

/** @param {string} username */
export const getUserProfile = async (username, client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT profile_data FROM get_user_profile($1, $2)",
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows[0].profile_data
}

// GET user followers
export const getUserFollowers = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_user_followers($1, $2, $3, $4)",
    values: [username, limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET user following
export const getUserFollowing = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_user_following($1, $2, $3, $4)",
    values: [username, limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET user posts
/** @param {string} username */
export const getUserPosts = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_user_posts($1, $2, $3, $4)",
    values: [username, limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts user has been mentioned in
export const getMentionedPosts = async ({ limit, offset, client_user_id }) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_mentioned_posts($1, $2, $3)",
    values: [limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts reacted by user
export const getReactedPosts = async ({ limit, offset, client_user_id }) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_reacted_posts($1, $2, $3)",
    values: [limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts saved by this user
export const getSavedPosts = async ({ limit, offset, client_user_id }) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT * FROM get_saved_posts($1, $2, $3)",
    values: [limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

/**
 * @param {object} param0
 * @param {"online" | "offline"} param0.connection_status
 * @param {Date} param0.last_active
 */
export const updateUserConnectionStatus = async ({
  client_user_id,
  connection_status,
  last_active,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `UPDATE i9l_user SET connection_status = $1, last_active = $2 WHERE id = $3`,
    values: [connection_status, last_active, client_user_id],
  }

  await dbQuery(query)
}

export const readUserNotification = async (notification_id, client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    UPDATE notification SET is_read = true
    WHERE id = $1 AND receiver_user_id = $2`,
    values: [notification_id, client_user_id],
  }

  await dbQuery(query)
}

// GET user notifications
/**
 *
 * @param {number} client_user_id
 * @param {Date} from
 */
export const getUserNotifications = async ({
  client_user_id,
  from,
  limit,
  offset,
}) => {
  const query = {
    text: "SELECT user_notifications FROM get_user_notifications($1, $2, $3, $4)",
    values: [client_user_id, from, limit, offset],
  }

  return (await dbQuery(query)).rows[0].user_notifications
}

export const getUnreadNotificationsCount = async (client_user_id) => {
  const query = {
    text: `
    SELECT COUNT(id) AS count 
    FROM notification 
    WHERE receiver_user_id = $1 AND is_read = false
    `,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows[0].count
}

export const getUserFolloweesIds = async (user_id) => {
  const query = {
    text: `
    SELECT array_agg(followee_user_id) ids
    FROM follow
    WHERE follower_user_id = $1`,
    values: [user_id],
  }

  return (await dbQuery(query)).rows[0].ids
}
