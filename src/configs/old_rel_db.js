import pg from "pg"

const { Pool } = pg

const pool = new Pool()

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
