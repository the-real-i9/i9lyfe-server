import * as userService from "../services/user.service.js"

export const getSessionUser = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const resp = await userService.getSessionUser(client_user_id)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const followUser = async (req, res) => {
  try {
    const { user_id: to_follow_user_id } = req.params

    const { client_user_id } = req.auth

    const resp = await userService.followUser(client_user_id, to_follow_user_id)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const unfollowUser = async (req, res) => {
  try {
    const { user_id } = req.params

    const { client_user_id } = req.auth

    const resp = await userService.unfollowUser(client_user_id, user_id)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const editProfile = async (req, res) => {
  try {
    const updateKVs = req.body

    const { client_user_id } = req.auth

    

    const resp = await userService.editProfile(client_user_id, updateKVs)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const readNotification = async (req, res) => {
  try {
    const { notification_id } = req.params

    const { client_user_id } = req.auth

    const resp = await userService.readNotification(
      notification_id,
      client_user_id
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const changeProfilePicture = async (req, res) => {
  try {
    const { picture_data } = req.body

    const resp = await userService.changeProfilePicture({
      picture_data,
      ...req.auth,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* GETs */

export const getHomeFeedPosts = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const resp = await userService.getHomeFeedPosts({
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getProfile = async (req, res) => {
  try {
    const { username } = req.params

    const resp = await userService.getProfile(
      username,
      req.auth?.client_user_idd
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getFollowers = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 50, offset = 0 } = req.query

    const resp = await userService.getFollowers({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getFollowing = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 50, offset = 0 } = req.query

    const resp = await userService.getFollowing({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getPosts = async (req, res) => {
  try {
    const { username } = req.params

    const { limit = 20, offset = 0 } = req.query

    const resp = await userService.getPosts({
      username,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getMentionedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const resp = await userService.getMentionedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const resp = await userService.getReactedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getSavedPosts = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const { limit = 20, offset = 0 } = req.query

    const resp = await userService.getSavedPosts({
      limit,
      offset,
      client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getNotifications = async (req, res) => {
  try {
    const { from, limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const resp = await userService.getNotifications({
      client_user_id,
      from,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
