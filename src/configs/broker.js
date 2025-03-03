import { KafkaJS } from "@confluentinc/kafka-javascript"

const { Kafka, logLevel } = KafkaJS

export const kafkaClient = new Kafka({
  kafkaJS: {
    clientId: "i9lyfe-server",
    logLevel: logLevel.NOTHING,
    brokers: [process.env.KAFKA_BROKER_ADDRESS],
  },
})

export const kafkaProducer = kafkaClient.producer()

await kafkaProducer.connect()