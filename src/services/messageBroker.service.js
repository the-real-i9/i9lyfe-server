import { Kafka } from "kafkajs"

const kafkaClient = new Kafka({
  clientId: "i9lyfe-server",
  brokers: [process.env.KAFKA_BROKER_ADDRESS],
})

const producer = kafkaClient.producer()

export const sendNewNotification = (receiver_username, data) => {
  producer.send({
    topic: `i9lyfe-user-${receiver_username}-alerts`,
    messages: [
      {
        value: JSON.stringify({
          event: "new notification",
          data,
        }),
      },
    ],
  })
}

export const sendChatEvent = (event, partner_username, data) => {
  producer.send({
    topic: `i9lyfe-user-${partner_username}-alerts`,
    messages: [
      {
        value: JSON.stringify({
          event,
          data,
        }),
      },
    ],
  })
}

/**
 *
 * @param {import("kafkajs").ITopicConfig[]} topics
 * @returns {import("kafkajs").Consumer}
 */
export const consumeTopics = async (topics) => {
  const kafkaClient = new Kafka({
    clientId: "i9lyfe-server",
    brokers: [process.env.KAFKA_BROKER_ADDRESS],
  })

  const admin = kafkaClient.admin()
  const consumer = kafkaClient.consumer({ groupId: "i9lyfe-topics" })

  await admin.connect()
  await admin.createTopics({ topics })
  admin.disconnect()

  await consumer.connect()
  await consumer.subscribe({ topics: topics.map((v) => v.topic) })

  return consumer
}
