// User-story-based testing for server applications
package tests

import (
	"bytes"
	"errors"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/initializers"
	"i9lyfe/src/routes/authRoute"
	"i9lyfe/src/routes/privateRoutes"
	"i9lyfe/src/routes/publicRoutes"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

const HOST_URL string = "http://localhost:8000"
const WSHOST_URL string = "ws://localhost:8000"

const signupPath string = HOST_URL + "/api/auth/signup"
const signinPath string = HOST_URL + "/api/auth/signin"
const forgotPasswordPath string = HOST_URL + "/api/auth/forgot_password"
const signoutPath string = HOST_URL + "/api/app/private/me/signout"

const appPathPriv = HOST_URL + "/api/app/private"
const appPathPublic = HOST_URL + "/api/app/public"
const wsPath = WSHOST_URL + "/api/app/private/ws"

type UserT struct {
	Email          string
	Username       string
	Name           string
	Password       string
	Birthday       int64
	Bio            string
	SessionCookie  string
	WSConn         *websocket.Conn
	ServerEventMsg chan map[string]any
}

var app *fiber.App

func TestMain(m *testing.M) {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}

	defer initializers.CleanUp()

	app = fiber.New()
	app.Use(helmet.New())
	app.Use(cors.New())

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Route("/api/auth", authRoute.Route)
	app.Route("/api/app/private", privateRoutes.Routes)
	app.Route("/api/app/public", publicRoutes.Routes)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	go func() {
		app.Listen("0.0.0.0:" + PORT)
	}()

	waitReady := time.NewTimer(2 * time.Second)
	<-waitReady.C

	c := m.Run()

	waitFinish := time.NewTimer(2 * time.Second)
	<-waitFinish.C

	app.Shutdown()

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
		helpers.LogError(err)
	}

	return bd.UTC().UnixMilli()
}

func startResumableUpload(uploadUrl string, contentType string, t *testing.T) string {
	req, err := http.NewRequest("POST", uploadUrl, nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-goog-resumable", "start")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		t.Log("unexpected error:", rb)

		require.NoError(t, fmt.Errorf("failed to start resumable upload: %s", res.Status))
	}

	sessionUrl := res.Header.Get("Location")

	if sessionUrl == "" {
		require.NoError(t, fmt.Errorf("No resumable session URL returned"))
	}

	return sessionUrl
}

const CHUNK_SIZE int64 = 512 * 1024 // 512 KiB

func uploadFileInChunks(sessionUrl string, filePath string, contentType string, onProgress func(int64, int64, *testing.T), t *testing.T) {
	stat, err := os.Stat(filePath)
	require.NoError(t, err)
	fileSize := stat.Size()

	file, err := os.Open(filePath)
	require.NoError(t, err)

	var offset int64 = 0
	buffer := make([]byte, CHUNK_SIZE)

	defer file.Close()

	for offset < fileSize {
		bytesToRead := min(CHUNK_SIZE, fileSize-offset)

		secReader := io.NewSectionReader(file, offset, bytesToRead)

		bytesRead, err := secReader.Read(buffer)
		if !errors.Is(err, io.EOF) {
			require.NoError(t, err)
		}

		chunk := bytes.NewReader(buffer[0:bytesRead])
		end := offset + int64(bytesRead) - 1

		req, err := http.NewRequest("PUT", sessionUrl, chunk)
		require.NoError(t, err)

		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", bytesRead))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, end, fileSize))

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if res.StatusCode == http.StatusPermanentRedirect {
			offset += int64(bytesRead)
			onProgress(offset, fileSize, t)
			continue
		}

		if res.StatusCode == http.StatusOK {
			offset += int64(bytesRead)
			onProgress(fileSize, fileSize, t)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		t.Log("unexpected error:", rb)

		require.NoError(t, fmt.Errorf("chunk upload failed: %s", res.Status))
	}
}

func logProgress(sent int64, total int64, t *testing.T) {
	percent := float64(sent) / float64(total) * 100

	t.Logf("Upload progress: %.2f%% (%d/%d)", percent, sent, total)
}
