package modelHelpers

import (
	"context"
	"i9lyfe/src/appTypes/UITypes"
	"runtime"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func PostMembersForUIPosts(ctx context.Context, postMembers []redis.Z, clientUsername string) ([]UITypes.Post, error) {
	// data parellelism
	// data is shared equally between numCPUs goroutines for parallel processing
	// the last goroutine gets a bigger portion if the shared data has a odd number length

	pmsLen := len(postMembers)

	postsAcc := make([]UITypes.Post, pmsLen)

	var numCPUs int = runtime.NumCPU()

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range numCPUs {
		eg.Go(func() error {
			j := i
			start, end := (pmsLen*j)/numCPUs, pmsLen*(j+1)/numCPUs

			for pIndx := start; pIndx < end; pIndx++ {
				postId := postMembers[pIndx].Member.(string)
				cursor := postMembers[pIndx].Score

				post, err := buildPostUIFromCache(sharedCtx, postId, clientUsername)
				if err != nil {
					return err
				}

				post.Cursor = cursor

				postsAcc[pIndx] = post
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return postsAcc, nil
}

func UserMembersForUIUserSnippets(ctx context.Context, userMembers []redis.Z, clientUsername string) ([]UITypes.UserSnippet, error) {
	// data parellelism
	// data is shared equally between numCPUs goroutines for parallel processing
	// the last goroutine gets a bigger portion if the shared data has a odd number length

	umsLen := len(userMembers)

	userSnippetsAcc := make([]UITypes.UserSnippet, umsLen)

	var numCPUs int = runtime.NumCPU()

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range numCPUs {
		eg.Go(func() error {
			j := i
			start, end := (umsLen*j)/numCPUs, umsLen*(j+1)/numCPUs

			for pIndx := start; pIndx < end; pIndx++ {
				username := userMembers[pIndx].Member.(string)
				cursor := userMembers[pIndx].Score

				userSnippet, err := buildUserSnippetUIFromCache(sharedCtx, username, clientUsername)
				if err != nil {
					return err
				}

				userSnippet.Cursor = cursor

				userSnippetsAcc[pIndx] = userSnippet
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return userSnippetsAcc, nil
}

func NotifMembersForUINotifSnippets(ctx context.Context, notifMembers []redis.Z) ([]UITypes.NotifSnippet, error) {
	// data parellelism
	// data is shared equally between numCPUs goroutines for parallel processing
	// the last goroutine gets a bigger portion if the shared data has a odd number length

	nmsLen := len(notifMembers)

	notifSnippetsAcc := make([]UITypes.NotifSnippet, nmsLen)

	var numCPUs int = runtime.NumCPU()

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range numCPUs {
		eg.Go(func() error {
			j := i
			start, end := (nmsLen*j)/numCPUs, nmsLen*(j+1)/numCPUs

			for pIndx := start; pIndx < end; pIndx++ {
				notifId := notifMembers[pIndx].Member.(string)
				cursor := notifMembers[pIndx].Score

				notifSnippet, err := buildNotifSnippetUIFromCache(sharedCtx, notifId)
				if err != nil {
					return err
				}

				notifSnippet.Cursor = cursor

				notifSnippetsAcc[pIndx] = notifSnippet
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return notifSnippetsAcc, nil
}
