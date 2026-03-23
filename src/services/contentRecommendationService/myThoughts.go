package contentRecommendationService

/*
-There should be several recommendation algorithms (RA) that we can run on a postId.
	- Personal RA: Based on the user's followings (users or interests)
	- Sub-personal RA: Based on the user's following network
	- Personal User Engagement RA: Based on the user's engagement rate with similar kinds of content.
	- Trending RA: Based on global similar content creation rate and hashtag usage rate
	- High Engagement RA: Based on global similar content engagement rate
	- High Priority Acccount RA: e.g. On X, everyone sees Elon Musk's post... I think.
*/

/*
- We need to consider several content recommendation scenarios
	- What content kind does a first-timer see? - Since, not enough stats have been gathered to recommend content for this particular user
		- Answer: We need to have a general feed cache where we queue posts that fulfill the app's recommendation requirements, based on non-personal RAs.
*/

/*
- Generally, when a post is fanned out to the recommendation system service.
	- The service pushes it to users' feed cache who it has decided they'll receive the post, based on the algorithms above.

- A content isn't recommended only at the point of creation. It can be recommended later on based on its performance metrics or interaction by users.
	- Suppose a user you follow likes a post. Then, it's worth showing you that post too.
	- A post whose metrics increase within a short period of time should be made to reach more users,
		so is a post in which it's owner paid for a boost

- Each user in the app has a feed cache
*/
