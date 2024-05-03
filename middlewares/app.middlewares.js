/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 * @typedef {import("express").NextFunction} ExpressNextFunction
 */

/**
 * Check `req.body` for fields that should contain binary data, based on the `type` property.
 *
 * Upload the binary data to cloud storage and replace the field's value with the returned `URL`
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 * @param {ExpressNextFunction} next
 */
export const uploadMessageFiles = (req, res, next) => {
  return next()
}

/**
 * Check `req.body` for fields that should contain binary data, based on the `type` property.
 *
 * Upload the binary data to cloud storage and replace the field's value with the returned `URL`
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 * @param {ExpressNextFunction} next
 */
export const uploadPostFiles = async (req, res, next) => {
  try {
    // "https://storage.cloud.google.com/i9lyfe-bucket/%s"
    
    req.body.media_urls = []
    
    delete req.body.media_blobs

    return next()
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const uploadCommentFiles = async (req, res, next) => {
  try {
    // "https://storage.cloud.google.com/i9lyfe-bucket/%s"
    
    
    req.body.attachment_url = ""
    
    delete req.body.attachment_blob

    return next()
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}