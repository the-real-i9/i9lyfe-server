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
