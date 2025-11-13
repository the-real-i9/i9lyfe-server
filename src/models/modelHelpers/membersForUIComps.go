package modelHelpers

import (
	"context"
	"i9lyfe/src/appTypes/UITypes"
	"runtime"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

/*
--- Explaining the memberData for UIComp swaping job sharing between threads ---

  - The goal is to map the list of similar members ([]redis.Z) provided
    to their UI component data populating values as needed
    e.g. members containing postIds are mapped to post ui components

  - Now we don't know how many members there will be, and each mapping function
    sends a couple of requests to the redis database, depending on the data needed
    to produce a full UI component from the ID (ID or username) in the member.
    Therefore, a sequential mapping process is not a scalable approach, a more
    optimized approach will be to share the mapping job between threads
    running in parallel, where each thread works on an even portion of the members list.
    Here's how:

  - A corresponding slice of UIComp of a length equal to that of the members ([]redis.Z)
    is created. Then each thread, iterates over its allocated portion of []redis.Z,
    and inserts the generated UIComp into the []UIComp index that corresponds to
    the current index it's working on in []redis.Z.
    So that, a post's ID in ([]redis.Z)[0] has its UI component data in ([]UIComp)[0],
    or a username in ([]redis.Z)[2] has its UI component data in ([]UIComp)[2]

  - By default, the job is shared evenly between numCPUs threads, although the last thread
    can take one extra job, provided the number of jobs is of odd length.
    But, in the case where the number of jobs is less than numCPUs threads,
    the number of threads to use is truncated to the number of jobs to maintain evenness.
    So, in the least case, we hae a number of threads equal to the number of jobs.

  - `start, end := (jobsLen*j)/threadNums, jobsLen*(j+1)/threadNums`
    is how each thread takes an even portion of jobs (except the last, in an odd case)

    j is the position of the next thread starting from 0.
    jobsLen is the total number of jobs being shared by the threads
    threadNums is the number of threads sharing the jobs (numCPUs, by default)

    threadN works from start to end (exclusive),
    threadN+1 works from the threadN's end (it's start)
    to threadN+2's start (it's end) (exclusive),
    taking a number of jobs equal to threadN's

  - Each thread (goroutine) is started by errgroup, which is just a sync.WaitGroup with
    an implementation to terminate all threads (goroutines) when one signals an error,
    which is the exact behaviour we want, seeing this is one unified task
    shared by many independent processors for resource utilization
*/

func PostMembersForUIPosts(ctx context.Context, postMembers []redis.Z, clientUsername string) ([]UITypes.Post, error) {

	pmsLen := len(postMembers)

	postsAcc := make([]UITypes.Post, pmsLen)

	var threadNums int = runtime.NumCPU()

	if pmsLen < threadNums {
		// it's errorneous to spawn more threads than jobs.
		// the least we can have is a number of threads
		// equal to the number of jobs to process
		threadNums = pmsLen
	}

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (pmsLen*j)/threadNums, pmsLen*(j+1)/threadNums

			for pIndx := start; pIndx < end; pIndx++ {
				postId := postMembers[pIndx].Member.(string)
				cursor := postMembers[pIndx].Score

				post, err := BuildPostUIFromCache(sharedCtx, postId, clientUsername)
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

	umsLen := len(userMembers)

	userSnippetsAcc := make([]UITypes.UserSnippet, umsLen)

	var threadNums int = runtime.NumCPU()

	if umsLen < threadNums {
		// it's errorneous to spawn more threads than jobs.
		// the least we can have is a number of threads
		// equal to the number of jobs to process
		threadNums = umsLen
	}

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (umsLen*j)/threadNums, umsLen*(j+1)/threadNums

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

	nmsLen := len(notifMembers)

	notifSnippetsAcc := make([]UITypes.NotifSnippet, nmsLen)

	var threadNums int = runtime.NumCPU()

	if nmsLen < threadNums {
		// it's errorneous to spawn more threads than jobs.
		// the least we can have is a number of threads
		// equal to the number of jobs to process
		threadNums = nmsLen
	}

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (nmsLen*j)/threadNums, nmsLen*(j+1)/threadNums

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

func ReactorMembersForUIReactorSnippets(ctx context.Context, reactorMembers []redis.Z, postOrComment, entityId string) ([]UITypes.ReactorSnippet, error) {

	rmsLen := len(reactorMembers)

	reactorSnippetsAcc := make([]UITypes.ReactorSnippet, rmsLen)

	var threadNums int = runtime.NumCPU()

	if rmsLen < threadNums {
		// it's errorneous to spawn more threads than jobs.
		// the least we can have is a number of threads
		// equal to the number of jobs to process
		threadNums = rmsLen
	}

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (rmsLen*j)/threadNums, rmsLen*(j+1)/threadNums

			for pIndx := start; pIndx < end; pIndx++ {
				username := reactorMembers[pIndx].Member.(string)
				cursor := reactorMembers[pIndx].Score

				reactorSnippet, err := buildReactorSnippetUIFromCache(sharedCtx, username, postOrComment, entityId)
				if err != nil {
					return err
				}

				reactorSnippet.Cursor = cursor

				reactorSnippetsAcc[pIndx] = reactorSnippet
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return reactorSnippetsAcc, nil
}
