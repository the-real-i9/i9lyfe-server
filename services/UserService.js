import {
  createFollowNotification,
  followUser,
  getMentionedPosts,
  getReactedPosts,
  getSavedPosts,
  getUnreadNotifications,
  getUserFollowers,
  getUserFollowing,
  getUserPosts,
  getUserProfile,
  unfollowUser,
  updateUserProfile,
} from "../models/UserModel.js"
import { getDBClient } from "../models/db.js"
import { NotificationService } from "./NotificationService.js"

export class UserService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  async follow(to_follow_user_id) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const new_follow_id = await followUser({
        client_user_id: this.client_user_id,
        to_follow_user_id,
      }, dbClient)
      
      await this.#createFollowNotification({
        followee_user_id: to_follow_user_id,
        new_follow_id,
      }, dbClient)
      dbClient.query("COMMIT")
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async #createFollowNotification({ followee_user_id, new_follow_id }, dbClient) {
    const notifData = await createFollowNotification({
      client_user_id: this.client_user_id,
      followee_user_id,
      new_follow_id,
    }, dbClient)

    const { receiver_user_id, ...restData } = notifData
    new NotificationService(receiver_user_id).pushNotification(restData)
  }

  async updateProfile(updatedUserInfoKVPairs) {
    const result = await updateUserProfile(
      this.client_user_id,
      updatedUserInfoKVPairs
    )

    const updatedUserData = result.rows[0]

    return updatedUserData
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

  /** @param {Date} from_date  */
  async getNotifications(from_date) {
    return await getUnreadNotifications(this.client_user_id, from_date)
  }

  /* DELETEs */
  async unfollow(followee_user_id) {
    await unfollowUser(this.client_user_id, followee_user_id)
  }
}
