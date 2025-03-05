import dotenv from "dotenv"
/* import { KafkaJS } from "@confluentinc/kafka-javascript"

dotenv.config()

const { Kafka, logLevel } = KafkaJS

export const kafkaClient = new Kafka({
  kafkaJS: {
    clientId: "i9lyfe-server",
    logLevel: logLevel.ERROR,
    brokers: [process.env.KAFKA_BROKER_ADDRESS],
  },
})

export const kafkaProducer = kafkaClient.producer()
export const kafkaAdmin = kafkaClient.admin()

await kafkaProducer.connect()
await kafkaAdmin.connect() */

import { KafkaClient, Producer } from "kafka-node"

dotenv.config()

const kafkaClient = new KafkaClient({
  kafkaHost: process.env.KAFKA_BROKER_ADDRESS,
})

export const kafkaProducer = new Producer(kafkaClient)

await new Promise((resolve) => {
  kafkaProducer.on("ready", resolve)
})

// export const kafkaAdmin = new Admin()
