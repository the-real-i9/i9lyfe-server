import { createNewPost, createReaction } from "../models/post_commentModel.js"

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
  post_owner_user_id,
  post_id,
  reaction_code_point,
}) => {
  await createReaction({
    user_id,
    post_owner_user_id,
    reacted_to: "post",
    reacted_to_id: post_id,
    reaction_code_point,
  })

  return {
    ok: true,
    err: null,
    data: null,
  }
}
