/**
 * @typedef {import("pg").PoolClient} PgPoolClient
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

import {
  generateJsonbMultiKeysSetParameters,
  generateMultiColumnUpdateSetParameters,
  generateMultiRowInsertValuesParameters,
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
    VALUES ($1) RETURNING id`,
    values: [info],
  }

  return (await dbClient.query(query)).rows[0].id
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
        ? generateJsonbMultiKeysSetParameters(
            "info",
            jsonbKeys,
            updateSetValues.length + 1
          )
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
 * @param {number[]} param0.participants_user_ids
 * @param {number} param0.conversation_id
 * @param {PgPoolClient} dbClient
 */
export const createUserConversation = async (
  { participants_user_ids, conversation_id },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "UserConversation" (user_id, conversation_id) 
    VALUES ${generateMultiRowInsertValuesParameters(
      participants_user_ids.length,
      2
    )}`,
    values: participants_user_ids
      .map((user_id) => [user_id, conversation_id])
      .flat(),
  }

  await dbClient.query(query)

  // In the group creation service, add to the GroupConversationActivityLog that these users were added
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

/**
 * @param {object} param0
 * @param {number} param0.client_user_id
 * @param {PgPoolClient} dbClient
 */
export const getAllUserConversations = async (client_user_id, dbClient) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT "conv".id AS conversation_id,
      "conv".info ->> 'type' AS conversation_type,
      "conv".info ->> 'title' AS group_title,
      "conv".info ->> 'cover_pic_url' AS group_cover_image,
      "other_user".name AS recipient_name,
      "other_user".profile_pic_url AS recipient_profile_pic,
      "other_user".connection_status AS recipient_connection_status,
      "other_user".last_active AS recipient_last_active,
      "client_user_conv".unread_messages_count,
      "last_message".msg_content - '{image_data_url,voice_data_url,video_data_url,file_url,location_coordinate,link_url}' AS last_message
      "last_activity".activity_info AS last_activity
    FROM "Conversation" "conv"
    INNER JOIN "UserConversation" "user_conv" ON "user_conv".id = "conv".id
    -- INNER JOIN "User" "user" ON "user".id = "user_conv".user_id
    LEFT JOIN "User" "other_user" ON "user_conv".user_id != $1
    LEFT JOIN "UserConversation" "client_user_conv" ON "user_conv".user_id = $1
    LEFT JOIN "Message" "last_message" ON "last_message".id = "conv".last_message_id
    LEFT JOIN "GroupConversationActivityLog" "last_activity" ON "last_activity".id = "conv".last_activity_id
    WHERE "last_message".msg_content IS NOT NULL OR "conv".info ->> 'type' = 'group'
    ORDER BY "conv".updated_at DESC
    `,
    values: [client_user_id],
  }

  await dbClient.query(query)
}

export const getAllConversationMessages = ({ conversation_id }) => {}

export const getUserConversationInfo = async ({ conversation_id }) => {}

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
 * @param {"text" | "image" | "video" | "voice" | "file" | "location" | "link"} param0.msg_content.type
 * @param {string} [param0.msg_content.text_content] Text content
 * If type is Image
 * @param {string} [param0.msg_content.image_data_url] Image URL
 * @param {string} [param0.msg_content.image_caption] Image caption.
 * If type is voice
 * @param {string} [param0.msg_content.voice_data_url] Voice data URL
 * @param {string} [param0.msg_content.voice_duration] Voice data duration
 * If type is video
 * @param {string} [param0.msg_content.video_data_url] Video URL
 * @param {string} [param0.msg_content.video_caption] Video caption.
 * If type is file
 * @param {"auido/*" | "document/*" | "compressed/*"} param0.msg_content.file_type A valid MIME file type
 * @param {string} param0.msg_content.file_url File URL
 * @param {string} param0.msg_content.file_name File name
 * If type is location
 * @param {GeolocationCoordinates} param0.msg_content.location_coordinate A valid geolocation coordinate
 * If type is link
 * @param {string} param0.msg_content.link_url Link URL
 * @param {string} param0.msg_content.link_description Link description
 * @param {PgPoolClient} dbClient
 */
export const createMessage = async (
  { sender_id, conversation_id, msg_content },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "Message" (sender_id, conversation_id, msg_content) 
    VALUES ($1, $2, $3, $4)`,
    values: [sender_id, conversation_id, msg_content],
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
  /** Extract `msg_content` values as it'd be handled separately as a `jsonb` type update
   * @type {Map<string, any> | undefined}
   */
  const msg_content = updateKVPairs.get("msg_content")

  /* Delete the values from the original Map to complete the extraction */
  msg_content || updateKVPairs.delete("msg_content")

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

  /**
   * Now we take the non-jsonb-type table columns and
   * generate multiple `SET` parameters from them as the number of table columns are arbitrary (unknown)
   *
   * We did something similar for `msg_content` jsonb keys, but for `jsonb_set` in this case
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
        ? generateJsonbMultiKeysSetParameters(
            "msg_content",
            msg_content_jsonbKeys,
            updateSetValues.length + 1
          )
        : ""
    } WHERE message_id = $${
      updateSetValues.length + msg_content_jsonbValues.length + 1
    }`,
    values: [...updateSetValues, ...msg_content_jsonbValues, message_id],
  }

  await dbClient.query(query)
}

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

/**
 * @param {object} param0
 * @param {number} param0.blocking_user_id
 * @param {number} param0.blocked_user_id
 * @param {PgPoolClient} dbClient
 */
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

/**
 * @param {object} param0
 * @param {number} param0.blocking_user_id
 * @param {number} param0.blocked_user_id
 * @param {PgPoolClient} dbClient
 */
export const deleteBlockedUser = async (
  { blocking_user_id, blocked_user_id },
  dbClient
) => {
  const query = {
    text: `DELETE FROM "BlockedUser" WHERE blocking_user_id = $1 AND blocked_user_id = $2`,
    values: [blocking_user_id, blocked_user_id],
  }

  await dbClient.query(query)
}

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
export const createGroupConversationActivityLog = async (
  { conversation_id, activity_info },
  dbClient
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    INSERT INTO "GroupConversationActivityLog" (conversation_id, activity_info) 
    VALUES ($1, $2)`,
    values: [conversation_id, activity_info],
  }

  await dbClient.query(query)
}

/**
 * @param {string} searchTerm
 * @param {PgPoolClient} dbClient
 */
export const getUsersForChat = async (searchTerm, dbClient) => {
  const query = {
    text: `
    SELECT "user".id, 
      "user".username, 
      "user".name, 
      "user".profile_pic_url, 
      "user_conv".conversation_id
    FROM "User" "user"
    LEFT JOIN "UserConversation" "user_conv" ON "user_conv".user_id = "user".id
    WHERE username LIKE '%$1%' OR name LIKE '%$1%'`,
    values: [searchTerm],
  }

  await dbClient.query(query)
}
