import { EventEmitter } from "node:events"
import { Kafka, logLevel, Partitioners } from "kafkajs"

export const userAlertEventEmitter = new EventEmitter()

const kafkaClient = new Kafka({
  clientId: "i9lyfe-server",
  logLevel: logLevel.NOTHING,
  brokers: [process.env.KAFKA_BROKER_ADDRESS],
})

const producer = kafkaClient.producer({
  createPartitioner: Partitioners.DefaultPartitioner,
})

await producer.connect()

export const sendNewNotification = (receiver_username, data) => {
  producer
    .send({
      topic: `i9lyfe-user-${receiver_username}-alerts`,
      messages: [
        {
          value: JSON.stringify({
            event: "new notification",
            data,
          }),
          partition: 0,
        },
      ],
    })
    .then(() => {
      userAlertEventEmitter.emit("new user alert")
    })
}

export const sendChatEvent = (event, partner_username, data) => {
  producer
    .send({
      topic: `i9lyfe-user-${partner_username}-alerts`,
      messages: [
        {
          value: JSON.stringify({
            event,
            data,
          }),
          partition: 0,
        },
      ],
    })
    .then(() => {
      userAlertEventEmitter.emit("new user alert")
    })
}

/**
 * @param {import("kafkajs").ITopicConfig[]} topics
 */
export const consumeTopics = async (topics) => {
  const admin = kafkaClient.admin()
  const consumer = kafkaClient.consumer({ groupId: "i9lyfe-topics" })

  await admin.connect()
  await admin.createTopics({ topics })
  await admin.disconnect()

  await consumer.connect()
  await consumer.subscribe({ topics: topics.map((v) => v.topic) })

  return consumer
}
