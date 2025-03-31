package main

import (
	"i9lyfe/src/initializers"
	"i9lyfe/src/routes/authRoutes"
	"i9lyfe/src/routes/privateRoutes"
	"i9lyfe/src/routes/publicRoutes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer initializers.CleanUp()

	socketio, c := initializers.InitSocket()

	defer func() {
		socketio.Close(func(err error) {
			log.Println("error closing socket server", err)
		})
	}()

	app := fiber.New()

	app.Use(helmet.New())
	app.Use(cors.New())

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Get("/socket.io", adaptor.HTTPHandler(socketio.ServeHandler(c)))

	app.Route("/api/auth", authRoutes.Routes)

	app.Route("/api/app/private", privateRoutes.Routes)
	app.Route("/api/app/public", publicRoutes.Routes)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))
}
