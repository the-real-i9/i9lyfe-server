import { test, expect } from "@jest/globals"
import {
  extractHashtags,
  extractMentions,
  generateMultiColumnUpdateSetParameters,
  generateMultiRowInsertValuesParameters,
} from "../utils/helpers.js"

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

test("generate multiple rows INSERT INTO [table] ... VALUES [parameters]", () => {
  const res = (rowsCount, columnsCount) =>
    generateMultiRowInsertValuesParameters(rowsCount, columnsCount)

  expect(res(3, 2)).toBe("($1, $2), ($3, $4), ($5, $6)")
  expect(res(3, 3)).toBe("($1, $2, $3), ($4, $5, $6), ($7, $8, $9)")
})

test("generate multiple columns update UPDATE [table] SET [parameters] string", async () => {
  const res = generateMultiColumnUpdateSetParameters(["name", "username"])

  console.log(res)
  expect(res).toBe("name = $1, username = $2")
})
