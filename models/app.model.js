import { dbQuery } from "./db.js"

export const getAllPosts = async (client_user_id) => {
  const query = {
    text: "SELECT all_posts FROM get_all_post($1)",
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows[0].all_posts
}

export const searchAndFilterPosts = async ({
  search,
  filter,
  client_user_id,
}) => {
  const query = {
    text: "SELECT res_posts FROM search_filter_posts($1, $2, $3)",
    values: [search, filter, client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const searchHashtags = async (search) => {
  const query = {
    text: `
    SELECT hashtag_name, COUNT(post_id) AS posts_count 
    FROM pc_hashtag
    WHERE hashtag_name LIKE $1
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
    WHERE username LIKE $1 OR name LIKE $1`,
    values: [`%${search}%`],
  }

  return (await dbQuery(query)).rows
}

export const getHashtagPosts = async (hashtag_name, client_user_id) => {
  const query = {
    text: "SELECT hashtag_posts FROM get_hashtag_posts($1, $2)",
    values: [hashtag_name, client_user_id],
  }

  return (await dbQuery(query)).rows[0].hashtag_posts
}
