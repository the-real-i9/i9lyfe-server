import { getUnreadNotificationsCount } from "../models/UserModel.js"

export class NotificationService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  static io = null
  /** @type {Map<number, import("socket.io").Socket>} */
  static sockClients = new Map()

  /**
   * @param {import("socket.io").Server} io
   * @param {import("socket.io").Socket} socket
   */
  static initWebSocket(io, socket) {
    const { client_user_id } = socket.jwt_payload
    NotificationService.io = io
    NotificationService.sockClients.set(client_user_id, socket)

    new NotificationService(client_user_id).notifyUnreadNotifications()
  }

  notifyNewNotification() {
    NotificationService.sockClients
      .get(this.client_user_id)
      .emit("new_notification", "just do {new_notifications_count}++")
  }

  async notifyUnreadNotifications() {
    const count = await getUnreadNotificationsCount(this.client_user_id)
    NotificationService.sockClients
      .get(this.client_user_id)
      .emit(
        "unread_notifications",
        count,
        `You have ${count} unread notifications.`
      )
  }
}
