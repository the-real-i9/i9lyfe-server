import { createNewPost } from "../models/postModel.js"
import { createReaction } from "../utils/post_comment_dbTasks.js"

/**
 * @param {object} post
 * @param {string} post.user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const postCreationService = async (post) => {
  const result = await createNewPost(post)

  const postData = result.rows[0]

  return {
    ok: true,
    err: null,
    data: postData,
  }
}

export const postReactionService = async ({
  user_id,
  post_id,
  reaction_code_point,
}) => {
  await createReaction({
    user_id,
    reaction_receiver: "post",
    reaction_receiver_id: post_id,
    reaction_code_point,
  })

  return {
    ok: true,
    err: null,
    data: null,
  }
}
