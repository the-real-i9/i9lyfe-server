import {
  followUser,
  unfollowUser,
  updateUserProfile,
} from "../models/UserModel.js"

export class UserService {
  constructor(user_id) {
    this.user_id = user_id
  }

  async follow(to_follow_user_id) {
    await followUser(this.user_id, to_follow_user_id)

    return {
      ok: true,
      err: null,
      data: null,
    }
  }

  async unfollow(followee_user_id) {
    await unfollowUser(this.user_id, followee_user_id)

    return {
      ok: true,
      err: null,
      data: null,
    }
  }

  async updateProfile(updatedUserInfoKVPairs) {
    await updateUserProfile(this.user_id, updatedUserInfoKVPairs)

    return {
      ok: true,
      err: null,
      data: null,
    }
  }

  async uploadProfilePicture() {
    // upload binary to CDN and get back file URL
  }
}
