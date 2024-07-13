import { UserService } from "../services/user.service.js"

export const getSessionUser = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const sessionUser = await UserService.getClientUser(client_user_id)

    res.status(200).send({ sessionUser })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const followUser = async (req, res) => {
  try {
    const { user_id: to_follow_user_id } = req.params

    const { client_user_id } = req.auth

    await UserService.follow(client_user_id, to_follow_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const unfollowUser = async (req, res) => {
  try {
    const { user_id: followee_user_id } = req.params

    const { client_user_id } = req.auth

    await UserService.unfollow(client_user_id, followee_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const editProfile = async (req, res) => {
  try {
    const updateDict = req.body

    const { client_user_id } = req.auth

    await UserService.editProfile(client_user_id, Object.entries(updateDict))

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const updateConnectionStatus = async (req, res) => {
  try {
    const { connection_status, last_active = null } = req.body

    const { client_user_id } = req.auth

    await UserService.updateConnectionStatus({
      client_user_id,
      connection_status,
      last_active,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const readNotification = async (req, res) => {
  try {
    const { notification_id } = req.params

    const { client_user_id } = req.auth

    await UserService.readNotification(notification_id, client_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const uploadProfilePicture = async (req, res) => {
  try {
    // upload binary data to CDN, and store the url in profile_pic_url for the session use
    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* GETs */

export const getHomeFeed = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const homeFeedPosts = await UserService.getFeedPosts({
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(homeFeedPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getProfile = async (req, res) => {
  try {
    const { username } = req.params

    const profileData = await UserService.getProfile(
      username,
      req.auth?.client_user_id
    )

    res.status(200).send(profileData)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getFollowers = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 50, offset = 0 } = req.query

    const userFollowers = await UserService.getFollowers({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(userFollowers)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getFollowing = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 50, offset = 0 } = req.query

    const userFollowing = await UserService.getFollowing({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(userFollowing)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getPosts = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 20, offset = 0 } = req.query

    const userPosts = await UserService.getPosts({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(userPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getMentionedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const mentionedPosts = await UserService.getMentionedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(mentionedPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const reactedPosts = await UserService.getReactedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(reactedPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getSavedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const savedPosts = await UserService.getSavedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(savedPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getNotifications = async (req, res) => {
  try {
    const { from, limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const notifications = await UserService.getNotifications({
      client_user_id,
      from,
      limit,
      offset,
    })

    res.status(200).send(notifications)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
