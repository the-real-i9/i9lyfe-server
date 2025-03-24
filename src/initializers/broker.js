import dotenv from "dotenv"

import { KafkaClient, Producer } from "kafka-node"

dotenv.config()

const kafkaClient = new KafkaClient({
  kafkaHost: process.env.KAFKA_BROKER_ADDRESS,
})

export const kafkaProducer = new Producer(kafkaClient)

await new Promise((resolve) => {
  kafkaProducer.on("ready", resolve)
})
