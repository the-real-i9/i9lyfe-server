import { Pool } from "pg"

class PoolSingleton {
  constructor(poolConfig) {
    if (!PoolSingleton.instance) {
      PoolSingleton.instance = new Pool(poolConfig)
      return PoolSingleton.instance
    }

    return PoolSingleton.instance
  }
}

const pool = new PoolSingleton()

export const query = (queryText, values, callback) => {
  return pool.query(queryText, values, callback)
}

export const getClient = () => {
  return pool.connect()
}
