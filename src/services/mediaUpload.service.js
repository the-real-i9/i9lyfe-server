import os from "node:os"
import fs from "node:fs"
import { Buffer } from "node:buffer"
import { fileTypeFromBuffer } from "file-type"
import { randomUUID } from "node:crypto"
import { getStorageBucket, storageBucketName } from "../configs/gcs.js"

export const upload = async ({ media_data, extension, pathToDestFolder }) => {
  const fileData = new Uint8Array(Buffer.from(media_data))

  const ext = extension || (await fileTypeFromBuffer(fileData)).ext

  const destination = `${pathToDestFolder}/_${randomUUID()}_.${ext}`

  fs.writeFile(os.tmpdir + `/tempfile.${ext}`, fileData, () => {
    getStorageBucket().upload(os.tmpdir + `/tempfile.${ext}`, {
      destination,
    })
  })

  return `https://storage.googleapis.com/${storageBucketName}/${destination}`
}
