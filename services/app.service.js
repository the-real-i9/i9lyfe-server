import * as AM from "../models/app.model.js"

export class AppService {
  static async getFeedPosts({ client_user_id, limit, offset }) {
    return await AM.getFeedPosts({ client_user_id, limit, offset })
  }

  static async getExplorePosts(client_user_id) {
    return await AM.getAllPosts(client_user_id)
  }

  /**
   * @param {object} param0
   * @param {string} param0.search
   * @param {"all" | "user" | "photo" | "video" | "reel" | "story" | "hashtag"} param0.filter
   */
  static async searchAndFilter({ search, filter, client_user_id }) {
    return filter === "hashtag"
      ? await AM.searchHashtags(search)
      : filter === "user"
      ? await AM.searchUsers(search)
      : await AM.searchAndFilterPosts({
          search,
          filter,
          client_user_id,
        })
  }

  static async getHashtagPosts(hashtag_name, client_user_id) {
    return await AM.getHashtagPosts(hashtag_name, client_user_id)
  }
}
