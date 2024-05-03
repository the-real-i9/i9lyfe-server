import express from "express"
import { expressjwt } from "express-jwt"
import dotenv from "dotenv"

import * as CC from "../../controllers/ChatControllers.js"
import { uploadMessageFiles } from "../../middlewares/appMiddlewares.js"

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

router.get("/users_to_chat", CC.getUsersToChatController)

router.post("/create_dm_conversation", CC.createDMConversationController)

router.post("/create_group_conversation", CC.createGroupConversationController)

router.post(
  "/group_conversation/add_participants",
  CC.addParticipantsToGroupController
)

router.put(
  "/group_conversation/remove_participant",
  CC.removeParticipantFromGroupController
)

router.post("/group_conversation/join_group", CC.joinGroupController)

router.put("/group_conversation/leave_group", CC.leaveGroupController)

router.put(
  "/group_conversation/make_participant_admin",
  CC.makeParticipantAdminController
)

router.put(
  "/group_conversation/remove_participant_from_admins",
  CC.removeParticipantFromAdminsController
)

router.put(
  "/group_conversation/change_group_title",
  CC.changeGroupTitleController
)

router.put(
  "/group_conversation/change_group_description",
  CC.changeGroupDescriptionController
)

router.get("/conversations", CC.getConversationsController)

router.get("/conversations/:conversation_id", CC.getConversationController)

router.delete(
  "/conversations/:conversation_id",
  CC.deleteConversationController
)

router.get(
  "/conversations/:conversation_id/history",
  CC.getConversationHistoryController
)

router.post(
  "/conversations/:conversation_id/send_message",
  uploadMessageFiles,
  CC.sendMessageController
)

router.put(
  "/conversations/:conversation_id/messages/:message_id/delivered",
  CC.ackMessageDeliveredController
)

router.put(
  "/conversations/:conversation_id/messages/:message_id/read",
  CC.ackMessageReadController
)

router.post(
  "/conversations/:conversation_id/messages/:message_id/react",
  CC.reactToMessageController
)

router.delete(
  "/conversations/:conversation_id/messages/:message_id/remove_reaction",
  CC.removeReactionToMessageController
)

router.delete(
  "/conversations/:conversation_id/messages/:message_id",
  CC.deleteMessageController
)

export default router
