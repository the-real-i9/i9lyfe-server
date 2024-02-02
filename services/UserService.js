import * as UM from "../models/UserModel.js"
import { NotificationService } from "./NotificationService.js"

export class UserService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  async getClientUser() {
    return await UM.getUserById(this.client_user_id, "id username email name profile_pic_url connection_status")
  }

  async follow(to_follow_user_id) {
    const followNotifData = await UM.followUser({
      client_user_id: this.client_user_id,
      to_follow_user_id,
    })

    const { receiver_user_id, ...restData } = followNotifData
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  async updateProfile(updateKVPairs) {
    return await UM.updateUserProfile(this.client_user_id, updateKVPairs)
  }

  async updateConnectionStatus(new_connection_status) {
    await UM.updateUserConnectionStatus(this.client_user_id, new_connection_status)
  }

  async readNotification(notification_id) {
    await UM.readUserNotification(notification_id, this.client_user_id)
  }

  async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  async getProfile(username) {
    return await UM.getUserProfile(username, this.client_user_id)
  }

  async getFollowers(username) {
    return await UM.getUserFollowers(username, this.client_user_id)
  }

  async getFollowing(username) {
    return await UM.getUserFollowing(username, this.client_user_id)
  }

  async getPosts(username) {
    return await UM.getUserPosts(username, this.client_user_id)
  }

  async getMentionedPosts() {
    return await UM.getMentionedPosts(this.client_user_id)
  }

  async getReactedPosts() {
    return await UM.getReactedPosts(this.client_user_id)
  }

  async getSavedPosts() {
    return await UM.getSavedPosts(this.client_user_id)
  }

  /** @param {Date} from  */
  async getNotifications(from) {
    return await UM.getUnreadNotifications(this.client_user_id, from)
  }

  /* DELETEs */
  async unfollow(followee_user_id) {
    await UM.unfollowUser(this.client_user_id, followee_user_id)
  }
}
