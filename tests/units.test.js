import { test, expect } from "@jest/globals"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import { updateUserProfile } from "../models/UserModel.js"

test("extract mentions", () => {
  const text = "This is a text with @kenny you @samuel"
  const res = extractMentions(text)

  expect(res).toContain("kenny")
  expect(res).toContain("samuel")
})

test("extract hashtags", () => {
  const text = "This is a text with #ayo you #yemisi"
  const res = extractHashtags(text)

  expect(res).toContain("ayo")
  expect(res).toContain("yemisi")
})

test("create multiple rows parameters", () => {
  const res = (rowsCount, fieldsCountPerRow) => multipleRowsParameters(rowsCount, fieldsCountPerRow)

  expect(res(3, 2)).toBe("($1, $2), ($3, $4), ($5, $6)")
  expect(res(3, 3)).toBe("($1, $2, $3), ($4, $5, $6), ($7, $8, $9)")
})

test("generate user profile update string", async () => {
  const res = await updateUserProfile(3, { name: "Kenny", username: "Bimbo" })

  console.log(res)
  expect(res.text).toBe('UPDATE "User" SET name = $1, username = $2 WHERE id = $3')
  expect(res.values[0]).toBe("Kenny")
  expect(res.values[1]).toBe("Bimbo")
  expect(res.values[2]).toBe(3)
})