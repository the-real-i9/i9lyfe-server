import { App } from "../graph_models/app.model.js"
import * as CRS from "../services/contentRecommendation.service.js"

export const getExploreFeed = async ({ limit, offset, client_username }) => {
  const exploreFeedPosts = await CRS.getExplorePosts({
    limit,
    offset,
    client_username,
    types: ["photo", "video"],
  })

  return {
    data: exploreFeedPosts,
  }
}

export const getExploreReels = async ({ limit, offset, client_username }) => {
  const exploreReelPosts = await CRS.getExplorePosts({
    limit,
    offset,
    client_username,
    types: ["reel"],
  })

  return {
    data: exploreReelPosts,
  }
}

export const searchAndFilter = async ({
  client_username,
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
          client_username,
        })

  return {
    data: results,
  }
}

export const getHashtagPosts = async ({
  hashtag_name,
  filter,
  limit,
  offset,
  client_username,
}) => {
  const hashtagPosts = await App.getHashtagPosts({
    hashtag_name,
    filter,
    limit,
    offset,
    client_username,
  })

  return {
    data: hashtagPosts,
  }
}
