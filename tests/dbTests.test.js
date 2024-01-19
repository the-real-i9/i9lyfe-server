import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import {
  createConversation,
  createGroupConversationActivityLog,
  createMessage,
  createMessageReaction,
  createUserConversation,
  getAllUserConversations,
  getConversationHistory,
  updateGroupMembership,
  updateMessage,
} from "../models/ChatModel.js"
import { getDBClient } from "../models/db.js"

dotenv.config()

test.skip('creating direct conversation', async () => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query('BEGIN')
    const convId = await createConversation({ type: 'direct' }, dbClient)

    await createUserConversation({ participantsUserIds: [4, 5], conversation_id: convId }, dbClient)

    await dbClient.query('COMMIT')
    expect(convId).toBeTruthy()
  } catch (error) {
    dbClient.query('ROLLBACK')
    console.error(error)
  } finally {
    dbClient.release()
  }
})

test.skip("creating group conversation", async () => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")
    // We need "usernames" for activity logs to reduce the complexity of JOINs. Since activity logs are just for display. The client side can change it to "You" if the username matches that of the client, else, it uses the username
    const convId = await createConversation(
      {
        type: "group",
        title: "King Coders",
        description: "Pool of all King Coders",
        created_by: "gen_i9",
      },
      dbClient
    )

    // adding users to the group conversation
    // the first participant to be added automatically gets the "admin" role in "GroupMembership"
    await createUserConversation(
      { participantsUserIds: [4, 5], conversation_id: convId },
      dbClient
    )

    // programmatically create one activity log for each statement that adds users to the group conversation
    await createGroupConversationActivityLog(
      {
        group_conversation_id: convId,
        activity_info: {
          type: "participants_added",
          added_by: "gen_i9",
          added_participants: ["gen_i9", "mckenney"],
        },
      },
      dbClient
    )

    await dbClient.query("COMMIT")
    expect(convId).toBeTruthy()
  } catch (error) {
    dbClient.query("ROLLBACK")
    console.error(error)
  } finally {
    dbClient.release()
  }
})

test("get all user conversations", async () => {

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
    console.error(error)
  }
})

test.skip("send new group message", async () => {
  try {
    await createMessage({
      sender_user_id: 5,
      conversation_id: 10,
      msg_content: { type: "text", text_content: "Hey! What's up guys? I'm new here!" },
    })

  } catch (error) {
    expect(error).toBeUndefined()
    console.error(error)
  }
})

test.skip("update message delivery status", async () => {
  try {
    await updateMessage({
      message_id: 4,
      updateKVPairs: new Map().set("delivery_status", 'delivered'),
    })
  } catch (error) {
    console.error(error)
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
  // Remember to only allow admins to update group membership
  // create an activity log for this too
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")

    const updated = await updateGroupMembership(
      {
        admin_user_id: 4,
        participant_user_id: 5,
        group_conversation_id: 10,
        role: "admin",
      },
      dbClient
    )
    if (updated) {
      await createGroupConversationActivityLog({
        group_conversation_id: 10,
        activity_info: {
          type: "make_admin",
          admin_username: "gen_i9",
          new_admin_username: "mckenney",
        },
      }, dbClient)
    }

    dbClient.query("COMMIT")
  } catch (error) {
    dbClient.query("ROLLBACK")
    console.error(error)
    expect(error).toBeUndefined()
  } finally {
    dbClient.release()
  }
})

test.skip("remove admin from admin", async () => {
  // Remember to only allow admins to update group membership
  // create an activity log for this too
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")

    const updated = await updateGroupMembership(
      {
        admin_user_id: 5,
        participant_user_id: 4,
        group_conversation_id: 10,
        role: "member",
      },
      dbClient
    )

    if (updated) {
      await createGroupConversationActivityLog({
        group_conversation_id: 10,
        activity_info: {
          type: "remove_from_admin",
          admin_username: "mckenney",
          ex_admin_username: "gen_i9",
        },
      }, dbClient)
    }

    dbClient.query("COMMIT")
  } catch (error) {
    dbClient.query("ROLLBACK")
    console.error(error)
    expect(error).toBeUndefined()
  } finally {
    dbClient.release()
  }
})

test.skip("get conversation history", async () => {
  try {
    const res = await getConversationHistory(10)

    console.log(res)
  } catch (error) {
    console.error(error)
    expect(error).toBeUndefined()
  }
})


test.skip("react to message", async () => {
  try {
    await createMessageReaction({ message_id: 7, reactor_user_id: 5, reaction_code_point: "🥰".codePointAt() })
  } catch (error) {
    console.log(error)
    expect(error).toBeUndefined()
  }
})