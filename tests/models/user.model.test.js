import { test, xtest, expect } from "@jest/globals"
import dotenv from "dotenv"
import { createUser, getUser } from "../../models/user.model"

dotenv.config()

test("create user", async () => {
  const res = await getUser("myi")

  expect(res).toHaveProperty("id")

  console.log(res)
})
