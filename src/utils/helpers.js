import jwt from "jsonwebtoken"
import bcrypt from "bcrypt"
import { Storage } from "@google-cloud/storage"

export const commaSeparateString = (str) => str.replaceAll(" ", ", ")

export const generateCodeWithExpiration = () => {
  const token = Math.trunc(Math.random() * 900000 + 100000)
  const expirationTime = new Date(Date.now() + 1 * 60 * 60 * 1000)

  return [token, expirationTime]
}

/**
 * @param {string|Buffer|JSON} payload
 * @returns {string} A JWT Token
 */
export const generateJwt = (payload) =>
  jwt.sign(payload, process.env.JWT_SECRET)

/**
 * 
 * @param {string} password 
 * @returns {Promise<string>} Hashed password
 */
export const hashPassword = async (password) => {
  return await bcrypt.hash(password, 10)
}

/** @param {Date} tokenExpiration */
export const tokenLives = (tokenExpiration) =>
  Date.now() < new Date(tokenExpiration)

/** @param {string} text */
export const extractMentions = (text) => {
  const matches = text.match(/(?<=@)\w+/g)
  return matches ? [...new Set(matches)] : []
}

/** @param {string} text */
export const extractHashtags = (text) => {
  const matches = text.match(/(?<=#)\w+/g)
  return matches ? [...new Set(matches)] : []
}

const bucketName = "i9lyfe-bucket"
const bucket = new Storage({
  credentials: {
    apiKey: process.env.GCS_API_KEY,
  }
}).bucket(bucketName)

export const getStorageBucket = () => {
  return bucket
}

export const getStorageBucketName = () => {
  return bucketName
}