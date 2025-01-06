import neo4j from "neo4j-driver"
import dotenv from "dotenv"

dotenv.config()

const driver = neo4j.driver(
  process.env.NEO4J_URL,
  neo4j.auth.basic(process.env.NEO4J_USER, process.env.NEO4J_PASSWORD),
  { disableLosslessIntegers: true }
)

export const neo4jDriver = {
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
    return driver.executeQuery(query, parameters, { routing: neo4j.routing.READ })
  },

  close() {
    driver.close()
  },

  /**
   * @param {import("neo4j-driver").SessionConfig?} config 
   */
  session(config) {
    return driver.session(config)
  }
}


