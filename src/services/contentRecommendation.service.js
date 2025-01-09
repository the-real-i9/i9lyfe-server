import { neo4jDriver } from "../configs/graph_db.js"

export async function getPost(post_id, client_user_id) {
  const { records } = await neo4jDriver.executeRead(
    `
    MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser:User)
    WHERE EXISTS {
      MATCH (post)-[:INCLUDES_HASHTAG]->(:Hashtag{ name: "trending" })
      UNION
      MATCH (:User{ id: $client_user_id })-[:FOLLOWS_USER]->(ownerUser)
      UNION
      MATCH (:User{ id: $client_user_id })-[:FOLLOWS_USER]->(:User)-[:FOLLOWS_USER]->(ownerUser)
    }
    WITH post, toString(post.created_at) AS created_at, ownerUser { .id, .username, .profile_pic_url } AS owner_user
    RETURN post { .*, created_at, owner_user } AS the_post
    `,
    { post_id, client_user_id }
  )

  return records[0].get("the_post")
}

export async function getHomePosts({ client_user_id, limit, offset, types }) {
  const { records } = await neo4jDriver.executeRead(
    `
    MATCH (clientUser:User{ id: $client_user_id })
    MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post WHERE post.type IN $types)
    WHERE EXISTS {
      MATCH (post)-[:INCLUDES_HASHTAG]->(:Hashtag{ name: "trending" })
      UNION
      MATCH (clientUser)-[:FOLLOWS_USER]->(ownerUser)
      UNION
      MATCH (clientUser)-[:FOLLOWS_USER]->(:User)-[:FOLLOWS_USER]->(ownerUser)
    }

    OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
    OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
    OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)

    WITH post, 
      toString(post.created_at) AS created_at, 
      ownerUser { .id, .username, .profile_pic_url } AS owner_user,
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
    ORDER BY post.created_at
    OFFSET toInteger($offset)
    LIMIT toInteger($limit)
    RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS feed_posts
    `,
    { types, client_user_id, limit, offset }
  )

  return records[0].get("feed_posts")
}

export async function getExplorePosts({ client_user_id, types, limit, offset }) {
  const { records } = await neo4jDriver.executeRead(
    `
    MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post WHERE post.type IN $types), (clientUser:User{ id: $client_user_id })

    OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
    OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
    OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)

    WITH post, 
      toString(post.created_at) AS created_at, 
      ownerUser { .id, .username, .profile_pic_url } AS owner_user,
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
    ORDER BY post.created_at DESC, post.reactions_count DESC, post.comments_count DESC
    OFFSET toInteger($offset)
    LIMIT toInteger($limit)
    RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS explore_posts
    `,
    { types, client_user_id, limit, offset }
  )

  return records[0].get("explore_posts")
}
