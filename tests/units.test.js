import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import {
  multipleInsertPlaceholders,
  multipleInsertReplacers,
} from "../models/postModel.js"

dotenv.config()

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

test("make three placeholders for each item to INSERT, from an array of items", () => {
  const res = multipleInsertPlaceholders(["a", "b", "c"])

  expect(res).toBe("($1, $2), ($3, $4), ($5, $6)")
})

test("make three replacers for each item INSERT, from an array of items", () => {
  const res = multipleInsertReplacers("3", ["kenny", "samuel", "dennis"])

  expect(res.toString()).toBe(
    ["3", "kenny", "3", "samuel", "3", "dennis"].toString()
  )
})


