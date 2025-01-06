import { neo4jDriver } from "../configs/graph_db.js";


export async function getPost(post_id, client_user_id) {
  const { records } = await neo4jDriver.executeRead(
    `
    MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser)
    WHERE EXISTS {
      MATCH (post)-[:INCLUDES_HASHTAG]->(:Hashtag{ name: "trending" })
      UNION
      MATCH (:User{ id: $client_user_id })-[:FOLLOWS_USER]->(ownerUser)
      UNION
      MATCH (:User{ id: $client_user_id })-[:FOLLOWS_USER]->(:User)-[:FOLLOWS_USER]->(ownerUser)
    }
    RETURN post { .*, owner_user: ownerUser { .id, .username, .profile_pic_url } } AS the_post
    `,
    { post_id, client_user_id }
  )

  return records[0].get('the_post')
}