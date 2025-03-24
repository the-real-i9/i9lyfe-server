import { Consumer, KafkaClient } from "kafka-node"
import { kafkaProducer as producer } from "../initializers/broker.js"

export const sendNewNotification = async (receiver_username, data) => {
  producer.send(
    [
      {
        topic: `i9lyfe-user-${receiver_username}-alerts`,
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

export const sendChatEvent = (event, partner_username, data) => {
  producer.send(
    [
      {
        topic: `i9lyfe-user-${partner_username}-alerts`,
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
 *
 * @param {import("kafka-node").OffsetFetchRequest[]} topics
 */
export const consumeTopics = async (topics) => {
  const kafkaClient = new KafkaClient({
    kafkaHost: process.env.KAFKA_BROKER_ADRESS,
  })

  /** @type {import("kafka-node").CreateTopicRequest[]} */
  const topicsToCreate = topics.map((v) => ({
    topic: v.topic,
    partitions: 1,
    replicationFactor: 1,
  }))

  await new Promise((resolve) => {
    kafkaClient.createTopics(topicsToCreate, (err, result) => {
      if (err) {
        console.error(err)
        return resolve()
      }

      resolve(result)
    })
  })

  const consumer = new Consumer(kafkaClient, topics, { autoCommit: true })

  return consumer
}
