package userModel

import (
	"context"
	"fmt"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Exists(ctx context.Context, uniqueIdent string) (bool, error) {
	userExists, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $1) AS exists
		`, uniqueIdent,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *userExists, nil
}

type NewUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name" db:"name_"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `json:"bio"`
	Presence      string `json:"presence"`
}

func New(ctx context.Context, email, username, password, name, bio string, birthday int64) (NewUserT, error) {
	newUser, err := db.QueryRowType[NewUserT](ctx,
		/* sql */ `
		INSERT INTO users (username, email, password_, name_, bio, birthday)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING email, username, name_, profile_pic_url, bio, presence
		`, username, email, password, name, bio, birthday,
	)
	if err != nil {
		helpers.LogError(err)
		return NewUserT{}, fiber.ErrInternalServerError
	}

	return *newUser, nil
}

type ToAuthUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name" db:"name_"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Presence      string `json:"presence"`
	Password      string `json:"-" db:"password_"`
}

func AuthFind(ctx context.Context, uniqueIdent string) (*ToAuthUserT, error) {
	user, err := db.QueryRowType[ToAuthUserT](
		ctx,
		/* sql */ `
		SELECT email, username, name_, profile_pic_url, presence, password_ 
		FROM users 
		WHERE username = $1 OR email = $1
		`, uniqueIdent,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return user, nil
}

func ChangePassword(ctx context.Context, email, newPassword string) error {
	err := db.Exec(
		ctx,
		/* sql */ `
		UPDATE users
		SET password_ = $2
		WHERE email = $1
		`, email, newPassword,
	)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func EditProfile(ctx context.Context, clientUsername string, updateKVMap map[string]any) (bool, error) {
	setChanges, params, place := "", []any{clientUsername}, 2

	for col, val := range updateKVMap {
		if setChanges != "" {
			setChanges += ", "
		}
		setChanges += fmt.Sprintf("%s = $%d", col, place)
		params = append(params, val)
		place++
	}

	done, err := db.QueryRowField[bool](
		ctx,
		fmt.Sprintf( /* sql */ `
		UPDATE users
		SET %s 
		WHERE username = $1
		RETURNING true AS done
		`, setChanges), params...,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func ChangeProfilePicture(ctx context.Context, clientUsername, pictureUrl string) (bool, error) {
	done, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		UPDATE users
		SET profile_pic_url = $2
		WHERE username = $1
		RETURNING done AS true
		`, clientUsername, pictureUrl,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func Follow(ctx context.Context, clientUsername, targetUsername string, at int64) (bool, error) {
	done, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		INSERT INTO user_follows_user (follower_username, following_username, at_)
		VALUES ($1, $2, $3)
		RETURNING true AS done
		`, clientUsername, targetUsername, at,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func Unfollow(ctx context.Context, clientUsername, targetUsername string) (bool, error) {
	done, err := db.QueryRowField[bool](
		ctx,
		/* sql */ `
		DELETE FROM user_follows_user
		WHERE follower_username = $1 AND following_username = $2
		RETURNING true AS done
		`, clientUsername, targetUsername,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func GetMentionedPosts(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })<-[:MENTIONS_USER]-(post:Post WHERE post.created_at < $offset)<-[:CREATES_POST]-(ownerUser)
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
		OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction, 
			CASE csaves 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_saved, 
			CASE creposts 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_reposted
		ORDER BY post.created_at DESC
		LIMIT $limit
		RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_mentioned_posts
		`,
		map[string]any{
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	ump, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_mentioned_posts")

	return ump, nil
}

func GetReactedPosts(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[crxn:REACTS_TO_POST]->(post:Post WHERE post.created_at < $offset)<-[:CREATES_POST]-(ownerUser)
		OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			crxn.reaction AS client_reaction, 
			CASE csaves 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_saved, 
			CASE creposts 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_reposted
		ORDER BY post.created_at DESC
		LIMIT $limit
		RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_reacted_posts
		`,
		map[string]any{
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	urp, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_reacted_posts")

	return urp, nil
}

func GetSavedPosts(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[:SAVES_POST]->(post:Post WHERE post.created_at < $offset)<-[:CREATES_POST]-(ownerUser)
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction, 
			true AS client_saved, 
			CASE creposts 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_reposted
		ORDER BY post.created_at DESC
		LIMIT $limit
		RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_saved_posts
		`,
		map[string]any{
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	usp, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_saved_posts")

	return usp, nil
}

func GetNotifications(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[:RECEIVES_NOTIFICATION]->(notif:Notification WHERE notif.created_at < $offset)

		WITH notif, toString(notif.created_at) AS created_at
		ORDER BY notif.created_at DESC
		LIMIT $limit
		RETURN collect(notif { .*, created_at }) AS notifications
		`,
		map[string]any{
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	notifs, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "notifications")

	return notifs, nil
}

func ReadNotification(ctx context.Context, clientUsername, notificationId string) error {
	res, err := db.Query(
		ctx,
		`
      MATCH (notif:Notification{ id: $notification_id })
      SET notif.is_read = true
      `,
		map[string]any{
			"client_username": clientUsername,
			"notification_id": notificationId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil
	}

	return nil
}

func GetProfile(ctx context.Context, clientUsername, targetUsername string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (profileUser:User{ username: $target_username })

		OPTIONAL MATCH (profileUser)<-[cfur:FOLLOWS_USER]-(:User{ username: $client_username })

		WITH profileUser,
			CASE cfur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows,
			coalesce(profileUser.posts_count, 0) AS posts_count,
			coalesce(profileUser.followers_count, 0) AS followers_count,
			coalesce(profileUser.following_count, 0) AS following_count

		RETURN profileUser { .username, .name, .profile_pic_url, .bio, posts_count, followers_count, following_count, client_follows } AS user_profile
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	userProfile, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "user_profile")

	return userProfile, nil
}

func GetFollowers(ctx context.Context, clientUsername, targetUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (follower:User)-[fur:FOLLOWS_USER]->(:User{ username: $target_username })
		WHERE fur.at < $offset

		OPTIONAL MATCH (follower)<-[cfur:FOLLOWS_USER]-(:User{ username: $client_username })

		WITH follower,
			CASE cfur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY fur.at DESC
		LIMIT $limit

		RETURN collect(follower { .id, .username, .profile_pic_url, client_follows }) AS user_followers
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	ufs, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_followers")

	return ufs, nil
}

func GetFollowing(ctx context.Context, clientUsername, targetUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:User{ username: $target_username })-[fur:FOLLOWS_USER]->(following:User)
		WHERE fur.at < $offset

		OPTIONAL MATCH (following)<-[cfur:FOLLOWS_USER]-(:User{ username: $client_username })

		WITH following,
			CASE cfur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY fur.at DESC
		LIMIT $limit
		RETURN collect(following { .id, .username, .profile_pic_url, client_follows }) AS user_following
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	uf, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_following")

	return uf, nil
}

func GetPosts(ctx context.Context, clientUsername, targetUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (ownerUser:User{ username: $username })-[:CREATES_POST]->(post:Post WHERE post.created_at < $offset)
		OPTIONAL MATCH (post)<-[crxn:REACTS_TO_POST]-(:User{ username: $client_username })
		OPTIONAL MATCH (post)<-[csaves:SAVES_POST]-(:User{ username: $client_username })
		OPTIONAL MATCH (post)<-[creposts:REPOSTS_POST]-(:User{ username: $client_username })
		WITH post, 
			toString(post.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction, 
			CASE csaves 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_saved, 
			CASE creposts 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS client_reposted
		ORDER BY post.created_at DESC
		LIMIT $limit
		RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS user_posts
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	ups, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "user_posts")

	return ups, nil
}

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen time.Time) error {
	var lastSeenVal any
	if presence == "online" {
		lastSeenVal = nil
	} else {
		lastSeenVal = lastSeen
	}

	err := db.Exec(
		ctx,
		/* sql */ `
		UPDATE users
		SET presence = $2, last_seen = $3
		WHERE username = $1
		`, clientUsername, presence, lastSeenVal,
	)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
