import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

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
    const { records } = await neo4jDriver.executeQuery(
      `
      CREATE (user:User{ id: randomUUID(), email: $info.email, username: $info.username, password: $info.password, name: $info.name, birthday: $info.birthday, bio: $info.bio, profile_pic_url: "", connection_status: "online" })
      RETURN user.id AS id, user.email AS email, user.username AS username, user.name AS name, user.profile_pic_url AS profile_pic_url, user.connection_status AS connection_status
      `,
      { info }
    )

    return records[0].toObject()
  }

  /**
   * @param {number | string} uniqueIdentifier
   */
  static async findOne(uniqueIdentifier) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ id: $uniqueIdentifier } | User{ username: $uniqueIdentifier } | User{ email: $uniqueIdentifier })
      RETURN user.id AS id, user.email AS email, user.username AS username, user.name AS name, user.profile_pic_url AS profile_pic_url, user.connection_status AS connection_status
      `,
      { uniqueIdentifier }
    )

    return records[0].toObject()
  }

  /**
   * @param {string} emailOrUsername
   */
  static async findOneIncPassword(emailOrUsername) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ username: $uniqueIdentifier } | User{ email: $uniqueIdentifier })
      RETURN user.id AS id, user.email AS email, user.username AS username, user.name AS name, user.profile_pic_url AS profile_pic_url, user.connection_status AS connection_status, user.password AS password
      `,
      { emailOrUsername }
    )

    return records[0].toObject()
  }

  /**
   * @param {string | number} uniqueIdentifier
   * @returns {Promise<boolean>}
   */
  static async exists(uniqueIdentifier) {
    const { records } = await neo4jDriver.executeQuery(
      `RETURN EXISTS { MATCH (user:User{ id: $uniqueIdentifier } | User{ username: $uniqueIdentifier } | User{ email: $uniqueIdentifier }) } AS userExists`,
      { uniqueIdentifier }
    )

    return records[0].get("userExists")
  }

  static async changePassword(email, newPassword) {
    await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ email: $email })
      SET user.password = $newPassword
      `,
      { email, newPassword }
    )
  }

  /**
   * @param {string} client_user_id
   * @param {string} to_follow_user_id
   */
  static async followUser(client_user_id, to_follow_user_id) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (tofollowUser:User{ id: $to_follow_user_id })
      CREATE (followNotif:Notification:FollowNotification{ id: randomUUID(), type: "follow", follower_user: { id: clientUser.id, username: clientUser.username, profile_pic_url: clientUser.profile_pic_url }, is_read: false, created_at: timestamp() }), 
        (clientUser)-[:FOLLOWS]->(tofollowUser)<-[:RECEIVES_NOTIFICATION]-(followNotif)
      RETURN followNotif.id AS id, followNotif.type AS type, followNotif.follower_user AS follower_user
      `,
      { client_user_id, to_follow_user_id }
    )

    return records[0].toObject()
  }

  /**
   * @param {string} client_user_id
   * @param {string} to_unfollow_user_id
   */
  static async unfollowUser(client_user_id, to_unfollow_user_id) {
    await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[fr:FOLLOWS]->(tounfollowUser:User{ id: $to_unfollow_user_id })
      DELETE fr
      `,
      { client_user_id, to_unfollow_user_id }
    )
  }

  /**
   * @param {string} client_user_id
   * @param {Object<string, any>} updateKVs
   */
  static async edit(client_user_id, updateKVs) {

    // construct SET key = $key, key = $key, ... from updateKVs keys
    let setUpdates = ""

    for (const key of Object.keys(updateKVs)) {
      if (setUpdates) {
        setUpdates = setUpdates + ", "
      }

      setUpdates = `${setUpdates}user.${key} = $${key}`
    }

    await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ id: $client_user_id })
      SET ${setUpdates}
      `,
      { client_user_id, ...updateKVs /* deconstruct the key:value in params */ } 
    )
  }

  static async changeProfilePicture(client_user_id, profile_pic_url) {
    await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ id: $client_user_id })
      SET user.profile_pic_url = $profile_pic_url
      `,
      { client_user_id, profile_pic_url } 
    )
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
    await neo4jDriver.executeQuery(
      `
      MATCH (user:User{ id: $client_user_id })
      SET user.connection_status = $connection_status, user.last_active = $last_active
      `,
      { client_user_id, connection_status, last_active } 
    )
  }

  static async readNotification(notification_id) {
    await neo4jDriver.executeQuery(
      `
      MATCH (notif:Notification{ id: $notification_id })
      SET notif.is_read = true
      `,
      { notification_id } 
    )
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

  static async getUnreadNotificationsCount(client_user_id) {
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

  /**
   * @param {number} user_id
   * @returns {Promise<number[]>}
   */
  static async getFolloweesIds(user_id) {
    const query = {
      text: `
    SELECT array_agg(followee_user_id) ids
    FROM follow
    WHERE follower_user_id = $1`,
      values: [user_id],
    }

    return (await dbQuery(query)).rows[0].ids
  }
}
