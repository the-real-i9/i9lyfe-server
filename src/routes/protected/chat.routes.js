import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/chat.controllers.js"
import * as chatValidators from "../../middlewares/validators/chat.validators.js"
import { uploadMessageMedia } from "../../middlewares/mediaUploaders.js"
import { validateIdParams } from "../../middlewares/validators/miscs.js"

dotenv.config()

const router = express.Router()

router.use(
  expressjwt({
    secret: process.env.JWT_SECRET,
    algorithms: ["HS256"],
  }),
  (err, req, res, next) => {
    if (err) {
      res.status(err.status).send({ msg: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.post(
  "/create_conversation",
  ...chatValidators.createConversation,
  CC.createConversation
)

router.get("/my_conversations", CC.getMyConversations)

router.delete(
  "/conversations/:conversation_id",
  ...validateIdParams,
  CC.deleteConversation
)

router.get(
  "/conversations/:conversation_id/history",
  ...validateIdParams,
  CC.getConversationHistory
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/send_message",
  ...validateIdParams,
  ...chatValidators.sendMessage,
  uploadMessageMedia,
  CC.sendMessage
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/delivered",
  ...validateIdParams,
  ...chatValidators.ackMessageDelivered,
  CC.ackMessageDelivered
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/read",
  ...validateIdParams,
  CC.ackMessageRead
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/react",
  ...validateIdParams,
  ...chatValidators.reactToMessage,
  CC.reactToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/remove_reaction",
  ...validateIdParams,
  CC.removeReactionToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id",
  ...validateIdParams,
  ...chatValidators.deleteMessage,
  CC.deleteMessage
)

export default router
