import { getAllUserConversationIds } from "../../models/ChatModel.js"

export class ChatRealtimeService {
  /** @type {import("socket.io").Server} */
  static io = null

  /** @type {Map<number, import("socket.io").Socket>} */
  static sockClients = new Map()

  /**
   * @param {import("socket.io").Server} io
   * @param {import("socket.io").Socket} socket
   */
  static async initRTC(io, socket) {
    const { client_user_id } = socket.jwt_payload
    ChatRealtimeService.io ??= io
    ChatRealtimeService.sockClients.set(client_user_id, socket)

    // On client's connection, get all the conversations they participate in and add join them to the corresponding rooms. This means that conversation rooms must have a naming convention with their coversation_id.
    const clientConversationRooms = (
      await getAllUserConversationIds(client_user_id)
    ).map(({ c_id }) => `convo-room-${c_id}`)

    socket.join(clientConversationRooms)
  }

  static createDMConversation({
    client_user_id,
    partner_user_id,
    dm_conversation_id,
  }) {
    ChatRealtimeService.sockClients
      .get(client_user_id)
      ?.join(`convo-room-${dm_conversation_id}`)
    ChatRealtimeService.sockClients
      .get(partner_user_id)
      ?.join(`convo-room-${dm_conversation_id}`)
  }

  static async addParticipantsToGroup(
    participantsUserIds,
    group_conversation_id
  ) {
    for (const new_part_user_id of participantsUserIds) {
      const userSock = ChatRealtimeService.sockClients.get(new_part_user_id)

      if (userSock) {
        userSock.join(`convo-room-${group_conversation_id}`)
        userSock.emit("new group", { group_conversation_id })
      }
    }
  }

  sendNewMessage(conversation_id, msgData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("new message", msgData)
  }

  sendMessageDelivered(conversation_id, infoData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("message delivered", infoData)
  }

  sendMessageRead(conversation_id, infoData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("message read", infoData)
  }

  sendMessageReaction(conversation_id, reactionData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("message reaction", reactionData)
  }

  sendMessageReactionRemoval(conversation_id, reactionRemovalData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("message reaction removal", reactionRemovalData)
  }

  sendMessageDeleted(conversation_id, deletionData) {
    ChatRealtimeService.io
      .to(`convo-room-${conversation_id}`)
      .emit("message deleted", deletionData)
  }

  sendGroupActivityLog(group_conversation_id, groupActivityLogData) {
    ChatRealtimeService.io
      .to(`convo-room-${group_conversation_id}`)
      .emit("new group activity", groupActivityLogData)
  }
}
