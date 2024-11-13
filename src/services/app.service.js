import { App } from "../models/app.model.js"

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

export const getExplorePosts = async ({ limit, offset, client_user_id }) => {
  const explorePosts = await App.getExplorePosts({
    limit,
    offset,
    client_user_id,
  })

  return {
    data: explorePosts,
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
