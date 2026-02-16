package main

import (
	"i9lyfe/src/initializers"
	"i9lyfe/src/routes/authRoute"
	"i9lyfe/src/routes/privateRoutes"
	"i9lyfe/src/routes/publicRoutes"
	"i9lyfe/src/routes/testSessionRoute"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

//	@title			i9lyfe Backend API
//	@version		1.0
//	@description	i9lyfe Social Media Backend API.

//	@contact.name	i9ine
//	@contact.email	oluwarinolasam@gmail.com

//	@host		localhost:8000
//	@BasePath	/api

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Cookie
//	@description				JWT API key in encrypted cookie to protect private endpoints

//	@accepts	application/vnd.msgpack
//	@produces	application/vnd.msgpack

// @schemes	http https
func main() {
	defer initializers.CleanUp()

	app := fiber.New()

	app.Use(limiter.New())

	app.Use(helmet.New(helmet.Config{
		// CrossOriginResourcePolicy: "cross-origin", /* for production */
	}))

	app.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:5173"}, /* production client host */
		// AllowCredentials: true,
	}))

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Route("/api/auth", authRoute.Route)

	app.Route("/api/app/private", privateRoutes.Routes)
	app.Route("/api/app/public", publicRoutes.Routes)

	app.Route("/__test_session", testSessionRoute.Route)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))
}
