/**
 * @typedef {import("pg").QueryConfig} PgQueryConfig
 */

import {
  generateMultiRowInsertValuesParameters,
  stripNulls,
} from "../utils/helpers.js"
import { dbQuery } from "./db.js"

/**
 * @param {object} client
 * @param {number} client.user_id
 * @param {string} client.username
 * @param {number} partner_user_id
 */
export const createDMConversation = async (client, partner_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
  WITH dm_convo_cte AS (
    INSERT INTO "Conversation" (info) 
    VALUES ($1) 
    RETURNING id AS dm_conversation_id
  ), user_convo_cte AS (
    INSERT INTO "UserConversation" (user_id, conversation_id) 
    VALUES ($2, (SELECT dm_conversation_id FROM dm_convo_cte)), ($3, (SELECT dm_conversation_id FROM dm_convo_cte))
  )
  SELECT dm_conversation_id FROM dm_convo_cte`,
    values: [
      { type: "direct", author: client.username },
      client.user_id,
      partner_user_id,
    ],
  }

  // return needed details
  return (await dbQuery(query)).rows[0].dm_conversation_id
}

/**
 * @param {object} param0
 * @param {object} param0.conversationInfo
 * @param {"group"} param0.conversationInfo.type
 * @param {string} param0.conversationInfo.title Group title, if `type` is "group"
 * @param {string} param0.conversationInfo.description Group description, if `type` is "group"
 * @param {string} param0.conversationInfo.cover_image_url Group cover image, if `type` is "group"
 * @param {string} param0.conversationInfo.created_by The User that created the group, if `type` is "group"
 * @param {number[]} param0.participantsUserIds
 * @param {object} param0.activity_info
 */
export const createGroupConversation = async ({
  conversationInfo,
  participantsUserIds,
  activity_info,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH group_convo_cte AS (
      INSERT INTO "Conversation" (info) 
      VALUES ($1) 
      RETURNING id AS group_conversation_id
    ), user_convo_cte AS (
      INSERT INTO "UserConversation" (user_id, conversation_id) 
      VALUES ${generateMultiRowInsertValuesParameters({
        rowsCount: participantsUserIds.length,
        columnsCount: 1,
        paramNumFrom: 3,
        // here I just concatenated each user_id column paceholder with conversation_id column value
      }).replace(
        /\$\d/g,
        (m) => `${m}, SELECT group_conversation_id FROM group_convo_cte`
      )}
    ), activity_log AS (
      INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
      VALUES (SELECT group_conversation_id FROM group_convo_cte, $2)
    )
    SELECT `,
    values: [
      conversationInfo,
      activity_info,
      ...participantsUserIds.map((user_id) => user_id),
    ],
  }

  // return needed details
  await dbQuery(query)
}

export const deleteUserConversation = async (client_user_id, conversation_id) => {
  const query = {
    text: `
    UPDATE "UserConversation" 
    SET deleted = true
    WHERE user_id = $1 AND conversation_id = $2`,
    values: [client_user_id, conversation_id]
  }

  await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.client_user_id
 * @param {number} param0.group_conversation_id
 * @param {Object<string, string>} param0.newInfoKVPair
 * @returns {Promise<boolean>}
 */
export const changeGroupInfo = async ({
  client_user_id,
  group_conversation_id,
  newInfoKVPair,
  activity_info,
}) => {
  const [[infoKey, newInfoValue]] = Object.entries(newInfoKVPair)

  const query = {
    text: `
    WITH client_is_group_admin AS (
      SELECT EXISTS(SELECT role 
        FROM "GroupMembership" 
        WHERE group_conversation_id = $1 AND user_id = $2 AND deleted = false AND role = 'admin')
    ), convo_cte AS (
      UPDATE "Conversation" SET info = jsonb_set(info, '{$4}', '$5')
      WHERE conversation_id = $1 AND (SELECT * FROM client_is_group_admin)
    ), activity_log AS (
      IF (SELECT * FROM client_is_group_admin) THEN
        INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
        VALUES ($1, $3)
      END IF
    )
    SLECT exists AS changed FROM client_is_group_admin`,
    values: [
      group_conversation_id,
      client_user_id,
      activity_info,
      infoKey,
      newInfoValue,
    ],
  }

  return (await dbQuery(query)).rows[0].changed
}

/**
 * @param {object} param0
 * @param {number[]} param0.participantsUserIds
 * @param {number} param0.conversation_id
 * @returns {Promise<boolean>}
 */
export const addParticipantsToGroup = async ({
  client_user_id,
  participantsUserIds,
  group_conversation_id,
  activity_info,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH client_is_group_admin AS (
      SELECT EXISTS(SELECT role 
        FROM "GroupMembership" 
        WHERE group_conversation_id = $1 AND user_id = $3 AND deleted = false AND role = 'admin')
    ), user_convo_cte AS (
      IF (SELECT * FROM client_is_group_admin) THEN
        INSERT INTO "UserConversation" (user_id, conversation_id) 
        VALUES ${generateMultiRowInsertValuesParameters({
          rowsCount: participantsUserIds.length,
          columnsCount: 1,
          paramNumFrom: 4,
          // here I just concatenated each user_id column paceholder with conversation_id column value
        }).replace(/\$\d/g, (m) => `${m}, $1`)}
      END IF
    ), activity_log AS (
      IF (SELECT * FROM client_is_group_admin) THEN
        INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
        VALUES ($1, $2)
      END IF
    )
    SELECT exists AS added FROM client_is_group_admin`,
    values: [
      group_conversation_id,
      activity_info,
      client_user_id,
      ...participantsUserIds.map((user_id) => user_id),
    ],
  }

  return (await dbQuery(query)).rows[0].added

  // After this, if conversation type is "group", create group membership is automatically "trigger"ed for each inserted "UserConversation"
  // Afterwards, we programmatically log the activity
}

/**
 * @returns {Promise<boolean>}
 */
export const removeParticipantFromGroup = async ({
  client_user_id,
  participant_user_id,
  group_conversation_id,
  activity_info,
}) => {
  const query = {
    text: `
    WITH client_is_group_admin AS (
      SELECT EXISTS(SELECT role 
        FROM "GroupMembership" 
        WHERE group_conversation_id = $1 AND user_id = $2 AND deleted = false AND role = 'admin')
    ), user_convo_cte AS (
      UPDATE "UserConversation" 
      SET deleted = true
      WHERE conversation_id = $1 AND user_id = $3 AND (SELECT * FROM client_is_group_admin)
    ), activity_log AS (
      IF (SELECT * FROM client_is_group_admin) THEN
        INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
        VALUES ($1, $4)
      END IF
    )
    SLECT exists AS removed FROM client_is_group_admin`,
    values: [
      group_conversation_id,
      client_user_id,
      participant_user_id,
      activity_info,
    ],
  }

  return (await dbQuery(query)).rows[0].removed
}

export const joinGroup = async ({
  participant_user_id,
  group_conversation_id,
  activity_info,
}) => {
  const query = {
    text: `
    WITH user_convo_cte AS (
      INSERT "UserConversation" (user_id, conversation_id)
      VALUES ($1, $2)
    ), activity_log AS (
      INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
      VALUES ($2, $3)
    )`,
    values: [participant_user_id, group_conversation_id, activity_info],
  }

  await dbQuery(query)
}

export const leaveGroup = async ({
  participant_user_id,
  group_conversation_id,
  activity_info,
}) => {
  const query = {
    text: `
    WITH convo_cte AS (
      UPDATE "UserConversation" 
      SET deleted = true
      WHERE user_id = $1 AND conversation_id = $2
    ), activity_log AS (
      INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
      VALUES ($2, $3)
    )`,
    values: [participant_user_id, group_conversation_id, activity_info],
  }

  await dbQuery(query)
}

/**
 *
 * @param {object} param0
 * @param {"admin" | "member"} param0.role
 * @returns {Promise<boolean>}
 */
export const changeGroupParticipantRole = async ({
  client_user_id,
  participant_user_id,
  group_conversation_id,
  activity_info,
  role,
}) => {
  const query = {
    text: `
    WITH client_is_group_admin AS (
      SELECT EXISTS(SELECT role 
        FROM "GroupMembership" 
        WHERE group_conversation_id = $1 AND user_id = $2 AND deleted = false AND role = 'admin')
    ), group_mem_cte AS (
      UPDATE "GroupMembership" 
      SET role = $4
      WHERE group_conversation_id = $1 AND user_id = $3 AND (SELECT * FROM client_is_group_admin)
    ), activity_log AS (
      IF (SELECT * FROM client_is_group_admin) THEN
        INSERT INTO "GroupConversationActivityLog" (group_conversation_id, activity_info)
        VALUES ($1, $5)
      END IF
    )
    SLECT exists AS done FROM client_is_group_admin`,
    values: [
      group_conversation_id,
      client_user_id,
      participant_user_id,
      role,
      activity_info,
    ],
  }

  return (await dbQuery(query)).rows[0].done
}

/**
 * @param {number} client_user_id
 */
export const getAllUserConversations = async (client_user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT conversation_id,
      conversation_type,
      group_title,
      group_cover_image,
      updated_at,
      partner_name,
      partner_profile_pic,
      partner_connection_status,
      partner_last_active,
      unread_messages_count,
      last_history_item
    FROM "UserConversationsListView"
    WHERE client_user_id = $1 AND partner_user_id != $1 AND last_history_item IS NOT NULL
    ORDER BY updated_at DESC
    `,
    values: [client_user_id],
  }

  return stripNulls((await dbQuery(query)).rows)
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
    SELECT * 
    FROM (SELECT * 
      FROM "ConversationHistoryView" 
      WHERE conversation_id = $1 
      ORDER BY created_at DESC 
      LIMIT $2 OFFSET $3)
    ORDER BY created_at ASC
    `,
    values: [conversation_id, limit, offset],
  }

  return stripNulls((await dbQuery(query)).rows)
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
  sender_user_id,
  conversation_id,
  msg_content,
}) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH message_cte AS (
      INSERT INTO "Message" (sender_user_id, conversation_id, msg_content) 
      VALUES ($1, $2, $3)
      RETURNING id AS message_id, sender_user_id, conversation_id, msg_content
    )
    SELECT message_id,
      conversation_id,
      msg_content,
      json_build_object(
        'user_id', "user".id,
        'profile_pic_url', "user".profile_pic_url,
        'username', "user".username,
        'name', "user".name,
        'connection_status', "user".connection_status
      ) AS sender
    FROM message_cte
    INNER JOIN "User" "user" ON "user".id = sender_user_id`,
    values: [sender_user_id, conversation_id, msg_content],
  }

  return (await dbQuery(query)).rows[0]
}

/**
 * The algorithm in this function explains how all `UPDATE` algorithms were implemented dynamically in this app (save a few ones). The documentation was added here as this seems to be the most complex implementation.
 * @param {number} message_id
 * @param {Map<string, any>} updateKVPairs
 */
export const updateDeliveryStatus = () => {}

/**
 * @param {number} message_id
 * @param {number} user_id
 * @returns {Promise<boolean>} Has message delivered to all conversation participants?
 */
export const acknowledgeMessageDelivered = async (user_id, message_id) => {
  // use CTE to prevent illegal operatons
  // prevent a message from being acknowledge twice
  // prevent a user that doesn't belong to conversation
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH already_acked AS (
      SELECT delivered_to @> ARRAY[CAST ($1 AS INTEGER)] 
      FROM "Message" 
      WHERE id = $2
    ), msg AS (
      UPDATE "Message" 
      SET delivered_to = array_append(delivered_to, $1) 
      WHERE id = $2 AND (SELECT * FROM already_acked) = false
      RETURNING (delivery_status = 'delivered') AS is_delivered
    )
    SELECT is_delivered FROM msg`,
    values: [user_id, message_id],
  }
  
  return (await dbQuery(query)).rows[0]?.is_delivered
}

/**
 * @param {number} message_id
 * @param {number} user_id
 * @returns {Promise<boolean>} Has message been read by all conversation participants?
 */
export const acknowledgeMessageRead = async (user_id, message_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    WITH already_acked AS (
      SELECT read_by @> ARRAY[CAST ($1 AS INTEGER)] 
      FROM "Message" 
      WHERE id = $2
    ), msg AS (
      UPDATE "Message" 
      SET read_by = array_append(read_by, $1) 
      WHERE id = $2 AND (SELECT * FROM already_acked) = false
      RETURNING (delivery_status = 'read') AS is_read
    )
    SELECT is_read FROM msg`,
    values: [user_id, message_id],
  }

  return (await dbQuery(query)).rows[0]?.is_read
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
    INSERT INTO "MessageReaction" (message_id, reactor_user_id, reaction_code_point) 
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
    DELETE FROM "MessageReaction" WHERE message_id = $1 AND reactor_user_id = $2`,
    values: [message_id, reactor_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} user_id
 * @param {"online" | "offline"} connection_status
 */
export const updateUserConnectionStatus = async (
  user_id,
  connection_status
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
    INSERT INTO "BlockedUser" (blocking_user_id, blocked_user_id) 
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
    text: `DELETE FROM "BlockedUser" WHERE blocking_user_id = $1 AND blocked_user_id = $2`,
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
    INSERT INTO "ReportedMessage" (message_id, reporter_user_id, reported_user_id, reason) 
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
  INSERT INTO "MessageDeletionLog" (deleter_user_id, message_id, deleted_for) 
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
    text: `
    SELECT "user".id, 
      "user".username, 
      "user".name, 
      "user".profile_pic_url, 
      "user".connection_status,
      "conv".id AS conversation_id
    FROM "User" "user"
    LEFT JOIN "UserConversation" "other_user_conv" 
      ON "other_user_conv".user_id = "user".id
    LEFT JOIN "UserConversation" "client_user_conv" 
      ON "client_user_conv".conversation_id = "other_user_conv".conversation_id AND "client_user_conv".user_id = $2
	  LEFT JOIN "Conversation" "conv" 
      ON "other_user_conv".conversation_id = "conv".id
	  WHERE (username LIKE $1 OR name LIKE $1) AND "user".id != $2 AND ("conv".info ->> 'type' != 'group' OR "conv".info ->> 'type' IS NULL)`,
    values: [`%${search}%`, client_user_id],
  }

  return (await dbQuery(query)).rows
}

/* Helpers */
/**
 * @param {number} user_id
 * @returns {Promise<number[]>}
 */
export const getAllUserConversationIds = async (user_id) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    SELECT conversation_id AS c_id
    FROM "UserConversation"
    WHERE user_id = $1 AND deleted = false
    `,
    values: [user_id],
  }

  return (await dbQuery(query)).rows
}

/* TRIGGERS */
// These are functions automatically triggered after a change is made to the database
