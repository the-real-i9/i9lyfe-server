import { dbQuery } from "./db.js"

export class App {
  static async getExplorePosts({ limit, offset, client_user_id }) {
    const query = {
      text: "SELECT * FROM get_explore_posts($1)",
      values: [limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  static async searchAndFilterPosts({
    search,
    filter,
    limit,
    offset,
    client_user_id,
  }) {
    const query = {
      text: "SELECT * FROM search_filter_posts($1, $2, $3, $4, $5)",
      values: [search, filter, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  static async searchHashtags({ search, limit, offset }) {
    const query = {
      text: `
    SELECT hashtag_name, COUNT(post_id) AS posts_count 
    FROM pc_hashtag
    WHERE hashtag_name ILIKE $1
    GROUP BY hashtag_name
    LIMIT $2 OFFSET $3`,
      values: [`%${search}%`, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async searchUsers({ search, limit, offset }) {
    const query = {
      text: `
    SELECT id AS user_id, 
      username, 
      name, 
      profile_pic_url
    FROM i9l_user
    WHERE username ILIKE $1 OR name ILIKE $1
    LIMIT $2 OFFSET $3`,
      values: [`%${search}%`, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async getHashtagPosts({
    hashtag_name,
    limit,
    offset,
    client_user_id,
  }) {
    const query = {
      text: "SELECT * FROM get_hashtag_posts($1, $2, $3, $4)",
      values: [hashtag_name, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  /**
   * @param {string} search
   */
  static async searchUsersToChat({ search, limit, offset, client_user_id }) {
    const query = {
      text: "SELECT * FROM get_users_to_chat($1, $2, $3, $4)",
      values: [`%${search}%`, limit, offset, client_user_id],
    }

    return (await dbQuery(query)).rows
  }
}
