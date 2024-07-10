import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/chat.controllers.js"
import * as CV from "../../middlewares/validators/chat.validators.js"
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
  CV.validateIdParams,
  CC.deleteConversation
)

router.get(
  "/conversations/:conversation_id/history",
  CV.validateIdParams,
  CC.getConversationHistory
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/send_message",
  CV.validateIdParams,
  uploadMessageFiles,
  CC.sendMessage
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/delivered",
  CV.validateIdParams,
  CC.ackMessageDelivered
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/read",
  CV.validateIdParams,
  CC.ackMessageRead
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/react",
  CV.validateIdParams,
  CC.reactToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/remove_reaction",
  CV.validateIdParams,
  CC.removeReactionToMessage
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id",
  CV.validateIdParams,
  CC.deleteMessage
)

export default router
