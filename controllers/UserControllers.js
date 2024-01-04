import { UserService } from "../services/UserService.js"

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const followUserController = async (req, res) => {
  try {
    // always get user_id from the jwtToken req.auth
    const { to_follow_user_id } = req.body

    const { user_id } = req.auth

    await new UserService(user_id).follow(to_follow_user_id)

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

    const { user_id } = req.auth

    await new UserService(user_id).unfollow(followee_user_id)

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

    const { user_id } = req.auth

    const updatedUserData = await new UserService(user_id).updateProfile(
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

    const profileData = await new UserService().getProfile({
      username,
      client_user_id: req.auth?.user_id,
    })

    res.status(200).send({ profileData })
  } catch (error) {
    console.log(error)
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

    const userFollowers = await new UserService().getFollowers({
      username,
      client_user_id: req.auth?.user_id,
    })

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

    const userFollowing = await new UserService().getFollowing({
      username,
      client_user_id: req.auth?.user_id,
    })

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

    const userPosts = await new UserService().getPosts({
      username,
      client_user_id: req.auth?.user_id,
    })

    res.status(200).send({ userPosts })
  } catch (error) {
    res.sendStatus(500)
  }
}

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const getUserMentionsController = async (req, res) => {
  try {
    
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
