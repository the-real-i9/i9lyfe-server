import pg from "pg"

const { Pool } = pg

const pool = new Pool({
  connectionString: process.env.DATABASE_URI,
})

/** @param {import('pg').QueryConfig} query */
export const dbQuery = (query) => {
  return pool.query(query)
}

export const getDBClient = () => {
  return pool.connect()
}

export const getDBPool = () => {
  return pool
}
