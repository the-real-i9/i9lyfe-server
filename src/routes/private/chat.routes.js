import express from "express"

import * as CC from "../../controllers/chat.controllers.js"
import { validateParams } from "../../validators/miscs.js"

const router = express.Router()

router.get("/my_chats", CC.getMyChats)

router.delete(
  "/chats/:partner_username",
  ...validateParams,
  CC.deleteChat
)


export default router
