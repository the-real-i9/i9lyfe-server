import os from "node:os"
import fs from "node:fs/promises"
import { Buffer } from "node:buffer"
import { fileTypeFromBuffer } from "file-type"
import { randomBytes, randomUUID } from "node:crypto"
import { getStorageBucket } from "../configs/gcs.js"

/**
 * @param {object} param0
 * @param {number[]} param0.media_data
 * @param {string?} param0.extension
 * @param {string} param0.path_to_dest_folder
 * @returns
 */
export const upload = async ({
  media_data,
  extension,
  path_to_dest_folder,
}) => {
  const fileData = new Uint8Array(Buffer.from(media_data))

  const ext = extension || (await fileTypeFromBuffer(fileData)).ext

  const destination = `${path_to_dest_folder}/_${randomUUID()}_.${ext}`

  const tmpFile = `/i9lyfe_tempfile_${randomBytes(6).toString()}_.${ext}`

  fs.writeFile(os.tmpdir + tmpFile, fileData)
    .then(() => {
      getStorageBucket()
        .upload(os.tmpdir + tmpFile, {
          destination,
        })
        .catch((err) => console.error(err))
        .finally(() => {
          fs.rm(os.tmpdir + tmpFile).catch((err) => console.error(err))
        })
    })
    .catch((err) => console.error(err))

  return `https://storage.googleapis.com/${process.env.GCS_BUCKET}/${destination}`
}
