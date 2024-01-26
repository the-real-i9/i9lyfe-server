import { dbQuery } from "./db.js"

export const getAllPosts = async (client_user_id) => {
  const query = {
    text: `
    SELECT json_build_object(
        'id', owner_user_id,
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