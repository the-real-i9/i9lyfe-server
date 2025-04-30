package appGlobals

import (
	"os"

	"cloud.google.com/go/storage"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/segmentio/kafka-go"
)

var GCSClient *storage.Client

var Neo4jDriver neo4j.DriverWithContext

var KafkaWriter = &kafka.Writer{
	Addr:                   kafka.TCP(os.Getenv("KAFKA_BROKER_ADDRESS")),
	AllowAutoTopicCreation: true,
}
