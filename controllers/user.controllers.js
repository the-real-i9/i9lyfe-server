import { UserService } from "../services/user.service.js"

export const getSessionUserController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const clientUser = await UserService.getClientUser(client_user_id)

    res.status(200).send({ clientUser })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}


export const followUserController = async (req, res) => {
  try {
    const { user_id: to_follow_user_id } = req.params

    const { client_user_id } = req.auth

    await UserService.follow(client_user_id, to_follow_user_id)

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const unfollowUserController = async (req, res) => {
  try {
    const { user_id: followee_user_id } = req.params

    const { client_user_id } = req.auth

    await UserService.unfollow(client_user_id, followee_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const editProfileController = async (req, res) => {
  try {
    const updateDict = req.body

    const { client_user_id } = req.auth

    await UserService.editProfile(client_user_id, Object.entries(updateDict))

    res.status(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const updateUserConnectionStatusController = async (req, res) => {
  try {
    const { connection_status, last_active = null } = req.body

    const { client_user_id } = req.auth

    await UserService.updateConnectionStatus({
      client_user_id,
      connection_status,
      last_active,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const readUserNotificationController = async (req, res) => {
  try {
    const { notification_id } = req.body

    const { client_user_id } = req.auth

    await UserService.readNotification(notification_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const uploadProfilePictureController = async (req, res) => {
  try {
    // upload binary data to CDN, and store the url in profile_pic_url for the session use
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/* GETs */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getHomeFeedController = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const homeFeedPosts = await UserService.getFeedPosts({
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send({ homeFeedPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserProfileController = async (req, res) => {
  try {
    const { username } = req.params

    const profileData = await UserService.getProfile(
      username,
      req.auth?.client_user_id
    )

    res.status(200).send({ profileData })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserFollowersController = async (req, res) => {
  try {
    const { username } = req.params

    const userFollowers = await UserService.getFollowers(
      username,
      req.auth?.client_user_id
    )

    res.status(200).send({ userFollowers })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserFollowingController = async (req, res) => {
  try {
    const { username } = req.params

    const userFollowing = await UserService.getFollowing(
      username,
      req.auth?.client_user_id
    )

    res.status(200).send({ userFollowing })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserPostsController = async (req, res) => {
  try {
    const { username } = req.params

    const userPosts = await UserService.getPosts(
      username,
      req.auth?.client_user_id
    )

    res.status(200).send({ userPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserMentionedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const mentionedPosts = await UserService.getMentionedPosts(client_user_id)

    res.status(200).send({ mentionedPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserReactedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const reactedPosts = await UserService.getReactedPosts(client_user_id)

    res.status(200).send({ reactedPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserSavedPostsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const savedPosts = await UserService.getSavedPosts(client_user_id)

    res.status(200).send({ savedPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getUserNotificationsController = async (req, res) => {
  try {
    const { from } = req.query

    const { client_user_id } = req.auth

    const notifications = await UserService.getUserNotifications(
      client_user_id,
      from
    )

    res.status(200).send({ notifications })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
