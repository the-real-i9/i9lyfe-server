import {
  followUser,
  getMentionedPosts,
  getNotifications,
  getReactedPosts,
  getSavedPosts,
  getUserFollowers,
  getUserFollowing,
  getUserPosts,
  getUserProfile,
  unfollowUser,
  updateUserProfile,
} from "../models/UserModel.js"

export class UserService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  async follow(to_follow_user_id) {
    await followUser(this.client_user_id, to_follow_user_id)
  }

  async updateProfile(updatedUserInfoKVPairs) {
    const result = await updateUserProfile(this.client_user_id, updatedUserInfoKVPairs)

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

  async getNotifications() {
    return await getNotifications(this.client_user_id)
  }

  /* DELETEs */
  async unfollow(followee_user_id) {
    await unfollowUser(this.client_user_id, followee_user_id)
  }
}
