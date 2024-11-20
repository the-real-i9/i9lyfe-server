import { Storage } from "@google-cloud/storage"


const bucket = new Storage({
  apiKey: process.env.GCS_API_KEY
}).bucket(process.env.GCS_BUCKET)

export const getStorageBucket = () => {
  return bucket
}
