import { User } from "../models/user.model.js"



export class NotificationService {
  /** @type {Map<number, import("socket.io").Socket>} */
  static sockClients = new Map()

  /** @param {import("socket.io").Socket} socket */
  static initRTC(socket) {
    const { client_user_id } = socket.jwt_payload
    NotificationService.sockClients.set(client_user_id, socket)
  
    // notify client of unread notifications when they're connected
    NotificationService.notifyUnreadNotifications(client_user_id)
  }

  // send a new notification update
  static sendNotification(receiver_user_id, notificationData) {
    NotificationService.sockClients
      .get(receiver_user_id)
      ?.emit("new notification", notificationData)
  }

  static async notifyUnreadNotifications(client_user_id) {
    const count = await User.getUnreadNotificationsCount(client_user_id)
    if (!Number(count)) return
    NotificationService.sockClients
      .get(client_user_id)
      ?.emit(
        "unread notifications",
        count,
        `You have ${count} unread notifications.`
      )
  }
}
