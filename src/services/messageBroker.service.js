import { kafkaProducer as producer, kafkaClient } from "../configs/broker.js"


export const sendNewNotification = (receiver_username, data) => {
  producer.send({
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
        partition: 0,
      },
    ],
  })
}

/**
 * @param {import("@confluentinc/kafka-javascript").KafkaJS.ITopicConfig[]} topics
 */
export const consumeTopics = async (topics) => {
  const admin = kafkaClient.admin()
  const consumer = kafkaClient.consumer({
    kafkaJS: { groupId: "i9lyfe-topics" },
  })

  await admin.connect()
  await admin.createTopics({ topics })
  await admin.disconnect()

  await consumer.connect()
  await consumer.subscribe({ topics: topics.map((v) => v.topic) })

  return consumer
}
