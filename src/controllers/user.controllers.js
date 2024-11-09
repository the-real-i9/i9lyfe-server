import { User } from "../models/user.model.js"
import { uploadProfilePicture } from "../services/mediaUploader.service.js"
import { sendNewNotification } from "../services/messageBroker.service.js"

export const getSessionUser = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const sessionUser = await User.findOne(client_user_id)

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

    const { follow_notif } = await User.followUser(
      client_user_id,
      to_follow_user_id
    )

    const { receiver_user_id, ...restData } = follow_notif
    sendNewNotification(receiver_user_id, restData)

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

    await User.unfollowUser(client_user_id, followee_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const editProfile = async (req, res) => {
  try {
    const updateKVPairs = req.body

    const { client_user_id } = req.auth

    await User.edit(client_user_id, updateKVPairs)

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

    await User.updateConnectionStatus({
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

    await User.readNotification(notification_id, client_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const changeProfilePicture = async (req, res) => {
  try {
    const { client_user_id, client_username } = req.auth

    const { picture_data } = req.body

    const profile_pic_url = await uploadProfilePicture(
      picture_data,
      client_username
    )

    await User.changeProfilePicture(client_user_id, profile_pic_url)

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

    const homeFeedPosts = await User.getFeedPosts({
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

    const profileData = await User.getProfile(
      username,
      req.auth?.client_user_idd
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

    const userFollowers = await User.getFollowers({
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

    const userFollowing = await User.getFollowing({
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

    const userPosts = await User.getPosts({
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

    const mentionedPosts = await User.getMentionedPosts({ limit, offset, client_user_id })

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

    const reactedPosts = await User.getReactedPosts({ limit, offset, client_user_id })

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

    const savedPosts = await User.getSavedPosts({ limit, offset, client_user_id })

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

    const notifications = await User.getNotifications({
      client_user_id,
      from: new Date(from),
      limit,
      offset,
    })

    res.status(200).send(notifications)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
