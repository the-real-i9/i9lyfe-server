/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

import {
  generateJsonbMultiKeysSetParameters,
  generateMultiColumnUpdateSetParameters,
} from "../utils/helpers"

/**
 * @param {object} info
 * @param {"individual" | "group"} info.type
 * @param {string} [info.title] Group title, if `type` is "group"
 * @param {string} [info.description] Group description, if `type` is "group"
 * @param {string} [info.cover_image_url] Group cover image, if `type` is "group"
 * @param {PgPoolClient} dbClient
 */
export const createConversation = async (info, dbClient) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "Conversation" (info) 
    VALUES ($1, $2)`,
    values: [info],
  }

  await dbClient.query(query)
}

// needs a trigger
/**
 * @param {object} param0
 * @param {number} param0.conversation_id
 * @param {Map<string, any>} param0.updateKVPairs
 * @param {PgPoolClient} dbClient
 */
export const updateConversation = async (
  { conversation_id, updateKVPairs },
  dbClient
) => {
  /** @type {Map<string, any> | undefined} */
  const info = updateKVPairs.get("info")
  info || updateKVPairs.delete("info")

  const [updateSetCols, updateSetValues] = [
    [...updateKVPairs.keys()],
    [...updateKVPairs.values()],
  ]

  const [jsonbKeys, jsonbValues] = info
    ? [[...info.keys()], [...info.values()]]
    : [[], []]

  const query = {
    text: `UPDATE "Conversation" SET ${generateMultiColumnUpdateSetParameters(
      updateSetCols
    )} ${
      jsonbKeys.length
        ? `info = jsonb_set(${generateJsonbMultiKeysSetParameters(
            "info",
            jsonbKeys,
            updateSetValues.length + 1
          )})`
        : ""
    } WHERE conversation_id = $${
      updateSetValues.length + jsonbValues.length + 1
    }`,
    values: [...updateSetValues, ...jsonbValues, conversation_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.conversation_id
 * @param {PgPoolClient} dbClient
 */
export const createUserConversation = async (
  { user_id, conversation_id },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "UserConversation" (user_id, conversation_id) 
    VALUES ($1, $2)`,
    values: [user_id, conversation_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.conversation_id
 * @param {Map<string, any>} param0.updateKVPairs
 * @param {PgPoolClient} dbClient
 */
export const updateUserConversation = async (
  { user_id, conversation_id, updateKVPairs },
  dbClient
) => {
  const [updateSetCols, updateSetValues] = [
    [...updateKVPairs.keys()],
    [...updateKVPairs.values()],
  ]

  const query = {
    text: `UPDATE "UserConversation" SET ${generateMultiColumnUpdateSetParameters(
      updateSetCols
    )} WHERE user_id = $${updateSetValues.length + 1} AND conversation_id = $${
      updateSetValues.length + 2
    }`,
    values: [...updateSetValues, user_id, conversation_id],
  }

  await dbClient.query(query)
}

export const deleteUserConversation = async ({ user_id }) => {}
export const getAllUserConversations = async ({ user_id }) => {}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.group_conversation_id
 * @param {"admin" | "member"} param0.role
 * @param {PgPoolClient} dbClient
 */
export const createGroupMembership = async (
  { user_id, group_conversation_id, role },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "GroupMembership" (user_id, group_conversation_id, role) 
    VALUES ($1, $2, $3)`,
    values: [user_id, group_conversation_id, role],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.group_conversation_id
 * @param {Map<string, any>} param0.updateKVPairs
 * @param {PgPoolClient} dbClient
 */
export const updateGroupMembership = async (
  { user_id, group_conversation_id, updateKVPairs },
  dbClient
) => {
  const [updateSetCols, updateSetValues] = [
    [...updateKVPairs.keys()],
    [...updateKVPairs.values()],
  ]

  const query = {
    text: `UPDATE "GroupMembership" SET ${generateMultiColumnUpdateSetParameters(
      updateSetCols
    )} WHERE group_conversation_id = $${
      updateSetValues.length + 1
    } AND user_id = $${updateSetValues.length + 2}`,
    values: [...updateSetValues, group_conversation_id, user_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.sender_id
 * @param {number} param0.conversation_id
 * @param {object} param0.msg_content
 * @param {"text" | "image" | "voice"} param0.msg_content.type
 * @param {string} [param0.msg_content.text_content] Text content
 * @param {string} param0.msg_content.image_data_url Image URL
 * @param {string} param0.msg_content.voice_data_url Voice data URL
 * @param {string} param0.msg_content.image_description Image description. If `type` is Image
 * @param {object} param0.msg_attachment
 * @param {"audio" | "video" | "location" | "document" | "compressed" | "other"} param0.msg_attachment.type
 * @param {string} param0.msg_attachment.file_type A valid MIME file type
 * @param {string} param0.msg_attachment.file_url File URL
 * @param {GeolocationCoordinates} param0.msg_attachment.location_coordinate A valid geolocation coordinate
 * @param {PgPoolClient} dbClient
 */
export const createMessage = async (
  { sender_id, conversation_id, msg_content, msg_attachment },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "Message" (sender_id, conversation_id, msg_content, msg_attachment) 
    VALUES ($1, $2, $3, $4)`,
    values: [sender_id, conversation_id, msg_content, msg_attachment],
  }

  await dbClient.query(query)
}

/**
 * The algorithm in this function explains how all `UPDATE` algorithms were implemented dynamically in this app (save a few ones). The documentation was added here as this seems to be the most complex implementation.
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {Map<string, any>} param0.updateKVPairs
 * @param {PgPoolClient} dbClient
 */
export const updateMessage = async (
  { message_id, updateKVPairs },
  dbClient
) => {
  /** Extract `msg_content` and `msg_attachment` values as they'd be handled separately as a `jsonb` type update
   * @type {(Map<string, any> | undefined)[]}
   */
  const [msg_content, msg_attachment] = [
    updateKVPairs.get("msg_content"),
    updateKVPairs.get("msg_attachment"),
  ]

  /* Delete the values from the original Map to complete the extraction */
  msg_content || updateKVPairs.delete("msg_content")
  msg_attachment || updateKVPairs.delete("msg_attachment")

  /* 
  The remnants are the non-jsonb-type table columns.
  We separate table columns/keys from values as we they'd be handled separately
   */
  const [updateSetCols, updateSetValues] = [
    [...updateKVPairs.keys()],
    [...updateKVPairs.values()],
  ]

  /* We seperate `msg_content` jsonb keys from values */
  const [msg_content_jsonbKeys, msg_content_jsonbValues] = msg_content
    ? [[...msg_content.keys()], [...msg_content.values()]]
    : [[], []]

  /* We do the same for `msg_attachment` */
  const [msg_attachment_jsonbKeys, msg_attachment_jsonbValues] = msg_attachment
    ? [[...msg_attachment.keys()], [...msg_attachment.values()]]
    : [[], []]

  /**
   * Now we take the non-jsonb-type table columns and
   * generate multiple `SET` parameters from them as the number of table columns are arbitrary (unknown)
   *
   * We did something similar for `msg_content` and `msg_attachment` jsonb keys, but for `jsonb_set` in this case
   *
   * Observe that, as we dynamically add parameters/placeholders for each object's keys' values, we spread/align/match (...) the values in the `values` parameter of the `QueryConfig`, and we jump `{sumOfValuesLengthFromPreviousObjects} + 1` so we can spread/align/match `(...)` the values of the next object just after the values of the previous one.
   *
   * And finally, for the `WHERE` clause, we do a final `+ 1` jump as its parameter is the final. If there were multiple parameters for the `WHERE` clause, we'll jump depending on their position from away from the previously added dynamic parameters.
   * @type {PgQueryConfig}
   */
  const query = {
    text: `UPDATE "Message" SET ${generateMultiColumnUpdateSetParameters(
      updateSetCols
    )} ${
      msg_content_jsonbKeys.length
        ? `msg_content = jsonb_set(${generateJsonbMultiKeysSetParameters(
            "msg_content",
            msg_content_jsonbKeys,
            updateSetValues.length + 1
          )})`
        : ""
    } ${
      msg_attachment_jsonbKeys.length
        ? `msg_attachment = jsonb_set(${generateJsonbMultiKeysSetParameters(
            "msg_attachment",
            msg_attachment_jsonbKeys,
            updateSetValues.length + msg_content_jsonbValues.length + 1
          )})`
        : ""
    } WHERE message_id = $${
      updateSetValues.length +
      msg_content_jsonbValues.length +
      msg_attachment_jsonbValues.length +
      1
    }`,
    values: [
      ...updateSetValues,
      ...msg_content_jsonbValues,
      ...msg_attachment_jsonbValues,
      message_id,
    ],
  }

  await dbClient.query(query)
}
export const deleteMessage = async ({ message_id }) => {}
export const getAllConversationMessages = ({ conversation_id }) => {}

/**
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {number} param0.reactor_user_id
 * @param {number} param0.reaction_code_point
 * @param {PgPoolClient} dbClient
 */
export const createMessageReaction = async (
  { message_id, reactor_user_id, reaction_code_point },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "MessageReaction" (message_id, reactor_user_id, reaction_code_point) 
    VALUES ($1, $2, $3)`,
    values: [message_id, reactor_user_id, reaction_code_point],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {number} param0.reactor_user_id
 * @param {PgPoolClient} dbClient
 */
export const deleteMessageReaction = async (
  { message_id, reactor_user_id },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    DELETE FROM "MessageReaction" WHERE message_id = $1 AND reactor_user_id = $2`,
    values: [message_id, reactor_user_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {"online" | "offline"} param0.status
 * @param {PgPoolClient} dbClient
 */
export const updateUserConnectionStatus = async (
  { user_id, connection_status },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    UPDATE "User" SET connection_status = $1, last_active = $2 WHERE user_id = $3`,
    values: [
      connection_status,
      connection_status === "offline" ? new Date() : null,
      user_id,
    ],
  }

  await dbClient.query(query)
}

export const createBlockedUser = async (
  { blocking_user_id, blocked_user_id },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "BlockedUser" (blocking_user_id, blocked_user_id) 
    VALUES ($1, $2)`,
    values: [blocking_user_id, blocked_user_id],
  }

  await dbClient.query(query)
}
export const deleteBlockedUser = async ({ blocked_user_id }) => {}

/**
 * @param {object} param0
 * @param {number} param0.message_id
 * @param {number} param0.reporter_user_id
 * @param {number} param0.reported_user_id
 * @param {string} param0.reason
 * @param {PgPoolClient} dbClient
 */
export const createReportedMessage = async (
  { message_id, reporter_user_id, reported_user_id, reason },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "ReportedMessage" (message_id, reporter_user_id, reported_user_id, reason) 
    VALUES ($1, $2, $3, $4)`,
    values: [message_id, reporter_user_id, reported_user_id, reason],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.message_id
 * @param {"me" | "everyone"} param0.deleted_for
 * @param {PgPoolClient} dbClient
 */
export const createMessageDeletionLog = async (
  { user_id, message_id, deleted_for },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
  INSERT INTO "MessageDeletionLog" (user_id, message_id, deleted_for) 
  VALUES ($1, $2, $3)`,
    values: [user_id, message_id, deleted_for],
  }

  await dbClient.query(query)
}

/**
 * @param {*} param0
 * @param {number} param0.conversation_id
 * @param {number} param0.subject_user_id
 * @param {string} param0.activity_message
 * @param {PgPoolClient} dbClient
 */
export const createConversationActivityLog = async (
  { conversation_id, subject_user_id, activity_message },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "ConversationActivityLog" (conversation_id, subject_user_id, activity_message) 
    VALUES ($1, $2, $3)`,
    values: [conversation_id, subject_user_id, activity_message],
  }

  await dbClient.query(query)
}

export const getUsersForChat = (searchTerm, dbClient) => {}
