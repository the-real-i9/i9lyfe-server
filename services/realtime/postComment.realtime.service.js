import { getUserFolloweesIds } from "../../models/user.model.js"


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
    ).map((followee_user_id) => `user_${followee_user_id}_new_post_room`)

    socket.join(followeesNewPostRooms)

    /* To receive metrics update for post when in view */
    socket.on(
      "subscribe to post metrics update",
      (post_id) => {
        socket.join(
          `post_${post_id}_metrics_update_subscribers`
        )
      }
    )

    /* To stop receiving metrics update for post when out of view */
    socket.on(
      "unsubscribe from post metrics update",
      (post_id) => {
        socket.leave(
          `post_${post_id}_metrics_update_subscribers`
        )
      }
    )

    socket.on(
      "subscribe to comment metrics update",
      (comment_id) => {
        socket.join(
          `comment_${comment_id}_metrics_update_subscribers`
        )
      }
    )

    /* To stop receiving metrics update for post when out of view */
    socket.on(
      "unsubscribe from comment metrics update",
      (comment_id) => {
        socket.leave(
          `comment_${comment_id}_metrics_update_subscribers`
        )
      }
    )
  }

  static sendNewPost(user_id, newPostData) {
    PostCommentRealtimeService.io
      ?.to(`user_${user_id}_new_post_room`)
      .emit("new post", newPostData)
  }


  /**
   * @param {object} param0 
   * @param {"post" | "comment"} param0.entity 
   */
  static sendEntityMetricsUpdate({entity, entity_id, data}) {
    PostCommentRealtimeService.io
      ?.to(`${entity}_${entity_id}_metrics_update_subscribers`)
      .emit(`latest ${entity} metric`, data)
  }
}
