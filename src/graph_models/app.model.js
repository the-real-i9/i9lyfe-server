import { neo4jDriver } from "../configs/graph_db.js"

export class App {
  static async searchAndFilterPosts({
    term,
    filter,
    limit,
    offset,
    client_user_id,
  }) {
    let applyFilter = ""

    if (filter !== "all") {
      applyFilter = "post.type = $filter AND "
    }

    const { records } = await neo4jDriver.executeRead(
      `
        MATCH (post:Post WHERE ${applyFilter}post.description CONTAINS $term)<-[:CREATES_POST]-(ownerUser:User), (clientUser:User{ id: $client_user_id })
    
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
        ORDER BY post.created_at DESC
        OFFSET $offset
        LIMIT $limit
        RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS search_results
        `,
      { term, filter, client_user_id, limit, offset }
    )

    return records[0].get("search_results")
  }

  static async searchHashtags({ term, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (ht:Hashtag WHERE ht.name CONTAINS $term)<-[:INCLUDES_HASHTAG]-(post:Post)
      WITH ht.name AS hashtag, count(post) AS posts_count
      ORDER BY post_count DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect({ hashtag, posts_count }) AS search_results
      `,
      { term, limit, offset }
    )

    return records[0].get("search_results")
  }

  static async searchUsers({ term, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (user:User WHERE user.username CONTAINS $term OR user.name CONTAINS $term)
      ORDER BY user.username, user.name
      OFFSET $offset
      LIMIT $limit
      RETURN collect(user { .username, .name, .profile_pic_url, .bio }) AS search_results
      `,
      { term, limit, offset }
    )

    return records[0].get("search_results")
  }

  static async getHashtagPosts({
    hashtag_name,
    filter,
    limit,
    offset,
    client_user_id,
  }) {
    let applyFilter = ""

    if (filter != "all") {
      applyFilter = " WHERE post.type = $filter"
    }

    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (ht:Hashtag{ name: $hashtag_name })<-[:INCLUDES_HASHTAG]-(post:Post${applyFilter})<-[:CREATES_POST]-(ownerUser:User), (clientUser:User{ id: $client_user_id })

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
      ORDER BY post.created_at DESC
      OFFSET $offset
      LIMIT $limit
      RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS search_results
      `,
      { filter, hashtag_name, limit, offset, client_user_id }
    )

    return records[0].get("search_results")
  }
}
