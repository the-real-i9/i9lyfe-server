
import { User } from "../models/user.model.js"
import { NotificationService } from "./notification.service.js"

export class UserService {
  static async getClientUser(client_user_id) {
    return await User.findOne(client_user_id)
  }

  static async follow(client_user_id, to_follow_user_id) {
    const { follow_notif } = await User.followUser(
      client_user_id,
      to_follow_user_id
    )

    const { receiver_user_id, ...restData } = follow_notif
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  static async editProfile(client_user_id, updateKVPairs) {
    await User.edit(client_user_id, updateKVPairs)
  }

  static async updateConnectionStatus({
    client_user_id,
    connection_status,
    last_active,
  }) {
    await User.updateConnectionStatus({
      client_user_id,
      connection_status,
      last_active,
    })
  }

  static async readNotification(notification_id, client_user_id) {
    await User.readNotification(notification_id, client_user_id)
  }

  static async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  static async getFeedPosts({ client_user_id, limit, offset }) {
    return await User.getFeedPosts({ client_user_id, limit, offset })
  }

  static async getProfile(username, client_user_id) {
    return await User.getProfile(username, client_user_id)
  }

  static async getFollowers({ username, limit, offset, client_user_id }) {
    return await User.getFollowers({
      username,
      limit,
      offset,
      client_user_id,
    })
  }

  static async getFollowing({ username, limit, offset, client_user_id }) {
    return await User.getFollowing({
      username,
      limit,
      offset,
      client_user_id,
    })
  }

  static async getPosts({ username, limit, offset, client_user_id }) {
    return await User.getPosts({ username, limit, offset, client_user_id })
  }

  static async getMentionedPosts({ limit, offset, client_user_id }) {
    return await User.getMentionedPosts({ limit, offset, client_user_id })
  }

  static async getReactedPosts({ limit, offset, client_user_id }) {
    return await User.getReactedPosts({ limit, offset, client_user_id })
  }

  static async getSavedPosts({ limit, offset, client_user_id }) {
    return await User.getSavedPosts({ limit, offset, client_user_id })
  }

  static async getNotifications({ client_user_id, from, limit, offset }) {
    return await User.getNotifications({
      client_user_id,
      from: new Date(from),
      limit,
      offset,
    })
  }

  static async unfollow(client_user_id, to_unfollow_user_id) {
    await User.unfollowUser(client_user_id, to_unfollow_user_id)
  }
}
