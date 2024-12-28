import neo4j from "neo4j-driver"
import dotenv from "dotenv"

dotenv.config()

export const neo4jDriver = neo4j.driver(
  process.env.NEO4J_URL,
  neo4j.auth.basic(process.env.NEO4J_USER, process.env.NEO4J_PASSWORD)
)
