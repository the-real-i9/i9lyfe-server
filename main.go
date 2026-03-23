package main

import (
	"i9lyfe/src/domain/auth/authRoutes"
	"i9lyfe/src/domain/user/userWSCommMan/wsCommRoute"
	"i9lyfe/src/initializers"
	"i9lyfe/src/middlewares/authMiddlewares"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer initializers.CleanUp()

	app := fiber.New(fiber.Config{
		MsgPackEncoder: msgpack.Marshal,
		MsgPackDecoder: msgpack.Unmarshal,
	})

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

	app.Route("/api/auth", authRoutes.Routes)

	app.Route("/api/app/private", func(router fiber.Router) {
		router.Use(authMiddlewares.UserAuthRequired)

		router.Route("", postCommentRoute.Route)
		router.Route("", privateUserRoute.Route)
		router.Route("", chatRoute.Route)
		router.Route("", wsCommRoute.Route)
	})
	app.Route("/api/app/public", func(router fiber.Router) {
		router.Use(authMiddlewares.UserAuthOptional)

		router.Route("", publicUserRoute.Route)

	})

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))
}
