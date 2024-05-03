import { getUnreadNotificationsCount } from "../models/user.model"


export class NotificationService {
  constructor(receiver_user_id) {
    this.receiver_user_id = receiver_user_id
  }

  static io = null
  /** @type {Map<number, import("socket.io").Socket>} */
  static sockClients = new Map()

  /**
   * @param {import("socket.io").Server} io
   * @param {import("socket.io").Socket} socket
   */
  static initRTC(io, socket) {
    const { client_user_id } = socket.jwt_payload
    NotificationService.io ??= io
    NotificationService.sockClients.set(client_user_id, socket)

    new NotificationService(client_user_id).notifyUnreadNotifications()
  }

  notifyNewNotification() {
    NotificationService.sockClients
      .get(this.receiver_user_id)
      ?.emit("new notification")
  }

  pushNotification(notificationData) {
    NotificationService.sockClients
      .get(this.receiver_user_id)
      ?.emit("push notification", notificationData)
    
    this.notifyNewNotification()
  }

  async notifyUnreadNotifications() {
    const count = await getUnreadNotificationsCount(this.receiver_user_id)
    if (!Number(count)) return
    NotificationService.sockClients
      .get(this.receiver_user_id)
      ?.emit(
        "unread notifications",
        count,
        `You have ${count} unread notifications.`
      )
  }
}
