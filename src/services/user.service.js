import * as messageBrokerService from "../services/messageBroker.service.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as CRS from "../services/contentRecommendation.service.js"
import { User } from "../graph_models/user.model.js"

export const getSessionUser = async (client_user_id) => {
  const sessionUser = await User.findOne(client_user_id)

  return {
    data: sessionUser,
  }
}

export const followUser = async (client_user_id, to_follow_user_id) => {
  const { follow_notif } = await User.followUser(
    client_user_id,
    to_follow_user_id
  )

  if (follow_notif) {
    messageBrokerService.sendNewNotification(to_follow_user_id, follow_notif)
  }

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
    path_to_dest_folder: `profile_pictures/${client_username}`,
  })

  await User.changeProfilePicture(client_user_id, profile_pic_url)

  return {
    data: { msg: "operation successful" },
  }
}

export const getHomeFeedPosts = async ({ limit, offset, client_user_id }) => {
  const homeFeedPosts = await CRS.getHomePosts({
    limit,
    offset,
    client_user_id,
    types: ["photo", "video"],
  })

  return {
    data: homeFeedPosts,
  }
}

export const getHomeStoryPosts = async ({ limit, offset, client_user_id }) => {
  const homeStoryPosts = await CRS.getHomePosts({
    limit,
    offset,
    client_user_id,
    types: ["story"],
  })

  return {
    data: homeStoryPosts,
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

export const getFollowings = async ({
  username,
  limit,
  offset,
  client_user_id,
}) => {
  const userFollowing = await User.getFollowings({
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

export const getNotifications = async ({ client_user_id, limit, offset }) => {
  const notifications = await User.getNotifications({
    client_user_id,
    limit,
    offset,
  })

  return {
    data: notifications,
  }
}
