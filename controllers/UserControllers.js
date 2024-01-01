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

    const response = await new UserService(user_id).updateProfile(
      updatedUserInfoKVPairs
    )

    res.status(200).send({ updatedUserData: response.data })
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
