package initializers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/segmentio/kafka-go"
	"google.golang.org/api/option"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GCS_API_KEY")))
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient

	return nil
}

func initNeo4jDriver() error {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := driver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("initializers.go: initNeo4jDriver: sess.Close:", err)
		}
	}()

	_, err2 := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var err error

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE (u.username) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_user_email IF NOT EXISTS FOR (u:User) REQUIRE (u.email) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_post IF NOT EXISTS FOR (post:Post) REQUIRE (post.id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_comment IF NOT EXISTS FOR (comment:Comment) REQUIRE (comment.id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_repost IF NOT EXISTS FOR (repost:Repost) REQUIRE (repost.reposter_username, repost.reposted_post_id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_hashtag IF NOT EXISTS FOR (ht:Hashtag) REQUIRE (ht.name) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_notification IF NOT EXISTS FOR (notif:Notification) REQUIRE (notif.id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_chat IF NOT EXISTS FOR (chat:Chat) REQUIRE (chat.owner_username, chat.partner_username) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_message IF NOT EXISTS FOR (msg:Message) REQUIRE (msg.id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE INDEX post_type_idx IF NOT EXISTS FOR (post:Post) ON (post.type)`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE TEXT INDEX username_search_idx IF NOT EXISTS FOR (u:User) ON (u.username)`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE TEXT INDEX user_name_search_idx IF NOT EXISTS FOR (u:User) ON (u.name)`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE TEXT INDEX hashtag_name_idx IF NOT EXISTS FOR (ht:Hashtag) ON (ht.name)`, nil)
		if err != nil {
			return nil, err
		}
		_, err = tx.Run(ctx, `CREATE FULLTEXT INDEX post_description_idx IF NOT EXISTS FOR (post:Post) ON EACH [post.description]`, nil)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err2 != nil {
		return err2
	}

	appGlobals.Neo4jDriver = driver

	return nil
}

func initKafkaWriter() error {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_BROKER_ADDRESS")),
		AllowAutoTopicCreation: true,
	}

	appGlobals.KafkaWriter = w

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	initKafkaWriter()

	return nil
}

func CleanUp() {
	if err := appGlobals.KafkaWriter.Close(); err != nil {
		log.Println("failed to close kafka writer:", err)
	}

	if err := appGlobals.Neo4jDriver.Close(context.TODO()); err != nil {
		log.Println("error closing neo4j driver", err)
	}
}
