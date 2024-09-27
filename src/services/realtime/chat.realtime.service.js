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
  }

  /**
   * @param {"new conversation" | "new message" | "message delivered" | "message read" | "message reaction" | "message reaction removed" | "message deleted"} event
   * @param {number} partner_user_id
   * @param {object} data
   */
  static send(event, partner_user_id, data) {
    ChatRealtimeService.sockClients.get(partner_user_id)?.emit(event, data)
  }
}
