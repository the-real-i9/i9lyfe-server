import fs from "node:fs"
import os from "node:os"
import { Buffer } from "node:buffer"
import { fileTypeFromBuffer } from "file-type"
import { Storage } from "@google-cloud/storage"

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