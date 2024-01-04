import {
  followUser,
  getUserFollowers,
  getUserFollowing,
  getUserPosts,
  getUserProfile,
  unfollowUser,
  updateUserProfile,
} from "../models/UserModel.js"

export class UserService {
  constructor(user_id) {
    this.user_id = user_id
  }

  async follow(to_follow_user_id) {
    await followUser(this.user_id, to_follow_user_id)
  }

  async unfollow(followee_user_id) {
    await unfollowUser(this.user_id, followee_user_id)
  }

  async updateProfile(updatedUserInfoKVPairs) {
    const result = await updateUserProfile(this.user_id, updatedUserInfoKVPairs)

    const updatedUserData = result.rows[0]

    return updatedUserData
  }

  async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }

  /* GETs */
  async getProfile({username, client_user_id}) {
    return await getUserProfile(username, client_user_id)
  }

  async getFollowers({username, client_user_id}) {
    return await getUserFollowers(username, client_user_id)
  }

  async getFollowing({username, client_user_id}) {
    return await getUserFollowing(username, client_user_id)
  }

  async getPosts({username, client_user_id}) {
    return await getUserPosts(username, client_user_id)
  }
}
