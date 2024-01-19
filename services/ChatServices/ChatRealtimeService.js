import { getAllUserConversationIds } from "../../models/ChatModel.js"

export class ChatRealtimeService {
  /** @type {import("socket.io").Server} */
  static io = null

  /**
   * @param {import("socket.io").Server} io
   * @param {import("socket.io").Socket} socket
   */
  static async initRTC(io, socket) {
    const { client_user_id } = socket.jwt_payload
    ChatRealtimeService.io ??= io

    // On client's connection, get all the conversations they participate in and add join them to the corresponding rooms. This means that conversation rooms must have a naming convention with their coversation_id.
    const clientConversationRooms = (
      await getAllUserConversationIds(client_user_id)
    ).map((c_id) => `convo-room-${c_id}`)

    socket.join(clientConversationRooms)
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

  sendMessageDeleted(conversation_id, deletionData) {
    ChatRealtimeService.io
    .to(`convo-room-${conversation_id}`)
    .emit("message deleted", deletionData)
  }
}
