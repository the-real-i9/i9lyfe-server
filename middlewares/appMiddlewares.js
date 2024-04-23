import fs from "fs/promises"
import { Buffer } from "buffer"
import path from "path"

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
    const media_urls = await Promise.all(req.body.media_blobs.map(async (dataArray, i) => {
      const now = Date.now()
      const filePath = path.resolve("static", "post_files", `${req.body.type}${i}-${now}.jpg`)
      const fileUrl = `http://localhost:5000/post_files/${req.body.type}${i}-${now}.jpg`

      await fs.writeFile(filePath, Buffer.from(dataArray))

      return fileUrl
    }))
    
    req.body.media_urls = media_urls
    
    delete req.body.media_blobs

    return next()
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
