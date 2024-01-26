import * as AppModel from "../models/AppModel.js"

export class AppService {
  async getExplorePosts(client_user_id) {
    return await AppModel.getAllPosts(client_user_id)
  }
}
