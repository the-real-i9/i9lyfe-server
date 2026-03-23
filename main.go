package main

import (
	"i9lyfe/src/domain/auth/authMiddlewares"
	"i9lyfe/src/domain/auth/authRoutes"
	"i9lyfe/src/domain/chat/chatRoutes"
	"i9lyfe/src/domain/postComment/postCommentRoutes"
	"i9lyfe/src/domain/user/privateUserRoutes"
	"i9lyfe/src/domain/user/publicUserRoutes"
	"i9lyfe/src/domain/user/userWSCommMan/wsCommRoute"
	"i9lyfe/src/initializers"
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

		router.Route("", postCommentRoutes.Routes)
		router.Route("", privateUserRoutes.Routes)
		router.Route("", chatRoutes.Routes)
		router.Route("", wsCommRoute.Route)
	})
	app.Route("/api/app/public", func(router fiber.Router) {
		router.Use(authMiddlewares.UserAuthOptional)

		router.Route("", publicUserRoutes.Routes)

	})

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	log.Fatalln(app.Listen("0.0.0.0:" + PORT))
}
