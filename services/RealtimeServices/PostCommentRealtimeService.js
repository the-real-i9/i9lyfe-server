import { getUserFolloweesIds } from "../../models/UserModel.js"

export class PostCommentRealtimeService {
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
    PostCommentRealtimeService.io ??= io
    PostCommentRealtimeService.sockClients.set(client_user_id, socket)

    /* To receive new post from those you follow */
    const followeesNewPostRooms = (
      await getUserFolloweesIds(client_user_id)
    ).map(({ followee_user_id }) => `user_${followee_user_id}_new_post_room`)

    socket.join(followeesNewPostRooms)

    /* To receive metrics update for post when in view */
    socket.on(
      "subscribe to post-comment metrics update",
      (post_or_comment_id) => {
        socket.join(
          `post-comment_${post_or_comment_id}_metrics_update_subscribers`
        )
      }
    )

    /* To stop receiving metrics update for post when out of view */
    socket.on(
      "unsubscribe from post-comment metrics update",
      (post_or_comment_id) => {
        socket.leave(
          `post-comment_${post_or_comment_id}_metrics_update_subscribers`
        )
      }
    )
  }

  sendNewPost(user_id, newPostData) {
    PostCommentRealtimeService.io
      .to(`user_${user_id}_new_post_room`)
      .emit("new post", newPostData)
  }

  sendPostCommentMetricsUpdate(post_or_comment_id, data) {
    PostCommentRealtimeService.io
      .to(`post-comment_${post_or_comment_id}_metrics_update_subscribers`)
      .emit("latest post-comment metric", data)
  }
}
