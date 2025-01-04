import express from "express"

import * as CC from "../../controllers/chat.controllers.js"
import * as chatValidators from "../../validators/chat.validators.js"
import { validateIdParams } from "../../validators/miscs.js"

const router = express.Router()

router.post(
  "/chats/:partner_user_id/send_message",
  ...validateIdParams,
  ...chatValidators.sendMessage,
  CC.sendMessage
)

router.get("/my_chats", CC.getMyChats)

router.delete(
  "/chats/:partner_user_id",
  ...validateIdParams,
  CC.deleteChat
)

router.get(
  "/chats/:partner_user_id/history",
  ...validateIdParams,
  CC.getChatHistory
)


router.put(
  "/chats/:partner_user_id/messages/:message_id/delivered",
  ...validateIdParams,
  ...chatValidators.ackMessageDelivered,
  CC.ackMessageDelivered
)

router.put(
  "/chats/:partner_user_id/messages/:message_id/read",
  ...validateIdParams,
  CC.ackMessageRead
)

router.post(
  "/chats/:partner_user_id/messages/:message_id/react",
  ...validateIdParams,
  ...chatValidators.reactToMessage,
  CC.reactToMessage
)

router.delete(
  "/chats/:partner_user_id/messages/:message_id/remove_reaction",
  ...validateIdParams,
  CC.removeReactionToMessage
)

router.delete(
  "/chats/:partner_user_id/messages/:message_id",
  ...validateIdParams,
  ...chatValidators.deleteMessage,
  CC.deleteMessage
)

export default router
