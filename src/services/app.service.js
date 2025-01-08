import { App } from "../graph_models/app.model.js"
import * as CRS from "../services/contentRecommendation.service.js"

export const searchUsersToChat = async ({
  term,
  limit,
  offset,
  client_user_id,
}) => {
  const users = await App.searchUsersToChat({
    term,
    limit,
    offset,
    client_user_id,
  })

  return {
    data: users,
  }
}

export const getExploreFeed = async ({ limit, offset, client_user_id }) => {
  const exploreFeedPosts = await CRS.getExplorePosts({
    limit,
    offset,
    client_user_id,
    types: ["photo", "video"],
  })

  return {
    data: exploreFeedPosts,
  }
}

export const getExploreReels = async ({ limit, offset, client_user_id }) => {
  const exploreReelPosts = await CRS.getExplorePosts({
    limit,
    offset,
    client_user_id,
    types: ["reel"],
  })

  return {
    data: exploreReelPosts,
  }
}

export const searchAndFilter = async ({
  client_user_id,
  term,
  filter,
  limit,
  offset,
}) => {
  const results =
    filter === "hashtag"
      ? await App.searchHashtags({ term, limit, offset })
      : filter === "user"
      ? await App.searchUsers({ term, limit, offset })
      : await App.searchAndFilterPosts({
          term,
          filter,
          limit,
          offset,
          client_user_id,
        })

  return {
    data: results,
  }
}

export const getHashtagPosts = async ({
  hashtag_name,
  limit,
  offset,
  client_user_id,
}) => {
  const hashtagPosts = await App.getHashtagPosts({
    hashtag_name,
    limit,
    offset,
    client_user_id,
  })

  return {
    data: hashtagPosts,
  }
}
