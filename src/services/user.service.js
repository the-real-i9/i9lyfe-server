import * as messageBrokerService from "../services/messageBroker.service.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as CRS from "../services/contentRecommendation.service.js"
import { User } from "../models/user.model.js"

export const getSessionUser = async (client_username) => {
  const sessionUser = await User.findOne(client_username)

  delete sessionUser.password

  return {
    data: sessionUser,
  }
}

export const followUser = async (client_username, to_follow_username) => {
  const { follow_notif } = await User.followUser(
    client_username,
    to_follow_username
  )

  if (follow_notif) {
    messageBrokerService.sendNewNotification(to_follow_username, follow_notif)
  }

  return {
    data: { msg: "operation successful" },
  }
}

export const unfollowUser = async (client_username, followee_username) => {
  await User.unfollowUser(client_username, followee_username)

  return {
    data: { msg: "operation successful" },
  }
}

export const editProfile = async (client_username, updateKVPairs) => {
  await User.edit(client_username, updateKVPairs)

  return {
    data: { msg: "operation successful" },
  }
}

export const updateConnectionStatus = async ({
  client_username,
  connection_status,
  last_active,
}) => {
  await User.updateConnectionStatus({
    client_username,
    connection_status,
    last_active,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const readNotification = async (notification_id, client_username) => {
  await User.readNotification(notification_id, client_username)

  return {
    data: { msg: "operation successful" },
  }
}

export const changeProfilePicture = async ({
  client_username,
  picture_data,
}) => {
  const profile_pic_url = await mediaUploadService.upload({
    media_dat: picture_data,
    extension: null,
    path_to_dest_folder: `profile_pictures/${client_username}`,
  })

  await User.changeProfilePicture(client_username, profile_pic_url)

  return {
    data: { msg: "operation successful" },
  }
}

export const getHomeFeedPosts = async ({ limit, offset, client_username }) => {
  const homeFeedPosts = await CRS.getHomePosts({
    limit,
    offset,
    client_username,
    types: ["photo", "video"],
  })

  return {
    data: homeFeedPosts,
  }
}

export const getProfile = async (username, client_username) => {
  const profileData = await User.getProfile(username, client_username)

  return {
    data: profileData,
  }
}

export const getFollowers = async ({
  username,
  limit,
  offset,
  client_username,
}) => {
  const userFollowers = await User.getFollowers({
    username,
    limit,
    offset,
    client_username,
  })

  return {
    data: userFollowers,
  }
}

export const getFollowings = async ({
  username,
  limit,
  offset,
  client_username,
}) => {
  const userFollowing = await User.getFollowings({
    username,
    limit,
    offset,
    client_username,
  })

  return {
    data: userFollowing,
  }
}

export const getPosts = async ({ username, limit, offset, client_username }) => {
  const userPosts = await User.getPosts({
    username,
    limit,
    offset,
    client_username,
  })

  return {
    data: userPosts,
  }
}

export const getMentionedPosts = async ({ limit, offset, client_username }) => {
  const mentionedPosts = await User.getMentionedPosts({
    limit,
    offset,
    client_username,
  })

  return {
    data: mentionedPosts,
  }
}

export const getReactedPosts = async ({ limit, offset, client_username }) => {
  const reactedPosts = await User.getReactedPosts({
    limit,
    offset,
    client_username,
  })

  return {
    data: reactedPosts,
  }
}

export const getSavedPosts = async ({ limit, offset, client_username }) => {
  const savedPosts = await User.getSavedPosts({
    limit,
    offset,
    client_username,
  })

  return {
    data: savedPosts,
  }
}

export const getNotifications = async ({ client_username, limit, offset }) => {
  const notifications = await User.getNotifications({
    client_username,
    limit,
    offset,
  })

  return {
    data: notifications,
  }
}
