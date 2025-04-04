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
      MATCH (user:User WHERE user.username = $uniqueIdent OR user.email = $uniqueIdent)
    } AS user_exists
		`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	userExists, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")

	return userExists, nil
}

func New(ctx context.Context, email, username, password, name, bio string, birthday time.Time) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
		CREATE (user:User{ email: $email, username: $username, password: $password, name: $name, birthday: $birthday, bio: $bio, profile_pic_url: "", connection_status: "offline", last_seen: $at })
    RETURN user { .email, .username, .name, .profile_pic_url, .connection_status } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
			"name":     name,
			"birthday": birthday,
			"bio":      bio,
			"at":       time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("userModel.go: New:", err)
		return nil, fiber.ErrInternalServerError
	}

	newUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return newUser, nil
}

func AuthFind(ctx context.Context, uniqueIdent string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (user:User WHERE user.username = $unique_ident OR user.email = $unique_ident)
		RETURN user { .email, .username, .name, .profile_pic_url, .connection_status, .password } AS found_user
		`,
		map[string]any{
			"unique_ident": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: AuthFind:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func Client(ctx context.Context, clientUsername string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (user:User { username: $client_username })
		RETURN user { .email, .username, .name, .profile_pic_url, .connection_status } AS client_user
		`,
		map[string]any{
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: Client:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	client_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "client_user")

	return client_user, nil
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
		log.Println("userModel.go: ChangePassword:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func EditProfile(ctx context.Context, clientUsername string, updateKVMap map[string]any) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (user:User{ username: $client_username })
    SET user += $update_kv_map
		`,
		map[string]any{
			"client_username": clientUsername,
			"update_kv_map":   updateKVMap,
		},
	)
	if err != nil {
		log.Println("userModel.go: EditProfile:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func ChangeProfilePicture(ctx context.Context, clientUsername, pictureUrl string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (user:User{ username: $client_username })
		SET user.profile_pic_url = $profile_pic_url
		`,
		map[string]any{
			"client_username": clientUsername,
			"profile_pic_url": pictureUrl,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangeProfilePicture:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func Follow(ctx context.Context, clientUsername, targetUsername string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username }), (targetUser:User{ username: $target_username })
		MERGE (clientUser)-[:FOLLOWS_USER]->(targetUser)

		WITH targetUser, clientUser
		CREATE (targetUser)-[:RECEIVES_NOTIFICATION]->(followNotif:Notification:FollowNotification{ id: randomUUID(), type: "follow", is_read: false, created_at: $at, follower_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

		WITH followNotif, toString(followNotif.created_at) AS created_at
		RETURN followNotif { .*,  created_at } AS follow_notif
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
			"at":              time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("userModel.go: Follow:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	follow_notif, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "follow_notif")

	return follow_notif, nil
}

func Unfollow(ctx context.Context, clientUsername, targetUsername string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (:User{ username: $client_username })-[fr:FOLLOWS_USER]->(:User{ username: $target_username })
    DELETE fr
		`,
		map[string]any{
			"client_username": clientUsername,
			"target_username": targetUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: Unfollow:", err)
		return nil
	}

	return nil
}

func GetMentionedPosts(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })<-[:MENTIONS_USER]-(post:Post WHERE post.created_at < $offset)
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
		OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			clientUser { .username, .profile_pic_url } AS owner_user,
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
		log.Println("userModel.go: GetMentionedPosts:", err)
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
		MATCH (clientUser:User{ username: $client_username })-[cxrn:REACTS_TO_POST]->(post:Post WHERE post.created_at < $offset)
		OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			clientUser { .username, .profile_pic_url } AS owner_user,
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
		log.Println("userModel.go: GetReactedPosts:", err)
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
		MATCH (clientUser:User{ username: $client_username })-[:SAVES_POST]->(post:Post WHERE post.created_at < $offset)
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
		OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
		WITH post, 
			toString(post.created_at) AS created_at, 
			clientUser { .username, .profile_pic_url } AS owner_user,
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
		log.Println("userModel.go: GetSavedPosts:", err)
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
		log.Println("userModel.go: GetNotifications:", err)
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
		log.Println("userModel.go: ReadNotification:", err)
		return fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil
	}

	return nil
}
