import { User } from "../models/user.model.js"



export class NotificationService {
  constructor(receiver_user_id) {
    this.receiver_user_id = receiver_user_id
  }

  /** @type {Map<number, import("socket.io").Socket>} */
  static sockClients = new Map()

  /** @param {import("socket.io").Socket} socket */
  static initRTC(socket) {
    const { client_user_id } = socket.jwt_payload
    NotificationService.sockClients.set(client_user_id, socket)
  
    // notify client of unread notifications when they're connected
    new NotificationService(client_user_id).notifyUnreadNotifications()
  }

  // send a new notification update
  sendNotification(notificationData) {
    NotificationService.sockClients
      .get(this.receiver_user_id)
      ?.emit("new notification", notificationData)
  }

  async notifyUnreadNotifications() {
    const count = await User.getUnreadNotificationsCount(this.receiver_user_id)
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
