package tests

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"i9lyfe/src/services/securityServices"
)

func requestNewAccountSetup(ctx context.Context, user UserT) error {
	return createDBUser(ctx, user)
}

func requestNewAccountCleanUp(ctx context.Context, username string) error {
	return removeDBUser(ctx, username)
}

func registerUserCleanUp(ctx context.Context, username string) error {
	if err := removeDBUser(ctx, username); err != nil {
		return err
	}

	return removeCacheUser(ctx, username)
}

func signinUserPrep(ctx context.Context, user UserT) error {
	return createDBUser(ctx, user)
}

func signinUserCleanUp(ctx context.Context, username string) error {
	return removeDBUser(ctx, username)
}

func forgotPasswordPrep(ctx context.Context, user UserT) error {
	if err := createDBUser(ctx, user); err != nil {
		return err
	}

	return addCacheUser(ctx, user)
}

func forgotPasswordCleanUp(ctx context.Context, username string) error {
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
		RETURNING email, username, name_, profile_pic_url, bio, presence
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
	_, err := rdb().HSet(ctx, "users", []string{user.Username, helpers.ToJson(user)}).Result()
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
