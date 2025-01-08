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
      MATCH (clientUser:User{ id: $client_user_id })
      MERGE (clientUser)-[:FOLLOWS_USER]->(tofollowUser:User{ id: $to_follow_user_id })
      
      CREATE (tofollowUser)-[:RECEIVES_NOTIFICATION]->(followNotif:Notification:FollowNotification{ id: randomUUID(), type: "follow", is_read: false, created_at: datetime() })-[:FOLLOWER_USER]->(clientUser)

      WITH followNotif, clientUser { .id, .username, .profile_pic_url } AS follower_user
      RETURN followNotif { .id, .type, follower_user } AS follow_notif
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
    await neo4jDriver.executeWrite(
      `
      MATCH (:User{ id: $client_user_id })-[fr:FOLLOWS_USER]->(:User{ id: $to_unfollow_user_id })
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

  /** @param {string} username */
  static async getProfile(username, client_user_id) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (profUser:User{ username: $username })

      MATCH (follower:User)-[:FOLLOWS_USER]->(profUser)-[:FOLLOWS_USER]->(following:User),
        (profUser)-[:CREATES_POST]->(post:Post)

      OPTIONAL MATCH (profUser)<-[fur:FOLLOWS_USER]-(:User{ id: $client_user_id })

      WITH profUser,
        count(post) AS posts_count,
        count(follower) AS followers_count,
        count(following) AS followings_count,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      RETURN profUser { .id, .username, .name, .profile_pic_url, .bio, posts_count, followers_count, followings_count, client_follows } AS user_profile
      `,
      { username, client_user_id }
    )

    return records[0].get("user_profile")
  }

  // GET user followers
  static async getFollowers({ username, limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (follower:User)-[:FOLLOWS_USER]->(:User{ username: $username })

      OPTIONAL MATCH (follower)<-[fur:FOLLOWS_USER]-(:User{ id: $client_user_id })

      WITH follower,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows,
        ORDER BY follower.username
        OFFSET $offset
        LIMIT $limit
      RETURN collect(follower { .id, .username, .profile_pic_url, client_follows }) AS user_followers
      `,
      { username, client_user_id, limit, offset }
    )

    return records[0].get("user_followers")
  }

  // GET user following
  static async getFollowings({ username, limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (:User{ username: $username })-[:FOLLOWS_USER]->(following:User)

      OPTIONAL MATCH (following)<-[fur:FOLLOWS_USER]-(:User{ id: $client_user_id })

      WITH following,
        CASE fur 
          WHEN IS NULL THEN false
          ELSE true 
        END AS client_follows
      ORDER BY following.username
      OFFSET $offset
      LIMIT $limit
      RETURN collect(following { .id, .username, .profile_pic_url, client_follows }) AS user_followings
      `,
      { username, client_user_id, limit, offset }
    )

    return records[0].get("user_followings")
  }

  // GET user posts
  /** @param {string} username */
  static async getPosts({ username, limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (ownerUser:User{ username: $username })-[:CREATES_POST]->(post:Post), (clientUser:User{ id: $client_user_id })
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        ownerUser { .id, .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        CASE csaves 
          WHEN IS NULL false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_posts
      `,
      { username, client_user_id, limit, offset }
    )

    return records[0].get("user_posts")
  }

  // GET posts user has been mentioned in
  static async getMentionedPosts({ limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ id: $client_user_id })<-[:MENTIONS_USER]-(post:Post)
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .id, .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        CASE csaves 
          WHEN IS NULL false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_mentioned_posts
      `,
      { client_user_id, limit, offset }
    )

    return records[0].get("user_mentioned_posts")
  }

  // GET posts reacted by user
  static async getReactedPosts({ limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[cxrn:REACTS_TO_POST]->(post:Post)
      OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .id, .username, .profile_pic_url } AS owner_user,
        crxn.reaction AS client_reaction, 
        CASE csaves 
          WHEN IS NULL false 
          ELSE true 
        END AS client_saved, 
        CASE creposts 
          WHEN IS NULL false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_reacted_posts
      `,
      { client_user_id, limit, offset }
    )

    return records[0].get("user_reacted_posts")
  }

  // GET posts saved by this user
  static async getSavedPosts({ limit, offset, client_user_id }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:SAVES_POST]->(post:Post)
      OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
      OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
      WITH post, 
        toString(post.created_at) AS created_at, 
        clientUser { .id, .username, .profile_pic_url } AS owner_user,
        CASE crxn 
          WHEN IS NULL THEN "" 
          ELSE crxn.reaction 
        END AS client_reaction, 
        true AS client_saved, 
        CASE creposts 
          WHEN IS NULL false 
          ELSE true 
        END AS client_reposted
      ORDER BY post.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_saved_posts
      `,
      { client_user_id, limit, offset }
    )

    return records[0].get("user_saved_posts")
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

    const last_active_param = last_active
      ? "datetime($last_active)"
      : "$last_active"

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
  static async getNotifications({ client_user_id, limit, offset }) {
    // from = new Date(from).toISOString()

    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientUser:User{ id: $client_user_id })
      MATCH (clientUser)-[:RECEIVES_NOTIFICATION]->(notif:Notification)-->(actionUser:User)
      WITH notif, toString(notif.created_at) AS created_at, actionUser { .username, .profile_pic_url } AS action_user
      ORDER BY notif.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(notif { .*, created_at, action_user }) AS notifications
      `,
      { client_user_id, limit, offset }
    )

    return records[0].get("notifications")
  }
}
