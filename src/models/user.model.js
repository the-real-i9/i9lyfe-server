import { dbQuery } from "../configs/db.js"

/** @typedef {import("pg").QueryConfig} PgQueryConfig */

export class User {
  /**
   * @param {Object} info
   * @param {string} info.email
   * @param {string} info.username
   * @param {string} info.password
   * @param {string} info.name
   * @param {Date} info.birthday
   * @param {string} info.bio
   */
  static async create(info) {
    
    /** @type {PgQueryConfig} */
    const query = {
      text: `SELECT * FROM create_user($1, $2, $3, $4, $5, $6)`,
      values: [
        info.email,
        info.username,
        info.password,
        info.name,
        info.birthday,
        info.bio,
      ],
    }

    return (await dbQuery(query)).rows[0]
  }

  /**
   * @param {number | string} uniqueIdentifier
   */
  static async findOne(uniqueIdentifier) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `SELECT * FROM get_user($1)`,
      values: [uniqueIdentifier],
    }

    return (await dbQuery(query)).rows[0]
  }

  /**
   * @param {string} emailOrUsername
   */
  static async findOneIncPassword(emailOrUsername) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `SELECT * FROM get_user($1), get_user_password($1)`,
      values: [emailOrUsername],
    }

    return (await dbQuery(query)).rows[0]
  }

  /**
   * @param {string | number} uniqueIdentifier
   * @returns {Promise<boolean>}
   */
  static async exists(uniqueIdentifier) {
    /** @type {PgQueryConfig} */
    const query = {
      text: `SELECT check_res FROM user_exists($1)`,
      values: [uniqueIdentifier],
    }

    return (await dbQuery(query)).rows[0].check_res
  }

  static async changePassword(email, newPassword) {
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
  static async followUser(client_user_id, to_follow_user_id) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT follow_notif FROM follow_user($1, $2)",
      values: [client_user_id, to_follow_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  /**
   * @param {number} client_user_id
   * @param {number} to_unfollow_user_id
   */
  static async unfollowUser(client_user_id, to_unfollow_user_id) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "DELETE FROM follow WHERE follower_user_id = $1 AND followee_user_id = $2;",
      values: [client_user_id, to_unfollow_user_id],
    }

    await dbQuery(query)
  }

  /**
   * @param {number} client_user_id
   * @param {[string, any][]} updateKVPairs
   */
  static async edit(client_user_id, updateKVPairs) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT edit_user($1, $2)",
      values: [client_user_id, updateKVPairs],
    }

    await dbQuery(query)
  }

  static async changeProfilePicture(client_user_id, profile_pic_url) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "UPDATE i9l_user SET profile_pic_url = $2 WHERE id = $1",
      values: [client_user_id, profile_pic_url],
    }

    await dbQuery(query)
  }

  /**
   * The stored function `fetch_home_feed_posts` aggregates posts
   * based on a "post recommendation algorithm"
   */
  static async getHomeFeedPosts({ client_user_id, limit, offset }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM fetch_home_feed_posts($1, $2, $3)",
      values: [client_user_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  /** @param {string} username */
  static async getProfile(username, client_user_id) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_user_profile($1, $2)",
      values: [username, client_user_id],
    }

    return (await dbQuery(query)).rows[0]
  }

  // GET user followers
  static async getFollowers({ username, limit, offset, client_user_id }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_user_followers($1, $2, $3, $4)",
      values: [username, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  // GET user following
  static async getFollowing({ username, limit, offset, client_user_id }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_user_following($1, $2, $3, $4)",
      values: [username, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  // GET user posts
  /** @param {string} username */
  static async getPosts({ username, limit, offset, client_user_id }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_user_posts($1, $2, $3, $4)",
      values: [username, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  // GET posts user has been mentioned in
  static async getMentionedPosts({ limit, offset, client_user_id }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_mentioned_posts($1, $2, $3)",
      values: [limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  // GET posts reacted by user
  static async getReactedPosts({ limit, offset, client_user_id }) {
    /** @type {PgQueryConfig} */
    const query = {
      text: "SELECT * FROM get_reacted_posts($1, $2, $3)",
      values: [limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  // GET posts saved by this user
  static async getSavedPosts({ limit, offset, client_user_id }) {
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
  static async updateConnectionStatus({
    client_user_id,
    connection_status,
    last_active,
  }) {
    
    /** @type {PgQueryConfig} */
    const query = {
      text: `UPDATE i9l_user SET connection_status = $1, last_active = $2 WHERE id = $3`,
      values: [connection_status, last_active, client_user_id],
    }

    await dbQuery(query)
  }

  static async readNotification(notification_id, client_user_id) {
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
  static async getNotifications({ client_user_id, from, limit, offset }) {
    const query = {
      text: "SELECT * FROM get_user_notifications($1, $2, $3, $4)",
      values: [client_user_id, from, limit, offset],
    }

    return (await dbQuery(query)).rows
  }
}
