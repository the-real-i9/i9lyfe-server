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
export const uploadPostFiles = (req, res, next) => {
  return next()
}