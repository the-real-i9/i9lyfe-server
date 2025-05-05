// User-story-based testing for server applications
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const HOST_URL string = "http://localhost:8000"
const WSHOST_URL string = "ws://localhost:8000"

const appPathPriv = HOST_URL + "/api/app/private"
const appPathPublic = HOST_URL + "/api/app/public"

type UserT struct {
	Email         string
	Username      string
	Name          string
	Password      string
	Birthday      int64
	Bio           string
	SessionCookie string
	WSConn        *websocket.Conn
	ServerWSMsg   chan map[string]any
}

func TestMain(m *testing.M) {
	dbDriver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	defer dbDriver.Close(ctx)

	// cleaup db
	neo4j.ExecuteQuery(ctx, dbDriver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)

	c := m.Run()

	os.Exit(c)
}

func makeReqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func succResBody[T any](body io.ReadCloser) (T, error) {
	var d T

	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return d, err
	}

	if err := json.Unmarshal(bt, &d); err != nil {
		return d, err
	}

	return d, nil
}

func errResBody(body io.ReadCloser) (string, error) {
	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(bt), nil
}

func bday(bdaystr string) int64 {
	bd, err := time.Parse(time.DateOnly, bdaystr)
	if err != nil {
		log.Println("main_test.go", err)
	}

	return bd.UTC().UnixMilli()
}
