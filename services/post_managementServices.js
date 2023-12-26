import { createNewPost } from "../models/postModel"



/**
 * @param {object} post
 * @param {string} post.user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const postCreationService = async (post) => {
  try {
    const result = await createNewPost(post)

    const postData = result.rows[0]

    return {
      ok: true,
      err: null,
      data: postData,
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: { code: 500, reason: "Internal server error" },
      data: null,
    }
  }
}
