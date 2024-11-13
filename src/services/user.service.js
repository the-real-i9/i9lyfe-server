import * as messageBrokerService from "../services/messageBroker.service.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import { User } from "../models/user.model.js"

export const getSessionUser = async (client_user_id) => {
  const sessionUser = await User.findOne(client_user_id)

  return {
    data: { sessionUser },
  }
}

export const followUser = async (client_user_id, to_follow_user_id) => {
  const { follow_notif } = await User.followUser(
    client_user_id,
    to_follow_user_id
  )

  const { receiver_user_id, ...restData } = follow_notif
  messageBrokerService.sendNewNotification(receiver_user_id, restData)

  return {
    data: { msg: "operation successful" },
  }
}

export const unfollowUser = async (client_user_id, followee_user_id) => {
  await User.unfollowUser(client_user_id, followee_user_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const editProfile = async (client_user_id, updateKVPairs) => {
  await User.edit(client_user_id, updateKVPairs)

  return {
    data: { msg: "operation successful" },
  }
}

export const updateConnectionStatus = async ({
  client_user_id,
  connection_status,
  last_active,
}) => {
  await User.updateConnectionStatus({
    client_user_id,
    connection_status,
    last_active,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const readNotification = async (notification_id, client_user_id) => {
  await User.readNotification(notification_id, client_user_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const changeProfilePicture = async ({
  client_user_id,
  client_username,
  picture_data,
}) => {
  const profile_pic_url = await mediaUploadService.upload({
    media_dat: picture_data,
    extension: null,
    pathToDestFolder: `profile_pictures/${client_username}`,
  })

  await User.changeProfilePicture(client_user_id, profile_pic_url)

  return {
    data: { msg: "operation successful" },
  }
}

/* GETs */

export const getHomeFeedPosts = async ({ client_user_id, limit, offset }) => {
  const homeFeedPosts = await User.getHomeFeedPosts({
    client_user_id,
    limit,
    offset,
  })

  return {
    data: homeFeedPosts,
  }
}

export const getProfile = async (username, client_user_id) => {
  const profileData = await User.getProfile(username, client_user_id)

  return {
    data: profileData,
  }
}

export const getFollowers = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  const userFollowers = await User.getFollowers({
    username,
    limit,
    offset,
    client_user_id,
  })

  return {
    data: userFollowers,
  }
}

export const getFollowing = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  const userFollowing = await User.getFollowing({
    username,
    limit,
    offset,
    client_user_id,
  })

  return {
    data: userFollowing,
  }
}

export const getPosts = async ({ username, limit, offset, client_user_id }) => {
  const userPosts = await User.getPosts({
    username,
    limit,
    offset,
    client_user_id,
  })

  return {
    data: userPosts,
  }
}

export const getMentionedPosts = async ({ limit, offset, client_user_id }) => {
  const mentionedPosts = await User.getMentionedPosts({
    limit,
    offset,
    client_user_id,
  })

  return {
    data: mentionedPosts,
  }
}

export const getReactedPosts = async ({ limit, offset, client_user_id }) => {
  const reactedPosts = await User.getReactedPosts({
    limit,
    offset,
    client_user_id,
  })

  return {
    data: reactedPosts,
  }
}

export const getSavedPosts = async ({ limit, offset, client_user_id }) => {
  const savedPosts = await User.getSavedPosts({
    limit,
    offset,
    client_user_id,
  })

  return {
    data: savedPosts,
  }
}

export const getNotifications = async ({
  client_user_id,
  from,
  limit,
  offset,
}) => {
  const notifications = await User.getNotifications({
    client_user_id,
    from: new Date(from),
    limit,
    offset,
  })

  return {
    data: notifications,
  }
}
