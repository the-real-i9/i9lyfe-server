import { it, expect } from "@jest/globals"
import { extractHashtags, extractMentions } from "../../utils/helpers.js"

it("should extract mentions", () => {
  const text = "This is a text with @kenny you @samuel"
  const res = extractMentions(text)

  expect(res).toContain("kenny")
  expect(res).toContain("samuel")
})

it("should extract hashtags", () => {
  const text = "This is a text with #ayo you #yemisi"
  const res = extractHashtags(text)

  expect(res).toContain("ayo")
  expect(res).toContain("yemisi")
})
