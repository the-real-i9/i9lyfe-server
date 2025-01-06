import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class User {
  /**
   * @param {Object} info
   * @param {string} info.email
   * @param {string} info.username
   * @param {string} info.password
   * @param {string} info.name
   * @param {string} info.birthday
   * @param {string} info.bio
   */
  static async create(info) {
    info.birthday = new Date(info.birthday).toISOString()
    
    const { records } = await neo4jDriver.executeWrite(
      `
      CREATE (user:User{ id: randomUUID(), email: $info.email, username: $info.username, password: $info.password, name: $info.name, birthday: datetime($info.birthday), bio: $info.bio, profile_pic_url: "", connection_status: "offline" })
      RETURN user {.id, .email, .username, .name, .profile_pic_url, .connection_status } AS new_user
      `,
      { info },
    )

    return records[0].get("new_user")
  }

  /**
   * @param {string} uniqueIdentifier
   */
  static async findOne(uniqueIdentifier) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (user:User)
      WHERE user.id = $uniqueIdentifier OR user.username = $uniqueIdentifier OR user.email = $uniqueIdentifier
      RETURN user {.id, .email, .username, .name, .profile_pic_url, .connection_status } AS found_user
      `,
      { uniqueIdentifier }
    )

    return records[0].get("found_user")
  }

  /**
   * @param {string} emailOrUsername
   */
  static async findOneIncPassword(emailOrUsername) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (user:User)
      WHERE user.username = $uniqueIdentifier OR user.email = $uniqueIdentifier
      RETURN user {.id, .email, .username, .name, .profile_pic_url, .connection_status, .password } AS found_user
      `,
      { emailOrUsername }
    )

    return records[0].get("found_user")
  }

  /**
   * @param {string | number} uniqueIdentifier
   * @returns {Promise<boolean>}
   */
  static async exists(uniqueIdentifier) {
    const { records } = await neo4jDriver.executeWrite(
      `RETURN EXISTS {
        MATCH (user:User)
        WHERE user.username = $uniqueIdentifier OR user.email = $uniqueIdentifier
      } AS userExists`,
      { uniqueIdentifier }
    )

    return records[0].get("userExists")
  }

  static async changePassword(email, newPassword) {
    await neo4jDriver.executeWrite(
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
    if (client_user_id === to_follow_user_id) {
      return { follow_notif: null }
    }
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (tofollowUser:User{ id: $to_follow_user_id })
      CREATE (followNotif:Notification:FollowNotification{ id: randomUUID(), type: "follow", is_read: false, created_at: datetime() })-[:FOLLOWER_USER]->(clientUser), 
        (clientUser)-[:FOLLOWS_USER { user_to_user: $user_to_user }]->(tofollowUser)-[:RECEIVES_NOTIFICATION]->(followNotif)
      WITH followNotif, clientUser { .id, .username, .profile_pic_url } AS clientUserView
      RETURN followNotif { .id, .type, follower_user: clientUserView } AS follow_notif
      `,
      { client_user_id, to_follow_user_id, user_to_user: `user-${client_user_id}_to_user-${to_follow_user_id}` }
    )

    return records[0].toObject()
  }

  /**
   * @param {string} client_user_id
   * @param {string} to_unfollow_user_id
   */
  static async unfollowUser(client_user_id, to_unfollow_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH ()-[fr:FOLLOWS_USER { user_to_user: $user_to_user }]->()
      DELETE fr
      `,
      { user_to_user: `user-${client_user_id}_to_user-${to_unfollow_user_id}` }
    )
  }

  /**
   * @param {string} client_user_id
   * @param {Object<string, any>} updateKVs
   */
  static async edit(client_user_id, updateKVs) {
    if (updateKVs.birthday) {
      updateKVs.birthday = new Date(updateKVs.birthday).toISOString()
    }

    // construct SET key = $key, key = $key, ... from updateKVs keys
    let setUpdates = ""

    for (const key of Object.keys(updateKVs)) {
      if (setUpdates) {
        setUpdates = setUpdates + ", "
      }

      if (key === "birthday") {
        setUpdates = `${setUpdates}user.${key} = datetime($${key})`
        continue
      }

      setUpdates = `${setUpdates}user.${key} = $${key}`
    }

    await neo4jDriver.executeWrite(
      `
      MATCH (user:User{ id: $client_user_id })
      SET ${setUpdates}
      `,
      { client_user_id, ...updateKVs /* deconstruct the key:value in params */ } 
    )
  }

  static async changeProfilePicture(client_user_id, profile_pic_url) {
    await neo4jDriver.executeWrite(
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
   * @param {string|null} param0.last_active
   */
  static async updateConnectionStatus({
    client_user_id,
    connection_status,
    last_active,
  }) {
    last_active = last_active ? new Date(last_active).toISOString() : null

    const last_active_param = last_active ? "datetime($last_active)" : "$last_active"

    await neo4jDriver.executeWrite(
      `
      MATCH (user:User{ id: $client_user_id })
      SET user.connection_status = $connection_status, user.last_active = ${last_active_param}
      `,
      { client_user_id, connection_status, last_active } 
    )
  }

  static async readNotification(notification_id) {
    await neo4jDriver.executeWrite(
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
   * @param {string} from
   */
  static async getNotifications({ client_user_id, from, limit, offset }) {
    from = new Date(from).toISOString()

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
