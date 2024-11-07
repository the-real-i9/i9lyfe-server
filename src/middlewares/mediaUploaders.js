import fs from "node:fs"
import os from "node:os"
import { Buffer } from "node:buffer"
import { fileTypeFromBuffer } from "file-type"
import { Storage } from "@google-cloud/storage"

const bucketName = "i9lyfe-bucket"
const bucket = new Storage({
  credentials: {
    apiKey: process.env.GCS_API_KEY,
  }
}).bucket(bucketName)

export const uploadMessageMedia = (req, res, next) => {
  const fileData = req.body.msg_content.props.media_data


  req.body.msg_content.props.media_url = ""
  // change "media_data" property to "media_url"

  delete req.body.msg_content.props.media_data

  return next()
}

export const uploadPostMediaDatas = async (req, res, next) => {
  try {
    // "https://storage.googleapis.com/i9lyfe-bucket/%s"

    const fileDataList = req.body.media_data_list
    
    req.body.media_urls = []
    
    delete req.body.media_data_list

    return next()
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const uploadCommentAttachment = async (req, res, next) => {
  try {
    // "https://storage.googleapis.com/i9lyfe-bucket/%s"
    
    const fileData = req.body.attachment_data

    if (!fileData) {
      req.body.attachment_url = ""
      return next()
    }

    // write uint8 array to a file
    
      
    req.body.attachment_url = ""
    
    delete req.body.attachment_data

    return next()
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const uploadProfilePicture = async (req, res, next) => {
  try {
    const fileData = new Uint8Array(Buffer.from(req.body.picture_data))

    const fileType = await fileTypeFromBuffer(fileData)

    const destination = `profile_pictures/${req.auth.client_username}/profile_pic_${Date.now()}.${fileType.ext}`

    fs.writeFile(os.tmpdir + `tempfile.${fileType.ext}`, fileData, (err) => {
      bucket.upload(os.tmpdir + `tempfile.${fileType.ext}`, {
        destination
      })
    })
    
    req.body.profile_pic_url = `https://storage.googleapis.com/${bucketName}/${destination}`
    
    delete req.body.picture_data

    return next()
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}