import pg from "pg"

const { Pool } = pg

const pool = new Pool({
  connectionString:
    process.env.NODE_ENV === "production"
      ? `postgresql://${process.env.POSTGRES_USER}:${process.env.POSTGRES_PASSWORD}@${process.env.POSTGRES_HOST}:5432/${process.env.POSTGRES_DB}`
      : process.env.DABABASE_URI,
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
