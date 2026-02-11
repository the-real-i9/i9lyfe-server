package cache

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
)

func GetPostFromDB[T any](ctx context.Context, postId string) (post *T, err error) {
	post, err = pgDB.QueryRowType[T](
		ctx,
		/* sql */ `
		SELECT id_, owner_user, type_, media_urls, description, created_at, reposted_by_user, null as snum FROM posts
		WHERE id_ = $1
		`, postId,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return post, nil
}

func GetCommentFromDB[T any](ctx context.Context, commentId string) (comment *T, err error) {
	comment, err = pgDB.QueryRowType[T](
		ctx,
		/* sql */ `
		SELECT comment_id, username AS owner_user, comment_text, attachment_url, at_, null as snum FROM user_comments_on
		WHERE id_ = $1
		`, commentId,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return comment, nil
}

func GetChatFromDB[T any](ctx context.Context, ownerUser, partnerUser string) (chat *T, err error) {
	chat, err = pgDB.QueryRowType[T](
		ctx,
		/* sql */ `
		SELECT partner_user FROM user_chats_user
		WHERE owner_user = $1 AND partner_user = $2
		`, ownerUser, partnerUser,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return chat, nil
}

func GetChatHistoryEntryFromDB[T any](ctx context.Context, CHEId string) (CHE *T, err error) {
	CHE, err = pgDB.QueryRowType[T](
		ctx,
		/* sql */ `
		SELECT CASE WHEN type_ = "message" THEN id_ ELSE null END AS id_, 
			type_, 
			content_, 
			delivery_status, 
			created_at, 
			delivered_at, 
			read_at, 
			sender_username, 
			reactor_username, 
			emoji, 
			(SELECT json_build_object('id', rep.id_, 'content', rep.content_, 'sender_username', rep.sender_username) FROM chat_history_entry rep WHERE rep.id_ = che.reply_to) AS reply_target_msg 
		FROM chat_history_entry che
		WHERE id_ = $1
		`, CHEId,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, err
	}

	return CHE, nil
}
