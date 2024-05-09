import * as AM from "../models/app.model.js"

export class AppService {
  static async getExplorePosts(client_user_id) {
    return await AM.getExplorePosts(client_user_id)
  }

  /**
   * @param {object} param0
   * @param {string} param0.search
   * @param {"all" | "user" | "photo" | "video" | "reel" | "story" | "hashtag"} param0.filter
   */
  static async searchAndFilter({
    search,
    filter,
    limit,
    offset,
    client_user_id,
  }) {
    return filter === "hashtag"
      ? await AM.searchHashtags({ search, limit, offset })
      : filter === "user"
      ? await AM.searchUsers({ search, limit, offset })
      : await AM.searchAndFilterPosts({
          search,
          filter,
          limit,
          offset,
          client_user_id,
        })
  }

  static async getHashtagPosts({
    hashtag_name,
    limit,
    offset,
    client_user_id,
  }) {
    return await AM.getHashtagPosts({
      hashtag_name,
      limit,
      offset,
      client_user_id,
    })
  }
}
