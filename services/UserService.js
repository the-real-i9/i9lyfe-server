import {
  followUser,
  getMentionedPosts,
  getReactedPosts,
  getSavedPosts,
  getUnreadNotifications,
  getUserFollowers,
  getUserFollowing,
  getUserPosts,
  getUserProfile,
  readUserNotification,
  unfollowUser,
  updateUserConnectionStatus,
  updateUserProfile,
} from "../models/UserModel.js"
import { NotificationService } from "./NotificationService.js"

export class UserService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  async follow(to_follow_user_id) {
    const followNotifData = await followUser({
      client_user_id: this.client_user_id,
      to_follow_user_id,
    })

    const { receiver_user_id, ...restData } = followNotifData
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  async updateProfile(updateKVPairs) {
    return await updateUserProfile(this.client_user_id, updateKVPairs)
  }

  async updateConnectionStatus(new_connection_status) {
    await updateUserConnectionStatus(this.client_user_id, new_connection_status)
  }

  async readNotification(notification_id) {
    await readUserNotification(notification_id, this.client_user_id)
  }

  async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  async getProfile(username) {
    return await getUserProfile(username, this.client_user_id)
  }

  async getFollowers(username) {
    return await getUserFollowers(username, this.client_user_id)
  }

  async getFollowing(username) {
    return await getUserFollowing(username, this.client_user_id)
  }

  async getPosts(username) {
    return await getUserPosts(username, this.client_user_id)
  }

  async getMentionedPosts() {
    return await getMentionedPosts(this.client_user_id)
  }

  async getReactedPosts() {
    return await getReactedPosts(this.client_user_id)
  }

  async getSavedPosts() {
    return await getSavedPosts(this.client_user_id)
  }

  /** @param {Date} from  */
  async getNotifications(from) {
    return await getUnreadNotifications(this.client_user_id, from)
  }

  /* DELETEs */
  async unfollow(followee_user_id) {
    await unfollowUser(this.client_user_id, followee_user_id)
  }
}
