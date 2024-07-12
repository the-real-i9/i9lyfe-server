import pg from "pg"

const { Pool } = pg

const pool = new Pool({
  // database: process.env.NODE_ENV === "test" ? process.env.PGDATABASE_TEST : process.env.PGDATABASE,
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
