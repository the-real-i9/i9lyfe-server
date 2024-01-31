import { dbQuery } from "./db.js"

export const getAllPosts = async (client_user_id) => {
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $1 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $1 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $1 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    `,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const searchAndFilterPosts = async ({
  search,
  type,
  client_user_id,
}) => {
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $3 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $3 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $3 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    WHERE (to_tsvector(description) @@ to_tsquery($1) AND type = $2) OR to_tsvector(description) @@ to_tsquery($1)`,
    values: [search, type, client_user_id],
  }

  return (await dbQuery(query)).rows
}

export const searchHashtags = async (search) => {
  const query = {
    text: `
    SELECT hashtag_name, COUNT(post_id) AS posts_count 
    FROM "PostCommentHashtag"
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
    FROM "User"
    WHERE username LIKE $1 OR name LIKE $1`,
    values: [`%${search}%`],
  }

  return (await dbQuery(query)).rows
}

export const getHashtagPosts = async (hashtag_name, client_user_id) => {
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      apv.post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $2 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $2 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $2 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView" apv
    INNER JOIN "PostCommentHashtag" pch USING post_id
    WHERE pch.hashtag_name = $1
    `,
    values: [hashtag_name, client_user_id],
  }

  return (await dbQuery(query)).rows
}
