package contentRecommendationService

/*
-There should be several recommendation algorithms (RA) that we can run on a postId.
	- Personal RA: Based on the user's followings
	- Sub-personal RA: Based on the user's following network
	- Personal User Engagement RA: Based on the user's content kind engagement stats
	- Trending RA: Based on global content kind creation rate and hashtag usage rate
	- High Engagement RA: Based on global content kind engagement rate
	- High Priority Acccount RA: e.g. On X, everyone sees Elon Musk's post
	- Recency RA: Based on recency of the content. This algorithm will control other algorithms.
		It's weird to show Elon Musk's 2years ago content, just becauses it's in the user content queue.

	- For each user each of these alrogithms has priority over the other.
*/

/*
- We need to consider several content recommendation scenarios
	- What content kind does a first-timer see? - Since, not enough stats have been gathered to recommend content for this particular user
*/

/*
- We can fan out a content to a user's queue. But, what if the user doesn't come online?
	- This suggests that a user's content queue will be dynamic.
	- Each content pushed to the queue will have an expiration time, or maybe
		we just make sure that the Recency RA is ran on top of other RAs.
*/
