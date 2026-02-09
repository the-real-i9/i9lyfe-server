package helpers

import (
	"errors"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

func LogError(err error) {
	if err == nil {
		return
	}

	pc, file, line, ok := runtime.Caller(1)
	fn := "unknown"
	if !ok {
		file = "???"
		line = 0
	} else {
		fn = runtime.FuncForPC(pc).Name()
	}

	log.Printf("[ERROR] %s:%d %s(): %v\n", file, line, fn, err)
}

func HandleDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			// Foreign key violation
			return humanizeForeignKeyError(pgErr)
		case "UX001":
			return fiber.NewError(fiber.StatusNotFound, pgErr.Message)
		}
	}

	return fiber.ErrInternalServerError
}

func humanizeForeignKeyError(pgErr *pgconn.PgError) error {
	switch pgErr.ConstraintName {

	case "user_comments_on_post_id_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified post does not exist")

	case "user_comments_on_parent_comment_id_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified comment does not exist")

	case "user_reacts_to_post_post_id_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified post does not exist")

	case "user_reacts_to_comment_comment_id_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified comment does not exist")

	case "user_saves_post_post_id_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified post does not exist")

	case "user_follows_user_following_username_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified user does not exist")

	case "user_chats_user_partner_user_fkey":
		return fiber.NewError(fiber.StatusNotFound, "the specified user does not exist")

	default:
		return fiber.ErrInternalServerError
	}
}
