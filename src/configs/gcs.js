import { Storage } from "@google-cloud/storage"
import dotenv from "dotenv"

dotenv.config()

const bucket = new Storage({
  apiKey: process.env.GCS_API_KEY,
}).bucket(process.env.GCS_BUCKET)

export const getStorageBucket = () => {
  console.log(process.env.GCS_API_KEY)
  return bucket
}
