import * as AppModel from "../models/AppModel.js"

export class AppService {
  async getExplorePosts(client_user_id) {
    return await AppModel.getAllPosts(client_user_id)
  }

  /**
   * @param {object} param0
   * @param {string} param0.search
   * @param {"all" | "photos" | "videos" | "reels" | "stories" | "hashtags"} param0.category
   */
  async searchAndFilter({ search, category, client_user_id }) {
    if (category !== "hashtags") {
      return await AppModel.searchAndFilterPosts({
        search,
        type: category,
        client_user_id,
      })
    }
    return await AppModel.searchHashtags(search)
  }
}
