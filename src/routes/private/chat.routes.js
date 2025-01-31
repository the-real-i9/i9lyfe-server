import express from "express"

import * as CC from "../../controllers/chat.controllers.js"
import * as chatValidators from "../../validators/chat.validators.js"
import { validateParams } from "../../validators/miscs.js"

const router = express.Router()

router.post(
  "/chats/:partner_username/send_message",
  ...validateParams,
  ...chatValidators.sendMessage,
  CC.sendMessage
)

router.get("/my_chats", CC.getMyChats)

router.delete(
  "/chats/:partner_username",
  ...validateParams,
  CC.deleteChat
)

router.get(
  "/chats/:partner_username/history",
  ...validateParams,
  CC.getChatHistory
)


router.put(
  "/chats/:partner_username/messages/:message_id/delivered",
  ...validateParams,
  ...chatValidators.ackMessageDelivered,
  CC.ackMessageDelivered
)

router.put(
  "/chats/:partner_username/messages/:message_id/read",
  ...validateParams,
  ...chatValidators.ackMessageRead,
  CC.ackMessageRead
)

router.post(
  "/chats/:partner_username/messages/:message_id/react",
  ...validateParams,
  ...chatValidators.reactToMessage,
  CC.reactToMessage
)

router.delete(
  "/chats/:partner_username/messages/:message_id/remove_reaction",
  ...validateParams,
  CC.removeReactionToMessage
)

router.delete(
  "/chats/:partner_username/messages/:message_id",
  ...validateParams,
  ...chatValidators.deleteMessage,
  CC.deleteMessage
)

export default router
