import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/chat.controllers.js"
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
      res.status(err.status).send({ error: err.inner.message })
    } else {
      next(err)
    }
  }
)

router.post("/create_conversation", CC.createConversationController)

router.get("/my_conversations", CC.getMyConversationsController)

router.delete(
  "/conversations/:conversation_id",
  CC.deleteConversationController
)

router.get(
  "/conversations/:conversation_id/history",
  CC.getConversationHistoryController
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/send_message",
  uploadMessageFiles,
  CC.sendMessageController
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/delivered",
  CC.ackMessageDeliveredController
)

router.put(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/read",
  CC.ackMessageReadController
)

router.post(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/react",
  CC.reactToMessageController
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id/remove_reaction",
  CC.removeReactionToMessageController
)

router.delete(
  "/conversations/:conversation_id/partner/:partner_user_id/messages/:message_id",
  CC.deleteMessageController
)

export default router
