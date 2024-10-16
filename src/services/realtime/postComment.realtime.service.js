import { User } from "../../models/user.model.js"

export class PostCommentRealtimeService {
  /** @type {import("socket.io").Server} */
  static io = null

  /**
   * @param {import("socket.io").Server} io
   * @param {import("socket.io").Socket} socket
   */
  static async initRTC(io, socket) {
    const { client_user_id } = socket.jwt_payload
    PostCommentRealtimeService.io ??= io

    /* To receive new post from those you follow */
    const followeesNewPostRooms = (
      await User.getFolloweesIds(client_user_id)
    ).map((followee_user_id) => `user_${followee_user_id}_new_post_room`)

    socket.join(followeesNewPostRooms)

    /* To start receiving metrics update for post when in view */
    socket.on("subscribe to post metrics update", (post_id) => {
      socket.join(`post_${post_id}_metrics_update_subscribers`)
    })

    /* To stop receiving metrics update for post when out of view */
    socket.on("unsubscribe from post metrics update", (post_id) => {
      socket.leave(`post_${post_id}_metrics_update_subscribers`)
    })

    /* To start receiving metrics update for post when in view */
    socket.on("subscribe to comment metrics update", (comment_id) => {
      socket.join(`comment_${comment_id}_metrics_update_subscribers`)
    })

    /* To stop receiving metrics update for comment when out of view */
    socket.on("unsubscribe from comment metrics update", (comment_id) => {
      socket.leave(`comment_${comment_id}_metrics_update_subscribers`)
    })
  }

  static sendNewPost(user_id, newPostData) {
    PostCommentRealtimeService.io
      ?.to(`user_${user_id}_new_post_room`)
      .emit("new post", newPostData)
  }

  static sendPostMetricsUpdate(post_id, data) {
    PostCommentRealtimeService.io
      ?.to(`post_${post_id}_metrics_update_subscribers`)
      .emit(`latest post metric`, data)
  }

  static sendCommentMetricsUpdate(comment_id, data) {
    PostCommentRealtimeService.io
      ?.to(`comment_${comment_id}_metrics_update_subscribers`)
      .emit(`latest comment metric`, data)
  }
}
