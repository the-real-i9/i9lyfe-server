import * as UM from "../models/user.model.js"
import { NotificationService } from "./notification.service.js"

export class UserService {
  static async getClientUser(client_user_id) {
    return await UM.getUser(client_user_id)
  }

  static async follow(client_user_id, to_follow_user_id) {
    const { follow_notif } = await UM.followUser(
      client_user_id,
      to_follow_user_id
    )

    const { receiver_user_id, ...restData } = follow_notif
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  static async editProfile(client_user_id, updateKVPairs) {
    await UM.editUser(client_user_id, updateKVPairs)
  }

  static async updateConnectionStatus({
    client_user_id,
    connection_status,
    last_active,
  }) {
    await UM.updateUserConnectionStatus({
      client_user_id,
      connection_status,
      last_active,
    })
  }

  static async readNotification(notification_id, client_user_id) {
    await UM.readUserNotification(notification_id, client_user_id)
  }

  static async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  static async getFeedPosts({ client_user_id, limit, offset }) {
    return await UM.getFeedPosts({ client_user_id, limit, offset })
  }

  static async getProfile(username, client_user_id) {
    return await UM.getUserProfile(username, client_user_id)
  }

  static async getFollowers({ username, limit, offset, client_user_id }) {
    return await UM.getUserFollowers({
      username,
      limit,
      offset,
      client_user_id,
    })
  }

  static async getFollowing({ username, limit, offset, client_user_id }) {
    return await UM.getUserFollowing({
      username,
      limit,
      offset,
      client_user_id,
    })
  }

  static async getPosts({ username, limit, offset, client_user_id }) {
    return await UM.getUserPosts({ username, limit, offset, client_user_id })
  }

  static async getMentionedPosts({ limit, offset, client_user_id }) {
    return await UM.getMentionedPosts({ limit, offset, client_user_id })
  }

  static async getReactedPosts({ limit, offset, client_user_id }) {
    return await UM.getReactedPosts({ limit, offset, client_user_id })
  }

  static async getSavedPosts({ limit, offset, client_user_id }) {
    return await UM.getSavedPosts({ limit, offset, client_user_id })
  }

  /** @param {Date} from  */
  static async getUserNotifications({ client_user_id, from, limit, offset }) {
    return await UM.getUserNotifications({
      client_user_id,
      from,
      limit,
      offset,
    })
  }

  /* DELETEs */
  static async unfollow(client_user_id, followee_user_id) {
    await UM.unfollowUser(client_user_id, followee_user_id)
  }
}
