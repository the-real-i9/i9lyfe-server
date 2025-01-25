import { neo4jDriver } from "../configs/db.js"

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
      CREATE (user:User{ email: $info.email, username: $info.username, password: $info.password, name: $info.name, birthday: datetime($info.birthday), bio: $info.bio, profile_pic_url: "", connection_status: "offline" })
      RETURN user { .email, .username, .name, .profile_pic_url, .connection_status } AS new_user
      `,
      { info }
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
      WHERE user.username = $uniqueIdentifier OR user.email = $uniqueIdentifier
      RETURN user { .email, .username, .name, .profile_pic_url, .connection_status, .password } AS found_user
      `,
      { uniqueIdentifier }
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
   * @param {string} client_username
   * @param {string} to_follow_username
   */
  static async followUser(client_username, to_follow_username) {
    if (client_username === to_follow_username) {
      return { follow_notif: null }
    }
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ username: $client_username }), (tofollowUser:User{ username: $to_follow_username })
      MERGE (clientUser)-[:FOLLOWS_USER]->(tofollowUser)
      
      CREATE (tofollowUser)-[:RECEIVES_NOTIFICATION]->(followNotif:Notification:FollowNotification{ id: randomUUID(), type: "follow", is_read: false, created_at: datetime(), follower_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

      WITH followNotif, toString(followNotif.created_at) AS created_at
      RETURN followNotif { .*,  created_at } AS follow_notif
      `,
      { client_username, to_follow_username }
    )

    return records[0].toObject()
  }

  /**
   * @param {string} client_username
   * @param {string} to_unfollow_username
   */
  static async unfollowUser(client_username, to_unfollow_username) {
    await neo4jDriver.executeWrite(
      `
      MATCH (:User{ username: $client_username })-[fr:FOLLOWS_USER]->(:User{ username: $to_unfollow_username })
      DELETE fr
      `,
      { client_username, to_unfollow_username }
    )
  }

  /**
   * @param {string} client_username
   * @param {Object<string, any>} updateKVs
   */
  static async edit(client_username, updateKVs) {
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
      MATCH (user:User{ username: $client_username })
      SET ${setUpdates}
      `,
      { client_username, ...updateKVs /* deconstruct the key:value in params */ }
    )
  }

  static async changeProfilePicture(client_username, profile_pic_url) {
    await neo4jDriver.executeWrite(
      `
      MATCH (user:User{ username: $client_username })
      SET user.profile_pic_url = $profile_pic_url
      `,
      { client_username, profile_pic_url }
    )
  }

  /** @param {string} username */
  static async getProfile(username, client_username) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (profUser:User{ username: $username })

      MATCH (follower:User)-[:FOLLOWS_USER]->(profUser)-[:FOLLOWS_USER]->(following:User),
        (profUser)-[:CREATES_POST]->(post:Post)

      OPTIONAL MATCH (profUser)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })

      WITH profUser,
        count(post) AS posts_count,
        count(follower) AS followers_count,
        count(following) AS followings_count,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      RETURN profUser { .username, .name, .profile_pic_url, .bio, posts_count, followers_count, followings_count, client_follows } AS user_profile
      `,
      { username, client_username }
    )

    return records[0].get("user_profile")
  }

  // GET user followers
  static async getFollowers({ username, limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (follower:User)-[:FOLLOWS_USER]->(:User{ username: $username })

      OPTIONAL MATCH (follower)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })

      WITH follower,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows,
        ORDER BY follower.username
        OFFSET toInteger($offset)
        LIMIT toInteger($limit)
      RETURN collect(follower { .id, .username, .profile_pic_url, client_follows }) AS user_followers
      `,
      { username, client_username, limit, offset }
    )

    return records[0].get("user_followers")
  }

  // GET user following
  static async getFollowings({ username, limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (:User{ username: $username })-[:FOLLOWS_USER]->(following:User)

      OPTIONAL MATCH (following)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })

      WITH following,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY following.username
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(following { .id, .username, .profile_pic_url, client_follows }) AS user_followings
      `,
      { username, client_username, limit, offset }
    )

    return records[0].get("user_followings")
  }

  // GET user posts
  /** @param {string} username */
  static async getPosts({ username, limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (ownerUser:User{ username: $username })-[:CREATES_POST]->(post:Post), (clientUser:User{ username: $client_username })
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        ownerUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        CASE csaves 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_posts
      `,
      { username, client_username, limit, offset }
    )

    return records[0].get("user_posts")
  }

  // GET posts user has been mentioned in
  static async getMentionedPosts({ limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ username: $client_username })<-[:MENTIONS_USER]-(post:Post)
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        CASE csaves 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_mentioned_posts
      `,
      { client_username, limit, offset }
    )

    return records[0].get("user_mentioned_posts")
  }

  // GET posts reacted by user
  static async getReactedPosts({ limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ username: $client_username })-[cxrn:REACTS_TO_POST]->(post:Post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .username, .profile_pic_url } AS owner_user,
        crxn.reaction AS client_reaction, 
        CASE csaves 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_reacted_posts
      `,
      { client_username, limit, offset }
    )

    return records[0].get("user_reacted_posts")
  }

  // GET posts saved by this user
  static async getSavedPosts({ limit, offset, client_username }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ username: $client_username })-[:SAVES_POST]->(post:Post)
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        true AS client_saved, 
        CASE creposts 
          WHEN IS NULL THEN false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_saved_posts
      `,
      { client_username, limit, offset }
    )

    return records[0].get("user_saved_posts")
  }

  /**
   * @param {object} param0
   * @param {"online" | "offline"} param0.connection_status
   * @param {string|null} param0.last_active
   */
  static async updateConnectionStatus({
    client_username,
    connection_status,
    last_active,
  }) {
    last_active = last_active ? new Date(last_active).toISOString() : null

    const last_active_param = last_active
      ? "datetime($last_active)"
      : "$last_active"

    await neo4jDriver.executeWrite(
      `
      MATCH (user:User{ username: $client_username })
      SET user.connection_status = $connection_status, user.last_active = ${last_active_param}
      `,
      { client_username, connection_status, last_active }
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
   * @param {number} client_username
   * @param {string} from
   */
  static async getNotifications({ client_username, limit, offset }) {
    // from = new Date(from).toISOString()

    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ username: $client_username })-[:RECEIVES_NOTIFICATION]->(notif:Notification)
      WITH notif, toString(notif.created_at) AS created_at
      ORDER BY notif.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(notif { .*, created_at }) AS notifications
      `,
      { client_username, limit, offset }
    )

    return records[0].get("notifications")
  }
}
