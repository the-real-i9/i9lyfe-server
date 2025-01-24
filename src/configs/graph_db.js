import neo4j from "neo4j-driver"
import dotenv from "dotenv"

dotenv.config()

const driver = neo4j.driver(
  process.env.NEO4J_URL,
  neo4j.auth.basic(process.env.NEO4J_USER, process.env.NEO4J_PASSWORD),
  { disableLosslessIntegers: true }
)

export const neo4jDriver = {
  async initDB() {
    const session = this.session()

    await session.executeWrite(async (tx) => {
      await tx.run(
        `CREATE CONSTRAINT unique_user_id IF NOT EXISTS FOR (u:User) REQUIRE (u.id) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_user_username IF NOT EXISTS FOR (u:User) REQUIRE (u.username) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_user_email IF NOT EXISTS FOR (u:User) REQUIRE (u.email) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_post IF NOT EXISTS FOR (post:Post) REQUIRE (post.id) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_comment IF NOT EXISTS FOR (comment:Comment) REQUIRE (comment.id) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_repost IF NOT EXISTS FOR (repost:Repost) REQUIRE (repost.reposter_username, repost.reposted_post_id) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_hashtag IF NOT EXISTS FOR (ht:Hashtag) REQUIRE (ht.name) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_notification IF NOT EXISTS FOR (notif:Notification) REQUIRE (notif.id) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_chat IF NOT EXISTS FOR (chat:Chat) REQUIRE (chat.owner_username, chat.partner_username) IS UNIQUE`
      )
      await tx.run(
        `CREATE CONSTRAINT unique_message IF NOT EXISTS FOR (msg:Message) REQUIRE (msg.id) IS UNIQUE`
      )
      await tx.run(`CREATE INDEX post_type_idx IF NOT EXISTS FOR (post:Post) ON (post.type)`)
      await tx.run(
        `CREATE TEXT INDEX username_search_idx IF NOT EXISTS FOR (u:User) ON (u.username)`
      )
      await tx.run(
        `CREATE TEXT INDEX user_name_search_idx IF NOT EXISTS FOR (u:User) ON (u.name)`
      )
      await tx.run(
        `CREATE TEXT INDEX hashtag_name_idx IF NOT EXISTS FOR (ht:Hashtag) ON (ht.name)`
      )
      await tx.run(
        `CREATE FULLTEXT INDEX post_description_idx IF NOT EXISTS FOR (post:Post) ON EACH [post.description]`
      )
    })

    session.close()
  },
  /**
   * @param {Qeury} query
   * @param {*?} parameters
   */
  async executeWrite(query, parameters) {
    return driver.executeQuery(query, parameters)
  },

  /**
   * @param {Query} query
   * @param {*?} parameters
   */
  async executeRead(query, parameters) {
    return driver.executeQuery(query, parameters, {
      routing: neo4j.routing.READ,
    })
  },

  close() {
    driver.close()
  },

  /**
   * @param {import("neo4j-driver").SessionConfig?} config
   */
  session(config) {
    return driver.session(config)
  },
}

await neo4jDriver.initDB()
