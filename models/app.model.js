import { dbQuery } from "./db.js"

export const getExplorePosts = async (client_user_id) => {
  const query = {
    text: "SELECT * FROM get_explore_posts($1)",
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const searchAndFilterPosts = async ({
  search,
  filter,
  limit,
  offset,
  client_user_id,
}) => {
  const query = {
    text: "SELECT * FROM search_filter_posts($1, $2, $3, $4, $5)",
    values: [search, filter, limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const searchHashtags = async (search) => {
  const query = {
    text: `
    SELECT hashtag_name, COUNT(post_id) AS posts_count 
    FROM pc_hashtag
    WHERE hashtag_name ILIKE $1
    GROUP BY hashtag_name`,
    values: [`%${search}%`],
  }

  return (await dbQuery(query)).rows
}

export const searchUsers = async (search) => {
  const query = {
    text: `
    SELECT id AS user_id, 
      username, 
      name, 
      profile_pic_url
    FROM i9l_user
    WHERE username ILIKE $1 OR name ILIKE $1`,
    values: [`%${search}%`],
  }

  return (await dbQuery(query)).rows
}

export const getHashtagPosts = async ({
  hashtag_name,
  limit,
  offset,
  client_user_id,
}) => {
  const query = {
    text: "SELECT * FROM get_hashtag_posts($1, $2, $3, $4)",
    values: [hashtag_name, limit, offset, client_user_id],
  }

  return (await dbQuery(query)).rows
}
