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

router.get("/users_for_chat", CC.getUsersForChatController)

router.post("/create_dm_conversation", CC.createDMConversationController)

router.get("/conversations/:dm_conversation_id")

router.post("/create_group_conversation", CC.createGroupConversationController)

router.post(
  "/conversations/:group_conversation_id/add_participants",
  CC.addParticipantsToGroupController
)

router.put(
  "/conversations/:group_conversation_id/remove_participant",
  CC.removeParticipantFromGroupController
)

router.post(
  "/conversations/:group_conversation_id/join_group",
  CC.joinGroupController
)

router.put(
  "/conversations/:group_conversation_id/leave_group",
  CC.leaveGroupController
)

router.put(
  "/conversations/:group_conversation_id/make_participant_admin",
  CC.makeParticipantAdminController
)

router.put(
  "/conversations/:group_conversation_id/remove_participant_from_admins",
  CC.removeParticipantFromAdminsController
)

router.put(
  "/conversations/:group_conversation_id/change_group_title",
  CC.changeGroupTitleController
)

router.put(
  "/conversations/:group_conversation_id/change_group_description",
  CC.changeGroupDescriptionController
)

router.get("/conversations/:group_conversation_id")

router.get("/my_conversations", CC.getMyConversationsController)

router.delete(
  "/conversations/:conversation_id",
  CC.deleteMyConversationController
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
  "/conversations/:conversation_id/messages/:message_id/ack_message_delivered",
  CC.ackMessageDeliveredController
)

router.put(
  "/conversations/:conversation_id/messages/:message_id/ack_message_read",
  CC.ackMessageReadController
)

router.post(
  "/conversations/:conversation_id/messages/:message_id/react_to_message",
  CC.reactToMessageController
)

router.delete(
  "/conversations/:conversation_id/messages/:message_id/remove_my_reaction",
  CC.removeMyReactionToMessageController
)

router.delete(
  "/conversations/:conversation_id/messages/:message_id",
  CC.deleteMyMessageController
)

export default router
