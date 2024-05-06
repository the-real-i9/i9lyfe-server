/**
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

import {
  stripNulls,
} from "../utils/helpers.js"
import { dbQuery } from "./db.js"

/**
 * @param {object} client
 * @param {number} client.user_id
 * @param {string} client.username
 * @param {number} partner_user_id
 */
export const createConversation = async (client, partner_user_id, init_message) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT client_res, partner_res FROM create_conversation($1, $2, $3)",
    values: [
      client.user_id,
      partner_user_id,
      init_message,
    ],
  }

  // return needed details
  return (await dbQuery(query)).rows[0]
}

export const deleteUserConversation = async (
  client_user_id,
  conversation_id
) => {
  const query = {
    text: `
    UPDATE user_conversation
    SET deleted = true
    WHERE user_id = $1 AND conversation_id = $2`,
    values: [client_user_id, conversation_id],
  }

  await dbQuery(query)
}


/**
 * @param {number} client_user_id
 */
export const getAllUserConversations = async (client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT user_convos FROM get_user_conversations($1)",
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows[0].user_convos
}

/**
 * To retrieve history in chunks, from (-ve)N offset to the newest history (0)
 * First fetch N rows from DESC row set
 * Finaly, reorder the row set to ASC
 *
 * This is how you can display conversation history in a chat history page
 * @example
 * SELECT * FROM
 * (SELECT * FROM "ConversationHistory"
 * WHERE conversation_id = $1
 * ORDER BY created_at DESC
 * LIMIT 20)
 * ORDER BY created_at ASC
 * @param {number} conversation_id
 * @param {number} limit
 * @param {number} offset
 */
export const getConversationHistory = async ({
  conversation_id,
  limit,
  offset,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT history FROM get_conversation_history($1, $2, $3)
    `,
    values: [conversation_id, limit, offset],
  }

  return (await dbQuery(query)).rows[0].history
}

/**
 * @param {object} param0
 * @param {number} param0.sender_user_id
 * @param {number} param0.conversation_id
 * @param {object} param0.msg_content
 * @param {"text" | "emoji" | "image" | "video" | "voice" | "file" | "location" | "link"} param0.msg_content.type
 * @param {string} [param0.msg_content.text_content] Text content. If type is text
 *
 * @param {string} [param0.msg_content.emoji_code_point] Emoji code. If type is emoji
 *
 * @param {string} [param0.msg_content.image_data_url] Image URL. If type is image
 * @param {string} [param0.msg_content.image_caption] Image caption. If type is image
 *
 * @param {string} [param0.msg_content.voice_data_url] Voice data URL. If type is voice
 * @param {string} [param0.msg_content.voice_duration] Voice data duration. If type is voice
 *
 * @param {string} [param0.msg_content.video_data_url] Video URL. If type is video
 * @param {string} [param0.msg_content.video_caption] Video caption. If type is video
 *
 * @param {"auido/*" | "document/*" | "compressed/*"} param0.msg_content.file_type A valid MIME file type. If type is file
 * @param {string} param0.msg_content.file_url File URL. If type is file
 * @param {string} param0.msg_content.file_name File name. If type is file
 *
 * @param {GeolocationCoordinates} param0.msg_content.location_coordinate A valid geolocation coordinate. If type is location
 *
 * @param {string} param0.msg_content.link_url Link URL. If type is link
 * @param {string} param0.msg_content.link_description Link description. If type is file
 */
export const createMessage = async ({
  client_user_id,
  conversation_id,
  msg_content,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT client_res, partner_res FROM create_message($1, $2, $3)",
    values: [conversation_id, client_user_id, msg_content],
  }

  return (await dbQuery(query)).rows[0]
}


export const acknowledgeMessageDelivered = async ({client_user_id, conversation_id, message_id, delivery_time}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT ack_msg_delivered($1, $2, $3, $4)",
    values: [client_user_id, conversation_id, message_id, delivery_time],
  }

  await dbQuery(query)
}

export const acknowledgeMessageRead = async (client_user_id, conversation_id, message_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: "SELECT ack_msg_delivered($1, $2, $3)",
    values: [client_user_id, conversation_id, message_id],
  }

  return await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {number} param0.reactor_user_id
 * @param {number} param0.reaction_code_point
 */
export const createMessageReaction = async ({
  message_id,
  reactor_user_id,
  reaction_code_point,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO message_reaction (message_id, reactor_user_id, reaction_code_point) 
    VALUES ($1, $2, $3)`,
    values: [message_id, reactor_user_id, reaction_code_point],
  }

  await dbQuery(query)
}

/**
 * @param {number} message_id
 * @param {number} reactor_user_id
 */
export const deleteMessageReaction = async (message_id, reactor_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    DELETE FROM message_reaction WHERE message_id = $1 AND reactor_user_id = $2`,
    values: [message_id, reactor_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} blocking_user_id
 * @param {number} blocked_user_id
 */
export const createBlockedUser = async (blocking_user_id, blocked_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO blocked_user (blocking_user_id, blocked_user_id) 
    VALUES ($1, $2)`,
    values: [blocking_user_id, blocked_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.blocking_user_id
 * @param {number} param0.blocked_user_id
 */
export const deleteBlockedUser = async (blocking_user_id, blocked_user_id) => {
  const query = {
    text: 'DELETE FROM blocked_user WHERE blocking_user_id = $1 AND blocked_user_id = $2',
    values: [blocking_user_id, blocked_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {number} param0.reporter_user_id
 * @param {number} param0.reported_user_id
 * @param {string} param0.reason
 */
export const createReportedMessage = async ({
  message_id,
  reporter_user_id,
  reported_user_id,
  reason,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO reported_message (message_id, reporter_user_id, reported_user_id, reason) 
    VALUES ($1, $2, $3, $4)`,
    values: [message_id, reporter_user_id, reported_user_id, reason],
  }

  await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.deleter_user_id
 * @param {number} param0.message_id
 * @param {"me" | "everyone"} param0.deleted_for
 */
export const createMessageDeletionLog = async ({
  deleter_user_id,
  message_id,
  deleted_for,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
  INSERT INTO message_deletion_log (deleter_user_id, message_id, deleted_for) 
  VALUES ($1, $2, $3)`,
    values: [deleter_user_id, message_id, deleted_for],
  }

  await dbQuery(query)
}

/**
 * @param {string} search
 */
export const getUsersToChat = async (client_user_id, search) => {
  const query = {
    text: "SELECT users_to_chat FROM get_users_to_chat($1, $2)",
    values: [`%${search}%`, client_user_id],
  }

  return (await dbQuery(query)).rows[0].users_to_chat
}
