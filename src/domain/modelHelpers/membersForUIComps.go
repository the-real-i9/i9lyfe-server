package modelHelpers

/* func x() {
	threadNums := min(pmsLen, runtime.NumCPU())

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (pmsLen*j)/threadNums, pmsLen*(j+1)/threadNums

			for pIndx := start; pIndx < end; pIndx++ {
				partnerUser := partnerMembers[pIndx].Member.(string)

				chatSnippet, err := buildChatSnippetUIFromCache(sharedCtx, clientUsername, partnerUser)
				if err != nil {
					return err
				}

				chatSnippetsAcc[pIndx] = chatSnippet
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
}
*/
