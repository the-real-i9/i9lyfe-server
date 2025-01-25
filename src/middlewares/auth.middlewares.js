import expressSession from "express-session"
import neo4jSessStore from "connect-neo4j"
import { neo4jDriver } from "../configs/db.js"

const Neo4jStore = neo4jSessStore(expressSession)

/**
 * @param {string} storeTableName
 * @param {string} sessionSecret
 * @param {string} cookiePath
 */
export const expressSessionMiddleware = (
  storeTableName,
  sessionSecret,
  cookiePath
) =>
  expressSession({
    store: new Neo4jStore({
      client: neo4jDriver,
      nodeLabel: storeTableName,
      disableTouch: true,
    }),
    resave: false,
    saveUninitialized: false,
    secret: sessionSecret,
    cookie: {
      maxAge: 60 * 60 * 1000,
      secure: false,
      path: cookiePath,
    },
  })
