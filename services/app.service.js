import * as AppModel from "../models/app.model.js"

export class AppService {
  static async getExplorePosts(client_user_id) {
    return await AppModel.getAllPosts(client_user_id)
  }

  /**
   * @param {object} param0
   * @param {string} param0.search
   * @param {"all" | "users" | "photos" | "videos" | "reels" | "stories" | "hashtags"} param0.category
   */
  static async searchAndFilter({ search, category, client_user_id }) {
    return category === "hashtags"
      ? await AppModel.searchHashtags(search)
      : category === "users"
      ? await AppModel.searchUsers(search)
      : await AppModel.searchAndFilterPosts({
          search,
          type: category,
          client_user_id,
        })
  }

  static async getHashtagPosts(hashtag_name, client_user_id) {
    return await AppModel.getHashtagPosts(hashtag_name, client_user_id)
  }
}
