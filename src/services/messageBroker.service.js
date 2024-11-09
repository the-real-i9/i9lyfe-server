import { KafkaClient, Producer } from "kafka-node"

const kafkaClient = new KafkaClient({ kafkaHost: process.env.KAFKA_HOST })

const producer = new Producer(kafkaClient)

export const sendNewNotification = (receiver_user_id, data) => {
  producer.send(
    [
      {
        topic: `user-${receiver_user_id}-alerts`,
        messages: JSON.stringify({
          event: "new notification",
          data,
        }),
      },
    ],
    (err) => {
      err && console.error(err)
    }
  )
}
export const sendPostUpdate = (post_id, data) => {
  producer.send(
    [
      {
        topic: `post-${post_id}-updates`,
        messages: JSON.stringify({
          event: "latest post update",
          data,
        }),
      },
    ],
    (err) => {
      err && console.error(err)
    }
  )
}

export const sendCommentUpdate = (comment_id, data) => {
  producer.send(
    [
      {
        topic: `comment-${comment_id}-updates`,
        messages: JSON.stringify({
          event: "latest comment update",
          data,
        }),
      },
    ],
    (err) => {
      err && console.error(err)
    }
  )
}

export const sendChatEvent = (event, partner_user_id, data) => {
  producer.send(
    [
      {
        topic: `user-${partner_user_id}-alerts`,
        messages: JSON.stringify({
          event,
          data,
        }),
      },
    ],
    (err) => {
      err && console.error(err)
    }
  )
}

/**
 * @param {string} topic
 */
export const createTopic = (topic) => {
  producer.createTopics([topic], (err) => {
    err && console.error(err)
  })
}
