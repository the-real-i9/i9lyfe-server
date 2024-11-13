import expressSession from "express-session"
import pgSession from "connect-pg-simple"
import { getDBPool } from "../configs/db.js"


const PGStore = pgSession(expressSession)

/**
 * @param {string} storeTableName
 * @param {string} sessionSecret
 * @param {string} cookiePath
 * @returns
 */
export const expressSessionMiddleware = (
  storeTableName,
  sessionSecret,
  cookiePath
) =>
  expressSession({
    store: new PGStore({
      pool: getDBPool(),
      tableName: storeTableName,
      createTableIfMissing: true,
    }),
    resave: false,
    saveUninitialized: false,
    secret: sessionSecret,
    cookie: {
      maxAge: 1 * 60 * 60 * 1000,
      secure: false,
      path: cookiePath,
    },
  })
