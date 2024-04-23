import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import {
  createDMConversation,
  createMessage,
  createMessageReaction,
  createUserConversation,
  getAllUserConversations,
  getConversationHistory,
  updateMessage,
} from "../models/ChatModel.js"
import { getDBClient } from "../models/db.js"
import {
  createComment,
  getAllCommentsOnPost_OR_RepliesToComment,
} from "../models/PostCommentModel.js"

dotenv.config()

test.skip("creating direct conversation", async () => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")
    const convId = await createDMConversation({ type: "direct" }, dbClient)

    await createUserConversation(
      { participantsUserIds: [4, 5], conversation_id: convId },
      dbClient
    )

    await dbClient.query("COMMIT")
    expect(convId).toBeTruthy()
  } catch (error) {
    dbClient.query("ROLLBACK")
    // console.error(error)
  } finally {
    dbClient.release()
  }
})

test.skip("creating group conversation", async () => {
  
})

test.skip("get all user conversations", async () => {
  const userConvos = await getAllUserConversations(5)

  console.log(userConvos)

  expect(userConvos).toBeTruthy()
})

test.skip("send new direct message", async () => {
  try {
    await createMessage({
      sender_user_id: 4,
      conversation_id: 9,
      msg_content: { type: "text", text_content: "Yeah! I'm good too!" },
    })
  } catch (error) {
    expect(error).toBeUndefined()
    // console.error(error)
  }
})

test.skip("send new group message", async () => {
  try {
    await createMessage({
      sender_user_id: 5,
      conversation_id: 10,
      msg_content: {
        type: "text",
        text_content: "Hey! What's up guys? I'm new here!",
      },
    })
  } catch (error) {
    expect(error).toBeUndefined()
    // console.error(error)
  }
})

test.skip("update message delivery status", async () => {
  try {
    await updateMessage({
      message_id: 4,
      updateKVPairs: new Map().set("delivery_status", "delivered"),
    })
  } catch (error) {
    // console.error(error)
    expect(error).toBeUndefined()
  }
  // only update a group message's delivery status when it is true for all members of the group
})

test.skip("update user conversation last read message", async () => {
  // this is automatically triggered for direct conversations
})

test.skip("update user conversation unread messages count", async () => {
  // this is automatically triggered for direct conversations
})

test.skip("make group member, admin", async () => {
  
})

test.skip("remove admin from admin", async () => {
  
})

test.skip("get conversation history", async () => {
  try {
    const res = await getConversationHistory(10)

    console.log(res)
  } catch (error) {
    // console.error(error)
    expect(error).toBeUndefined()
  }
})

test.skip("react to message", async () => {
  try {
    await createMessageReaction({
      message_id: 7,
      reactor_user_id: 5,
      reaction_code_point: "🥰".codePointAt(),
    })
  } catch (error) {
    console.log(error)
    expect(error).toBeUndefined()
  }
})

test.skip("create comment", async () => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query('BEGIN')
    const data = await createComment({
      commenter_user_id: 4,
      content_owner_user_id: 5,
      post_or_comment_id: 4,
      post_or_comment: "post",
      comment_text:
        "This is a comment for testing Common Table Expressions",
      attachment_url: null,
    }, dbClient)
    console.log(data)
  
  dbClient.query("COMMIT")
} catch (error) {
  dbClient.query("ROLLBACK")
  // console.error(error)
  expect(error).toBeUndefined()
} finally {
  dbClient.release()
}
})

test.skip("comments/replies to post/comment", async () => {
  try {
    const data = await getAllCommentsOnPost_OR_RepliesToComment({
      post_or_comment: "post",
      post_or_comment_id: 4,
      client_user_id: 4,
    })
    console.log(data)
  } catch (error) {
    console.log(error)
    expect(error).toBeUndefined()
  }
})

