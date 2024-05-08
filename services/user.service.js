import * as UM from "../models/user.model.js"
import { NotificationService } from "./notification.service.js"

export class UserService {
  static async getClientUser(client_user_id) {
    return await UM.getUser(client_user_id)
  }

  static async follow(client_user_id, to_follow_user_id) {
    const followNotifData = await UM.followUser(client_user_id, to_follow_user_id)

    const { receiver_user_id, ...restData } = followNotifData
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  static async updateProfile(client_user_id, updateKVPairs) {
    await UM.editUser(client_user_id, updateKVPairs)
  }

  static async updateConnectionStatus(client_user_id, new_connection_status) {
    await UM.updateUserConnectionStatus(client_user_id, new_connection_status)
  }

  static async readNotification(notification_id, client_user_id) {
    await UM.readUserNotification(notification_id, client_user_id)
  }

  static async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  static async getProfile(username, client_user_id) {
    return await UM.getUserProfile(username, client_user_id)
  }

  static async getFollowers(username, client_user_id) {
    return await UM.getUserFollowers(username, client_user_id)
  }

  static async getFollowing(username, client_user_id) {
    return await UM.getUserFollowing(username, client_user_id)
  }

  static async getPosts(username, client_user_id) {
    return await UM.getUserPosts(username, client_user_id)
  }

  static async getMentionedPosts(client_user_id) {
    return await UM.getMentionedPosts(client_user_id)
  }

  static async getReactedPosts(client_user_id) {
    return await UM.getReactedPosts(client_user_id)
  }

  static async getSavedPosts(client_user_id) {
    return await UM.getSavedPosts(client_user_id)
  }

  /** @param {Date} from  */
  static async getNotifications(client_user_id, from) {
    return await UM.getUnreadNotifications(client_user_id, from)
  }

  /* DELETEs */
  static async unfollow(client_user_id, followee_user_id) {
    await UM.unfollowUser(client_user_id, followee_user_id)
  }
}
