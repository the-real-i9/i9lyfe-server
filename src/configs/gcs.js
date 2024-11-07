import { Storage } from "@google-cloud/storage"

export const storageBucketName = "i9lyfe-bucket"

const bucket = new Storage({
  credentials: {
    apiKey: process.env.GCS_API_KEY,
  }
}).bucket(storageBucketName)

export const getStorageBucket = () => {
  return bucket
}
