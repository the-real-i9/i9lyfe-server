import express from "express"

import * as CC from "../../controllers/chat.controllers.js"
import * as chatValidators from "../../validators/chat.validators.js"
import { validateIdParams } from "../../validators/miscs.js"

const router = express.Router()

router.post(
  "/new_chat",
  ...chatValidators.createChat,
  CC.createChat
)

router.get("/my_chats", CC.getMyChats)

router.delete(
  "/chats/:chat_id",
  ...validateIdParams,
  CC.deleteChat
)

router.get(
  "/chats/:chat_id/history",
  ...validateIdParams,
  CC.getChatHistory
)

router.post(
  "/chats/:chat_id/send_message",
  ...validateIdParams,
  ...chatValidators.sendMessage,
  CC.sendMessage
)

router.put(
  "/chats/:chat_id/messages/:message_id/delivered",
  ...validateIdParams,
  ...chatValidators.ackMessageDelivered,
  CC.ackMessageDelivered
)

router.put(
  "/chats/:chat_id/messages/:message_id/read",
  ...validateIdParams,
  CC.ackMessageRead
)

router.post(
  "/chats/:chat_id/messages/:message_id/react",
  ...validateIdParams,
  ...chatValidators.reactToMessage,
  CC.reactToMessage
)

router.delete(
  "/chats/:chat_id/messages/:message_id/remove_reaction",
  ...validateIdParams,
  CC.removeReactionToMessage
)

router.delete(
  "/chats/:chat_id/messages/:message_id",
  ...validateIdParams,
  ...chatValidators.deleteMessage,
  CC.deleteMessage
)

export default router
