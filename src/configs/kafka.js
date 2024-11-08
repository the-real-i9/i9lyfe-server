import { KafkaClient } from "kafka-node"

const kafkaClient = new KafkaClient({ kafkaHost: process.env.KAFKA_HOST })

export const getKafkaClient = () => {
  return kafkaClient
}
