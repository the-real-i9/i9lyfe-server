import { UserService } from "../services/UserService.js"

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const followUserController = async (req, res) => {
  try {
    // always get user_id from the jwtToken req.auth
    const { to_follow_user_id } = req.body

    const { client_user_id } = req.auth

    await new UserService(client_user_id).follow(to_follow_user_id)

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const unfollowUserController = async (req, res) => {
  try {
    const { followee_user_id } = req.body

    const { client_user_id } = req.auth

    await new UserService(client_user_id).unfollow(followee_user_id)

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const updateUserProfileController = async (req, res) => {
  try {
    const updatedUserInfoKVPairs = req.body

    const { client_user_id } = req.auth

    const updatedUserData = await new UserService(client_user_id).updateProfile(
      updatedUserInfoKVPairs
    )

    res.status(200).send({ updatedUserData })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const uploadProfilePictureController = async (req, res) => {
  try {
    // upload binary data to CDN, and store the url in profile_pic_url for the session use
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* GETs */

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserProfileController = async (req, res) => {
  try {
    const { username } = req.params

    const profileData = await new UserService(req.auth?.client_user_id).getProfile(username)

    res.status(200).send({ profileData })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserFollowersController = async (req, res) => {
  try {
    const { username } = req.params

    const userFollowers = await new UserService(req.auth?.client_user_id).getFollowers(username)

    res.status(200).send({ userFollowers })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserFollowingController = async (req, res) => {
  try {
    const { username } = req.params

    const userFollowing = await new UserService(req.auth?.client_user_id).getFollowing(username)

    res.status(200).send({ userFollowing })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserPostsController = async (req, res) => {
  try {
    const { username } = req.params

    const userPosts = await new UserService(req.auth?.client_user_id).getPosts(username)

    res.status(200).send({ userPosts })
  } catch (error) {
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserMentionedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const mentionedPosts = await new UserService(client_user_id).getMentionedPosts()

    res.status(200).send({ mentionedPosts })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserReactedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const reactedPosts = await new UserService(client_user_id).getReactedPosts()

    res.status(200).send({ reactedPosts })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserSavedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const savedPosts = await new UserService(client_user_id).getSavedPosts()

    res.status(200).send({ savedPosts })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserNotificationsController = async (req, res) => {
  try {
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
