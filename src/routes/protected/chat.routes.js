import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/chat.controllers.js"
import * as chatValidators from "../../middlewares/validators/chat.validators.js"
import { uploadMessageFiles } from "../../middlewares/app.middlewares.js"

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

router.post("/create_conversation", CC.createConversation)

router.get("/my_conversations", CC.getMyConversations)

router.delete(
  "/conversations/:conversation_id",
  ...chatValidators.validateIdParams,
  CC.deleteConversation
)

router.get(
  "/conversations/:conversation_id/history",
  ...chatValidators.validateIdParams,
  CC.getConversationHistory
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/send_message",
  ...chatValidators.validateIdParams,
  uploadMessageFiles,
  CC.sendMessage
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/delivered",
  ...chatValidators.validateIdParams,
  CC.ackMessageDelivered
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/read",
  ...chatValidators.validateIdParams,
  CC.ackMessageRead
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/react",
  ...chatValidators.validateIdParams,
  CC.reactToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/remove_reaction",
  ...chatValidators.validateIdParams,
  CC.removeReactionToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id",
  ...chatValidators.validateIdParams,
  CC.deleteMessage
)

export default router
