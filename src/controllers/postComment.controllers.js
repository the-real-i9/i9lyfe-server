import * as postCommentService from "../services/postComment.service.js"

export const createNewPost = async (req, res) => {
  try {
    const { media_data_list, type, description = "" } = req.body

    const { client_username } = req.auth

    const resp = await postCommentService.createNewPost({
      client_username,
      media_data_list,
      type,
      description,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToPost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { reaction } = req.body

    const { client_username } = req.auth

    const resp = await postCommentService.reactToPost({
      client_username,
      post_id,
      reaction,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const commentOnPost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { comment_text, attachment_data } = req.body

    const { client_username } = req.auth

    const resp = await postCommentService.commentOnPost({
      client_username,
      post_id,
      comment_text,
      attachment_data,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToComment = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { reaction } = req.body

    const { client_username } = req.auth

    const resp = await postCommentService.reactToComment({
      client_username,
      comment_id,
      reaction,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const commentOnComment = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { comment_text, attachment_data } = req.body

    const { client_username } = req.auth

    const resp = await postCommentService.commentOnComment({
      client_username,
      comment_id,
      comment_text,
      attachment_data,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const createRepost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.createRepost(post_id, client_username)

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const savePost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_username } = req.auth

    const resp = await postCommentService.savePost(post_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const unsavePost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.unsavePost(post_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* The GETs */

export const getPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_username } = req.auth

    const resp = await postCommentService.getPost(post_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getCommentsOnPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getCommentsOnPost({
      post_id,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_username } = req.auth

    const resp = await postCommentService.getComment(comment_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsToPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getReactorsToPost({
      post_id,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsWithReactionToPost = async (req, res) => {
  try {
    const { post_id, reaction } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getReactorsWithReactionToPost({
      post_id,
      reaction,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getCommentsOnComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getCommentsOnComment({
      comment_id,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsToComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getReactorsToComment({
      comment_id,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsWithReactionToComment = async (req, res) => {
  try {
    const { comment_id, reaction } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_username } = req.auth

    const resp = await postCommentService.getReactorsWithReactionToComment({
      comment_id,
      reaction,
      client_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* DELETEs */

export const deletePost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.deletePost(post_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToPost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.removeReactionToPost(
      post_id,
      client_username
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeCommentOnPost = async (req, res) => {
  try {
    const { post_id, comment_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.removeCommentOnPost({
      post_id,
      comment_id,
      client_username,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeCommentOnComment = async (req, res) => {
  try {
    const { parent_comment_id, comment_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.removeCommentOnComment({
      parent_comment_id,
      comment_id,
      client_username,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToComment = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.removeReactionToComment(
      comment_id,
      client_username
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteRepost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_username } = req.auth

    const resp = await postCommentService.deletePost(post_id, client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
