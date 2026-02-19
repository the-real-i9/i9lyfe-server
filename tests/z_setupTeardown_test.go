package tests

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/services/securityServices"

	"github.com/redis/go-redis/v9"
)

func getUserProfileSetup(ctx context.Context, user1, user2 UserT) (func(context.Context) error, error) {
	pipe := rdb().TxPipeline()

	newUsers := []string{
		user1.Username, helpers.ToMsgPack(user1),
		user2.Username, helpers.ToMsgPack(user2),
	}

	pipe.HSet(ctx, "users", newUsers)
	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:followers", user1.Username), redis.Z{Score: 1, Member: user2.Username})
	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:followings", user2.Username), redis.Z{Score: 1, Member: user1.Username})

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) (err error) {
		pipe := rdb().TxPipeline()

		pipe.HDel(ctx, "users", user1.Username, user2.Username)
		pipe.ZRem(ctx, fmt.Sprintf("user:%s:followers", user1.Username), user2.Username)
		pipe.ZRem(ctx, fmt.Sprintf("user:%s:followings", user2.Username), user1.Username)

		_, err = pipe.Exec(ctx)

		return
	}, nil
}

func requestNewAccountSetup(ctx context.Context, user UserT) error {
	return createDBUser(ctx, user)
}

func requestNewAccountTeardown(ctx context.Context, username string) error {
	return removeDBUser(ctx, username)
}

func registerUserTeardown(ctx context.Context, username string) error {
	if err := removeDBUser(ctx, username); err != nil {
		return err
	}

	return removeCacheUser(ctx, username)
}

func signinUserPrep(ctx context.Context, user UserT) error {
	return createDBUser(ctx, user)
}

func signinUserTeardown(ctx context.Context, username string) error {
	return removeDBUser(ctx, username)
}

func forgotPasswordPrep(ctx context.Context, user UserT) error {
	if err := createDBUser(ctx, user); err != nil {
		return err
	}

	return addCacheUser(ctx, user)
}

func forgotPasswordTeardown(ctx context.Context, username string) error {
	if err := removeDBUser(ctx, username); err != nil {
		return err
	}

	return removeCacheUser(ctx, username)
}

func createDBUser(ctx context.Context, user UserT) error {
	userPass, err := securityServices.HashPassword(user.Password)
	if err != nil {
		return err
	}

	return pgDB.Exec(ctx,
		/* sql */ `
		INSERT INTO users (username, email, password_, name_, bio, birthday)
		VALUES ($1, $2, $3, $4, $5, $6)
		`, user.Username, user.Email, userPass, user.Name, user.Bio, user.Birthday,
	)
}

func removeDBUser(ctx context.Context, username string) error {
	return pgDB.Exec(ctx,
		/* sql */ `
		DELETE FROM users
		WHERE username = $1
		`, username,
	)
}

func addCacheUser(ctx context.Context, user UserT) error {
	_, err := rdb().HSet(ctx, "users", []string{user.Username, helpers.ToMsgPack(user)}).Result()
	if err != nil {
		return err
	}

	return nil
}

func removeCacheUser(ctx context.Context, username string) error {
	_, err := rdb().HDel(ctx, "users", username).Result()
	if err != nil {
		return err
	}

	return nil
}

// sample users
var user1 = UserT{
	Email:    "suberu@gmail.com",
	Username: "suberu",
	Name:     "Suberu Garuda",
	Password: "sketeppy",
	Birthday: bday("2000-11-07"),
	Bio:      "Whatever!",
}

var user2 = UserT{
	Email:    "harveyspecter@gmail.com",
	Username: "harvey",
	Name:     "Harvey Specter",
	Password: "harvey_psl",
	Birthday: bday("1993-11-07"),
	Bio:      "Whatever!",
}

var user3 = UserT{
	Email:    "mikeross@gmail.com",
	Username: "mikeross",
	Name:     "Mike Ross",
	Password: "mikeross_psl",
	Birthday: bday("1999-11-07"),
	Bio:      "Whatever!",
}

var user4 = UserT{
	Email:    "alexwilliams@gmail.com",
	Username: "alex",
	Name:     "Alex Williams",
	Password: "williams_psl",
	Birthday: bday("1999-11-07"),
	Bio:      "Whatever!",
}
