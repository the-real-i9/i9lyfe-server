package userModel

import (
	"context"
	"i9lyfe/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Exists(ctx context.Context, uniqueIdent string) (bool, error) {
	res, err := db.Query(ctx,
		`
		RETURN EXISTS {
      MATCH (user:User) WHERE user.username = $uniqueIdent OR user.email = $uniqueIdent
    } AS userExists
		`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	userExists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "userExists")
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	return userExists, nil
}

func New(ctx context.Context, email, username, password, name, bio string, birthday time.Time) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
		CREATE (user:User{ email: $email, username: $username, password: $password, name: $name, birthday: $birthday, bio: $bio, profile_pic_url: "", connection_status: "offline", last_seen: datetime() })
    RETURN user { .email, .username, .name, .profile_pic_url, .connection_status } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
			"name":     name,
			"birthday": birthday.UTC(),
			"bio":      bio,
		},
	)
	if err != nil {
		log.Println("userModel.go: New:", err)
		return nil, fiber.ErrInternalServerError
	}

	newUser, _, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")
	if err != nil {
		log.Println("userModel.go: New:", err)
		return nil, fiber.ErrInternalServerError
	}

	return newUser, nil
}

func SigninFind(ctx context.Context, uniqueIdent string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (user:User)
		WHERE user.username = $uniqueIdent OR user.email = $uniqueIdent
		RETURN user { .email, .username, .name, .profile_pic_url, .connection_status, .password } AS found_user
		`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: SigninFind:", err)
		return nil, fiber.ErrInternalServerError
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func ChangePassword(ctx context.Context, email, newPassword string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (user:User{ email: $email })
		SET user.password = $newPassword
		`,
		map[string]any{
			"email":       email,
			"newPassword": newPassword,
		},
	)
	if err != nil {
		log.Println("userModel.go: SigninFind:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
