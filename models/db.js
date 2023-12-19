import { Pool } from "pg";

const pool = new Pool()

export const query = (queryText, values, callback) => {
  return pool.query(queryText, values, callback)
}

export const getClient = () => {
  return pool.connect()
}