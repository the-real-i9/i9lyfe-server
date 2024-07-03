import { App } from "../models/app.model.js"


export class AppService {
  static async getExplorePosts({ limit, offset, client_user_id }) {
    return await App.getExplorePosts({ limit, offset, client_user_id })
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
      ? await App.searchHashtags({ search, limit, offset })
      : filter === "user"
      ? await App.searchUsers({ search, limit, offset })
      : await App.searchAndFilterPosts({
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
    return await App.getHashtagPosts({
      hashtag_name,
      limit,
      offset,
      client_user_id,
    })
  }

  static async searchUsersToChat({ client_user_id, search, limit, offset }) {
    return await App.searchUsersToChat({
      search,
      limit,
      offset,
      client_user_id,
    })
  }
}
