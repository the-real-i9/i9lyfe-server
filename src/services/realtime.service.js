import { Consumer, KafkaClient } from "kafka-node"

export const initRTC = (socket) => {
  const { client_user_id } = socket.jwt_payload

  const kafkaClient = new KafkaClient({ kafkaHost: process.env.KAFKA_HOST })
  
  const consumer = new Consumer(kafkaClient, [
    { topic: `user-${client_user_id}` },
  ])

  consumer.on("message", (message) => {
    const { event, data } = JSON.parse(message.value.toString())

    socket.emit(event, data)
  })

  socket.on("start receiving post updates", (post_id) => {
    consumer.addTopics([`post-${post_id}-updates`])  
  })

  socket.on("stop receiving post updates", (post_id) => {
    consumer.removeTopics([`post-${post_id}-updates`])  
  })

  socket.on("start receiving comment updates", (comment_id) => {
    consumer.addTopics([`comment-${comment_id}-updates`])  
  })

  socket.on("stop receiving comment updates", (comment_id) => {
    consumer.removeTopics([`comment-${comment_id}-updates`])  
  })

  socket.on("disconnect", () => {
    consumer.close((err) => {
      console.error(err)
    })
  })

}