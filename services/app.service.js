import * as AppModel from "../models/app.model.js"

export class AppService {
  static async getExplorePosts(client_user_id) {
    return await AppModel.getAllPosts(client_user_id)
  }

  /**
   * @param {object} param0
   * @param {string} param0.search
   * @param {"all" | "user" | "photo" | "video" | "reel" | "story" | "hashtag"} param0.filter
   */
  static async searchAndFilter({ search, filter, client_user_id }) {
    return filter === "hashtag"
      ? await AppModel.searchHashtags(search)
      : filter === "user"
      ? await AppModel.searchUsers(search)
      : await AppModel.searchAndFilterPosts({
          search,
          filter,
          client_user_id,
        })
  }

  static async getHashtagPosts(hashtag_name, client_user_id) {
    return await AppModel.getHashtagPosts(hashtag_name, client_user_id)
  }
}
