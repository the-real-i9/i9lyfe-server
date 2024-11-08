import os from "node:os"
import fs from "node:fs"
import { Buffer } from "node:buffer"
import { fileTypeFromBuffer } from "file-type"
import { getStorageBucket, storageBucketName } from "../configs/gcs.js"
import { randomUUID } from "node:crypto"

/**
 * @param {number[][]} media_data_list
 */
export const uploadPostMediaDataList = async (media_data_list) => {
  const media_urls = media_data_list.map(async (media_data) => {
    const fileData = new Uint8Array(Buffer.from(media_data))

    const fileType = await fileTypeFromBuffer(fileData)

    const destination = `post_medias/_${randomUUID()}_.${fileType.ext}`

    fs.writeFile(os.tmpdir + `/tempfile.${fileType.ext}`, fileData, () => {
      getStorageBucket().upload(os.tmpdir + `/tempfile.${fileType.ext}`, {
        destination,
      })
    })

    return `https://storage.googleapis.com/${storageBucketName}/${destination}`
  })

  return media_urls
}

/**
 * @param {number[]} attachment_data
 */
export const uploadCommentAttachmentData = async (attachment_data) => {
  const fileData = new Uint8Array(Buffer.from(attachment_data))

  const fileType = await fileTypeFromBuffer(fileData)

  const destination = `comment_attachments/_${Date.now()}_.${fileType.ext}`

  fs.writeFile(os.tmpdir + `/tempfile.${fileType.ext}`, fileData, () => {
    getStorageBucket().upload(os.tmpdir + `/tempfile.${fileType.ext}`, {
      destination,
    })
  })

  return `https://storage.googleapis.com/${storageBucketName}/${destination}`
}

/**
 * @param {number[]} media_data
 */
export const uploadMessageMediaData = async (media_data, extension) => {
  const fileData = new Uint8Array(Buffer.from(media_data))

  const fileType = await fileTypeFromBuffer(fileData)

  const ext = extension || fileType.ext

  const destination = `message_medias/_${randomUUID()}_.${ext}`

  fs.writeFile(os.tmpdir + `/tempfile.${ext}`, fileData, () => {
    getStorageBucket().upload(os.tmpdir + `/tempfile.${ext}`, {
      destination,
    })
  })

  return `https://storage.googleapis.com/${storageBucketName}/${destination}`
}

/**
 * @param {number[]} picture_data
 * @param {string} username
 */
export const uploadProfilePicture = async (picture_data, username) => {
  const fileData = new Uint8Array(Buffer.from(picture_data))

  const fileType = await fileTypeFromBuffer(fileData)

  const destination = `profile_pictures/${username}/profile_pic_${randomUUID()}.${
    fileType.ext
  }`

  fs.writeFile(os.tmpdir + `/tempfile.${fileType.ext}`, fileData, () => {
    getStorageBucket().upload(os.tmpdir + `/tempfile.${fileType.ext}`, {
      destination,
    })
  })

  return `https://storage.googleapis.com/${storageBucketName}/${destination}`
}
