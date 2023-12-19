import { Pool } from "pg"

const pool = new Pool()

export const dbQuery = (queryText, values, callback) => {
  return pool.query(queryText, values, callback)
}

export const getDBClient = () => {
  return pool.connect()
}

export const getDBPool = () => {
  return pool
}