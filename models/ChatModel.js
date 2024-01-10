/**
 * @param {object} param0
 * @param {"individual" | "group"} param0.type
 * @param {object} param0.info
 * @param {"individual" | "group"} param0.info.type
 * @param {string} param0.info.group_title Group title, if `type` is "group"
 * @param {string} param0.info.group_description Group description, if `type` is "group"
 * @param {string} param0.info.group_cover_image_url Group cover image, if `type` is "group"
 * @param {import('pg').PoolClient} dbClient
 */
export const createConversation = async ({ type, info }, dbClient) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "Conversation" (type, info) 
    VALUES ($1, $2)`,
    values: [type, info],
  }

  await dbClient.query(query)
}

// needs a trigger
export const updateConversation = async ({ conversation_id, updateKVPairs }) => {}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.conversation_id
 * @param {import('pg').PoolClient} dbClient
 */
export const createUserConversation = async (
  { user_id, conversation_id },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "UserConversation" (user_id, conversation_id) 
    VALUES ($1, $2)`,
    values: [user_id, conversation_id],
  }

  await dbClient.query(query)
}

export const updateUserConversation = async ({ user_id, updateKVPairs }) => {}
export const deleteUserConversation = async ({ user_id }) => {}
export const getAllUserConversations = async ({ user_id }) => {}

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.group_conversation_id
 * @param {"admin" | "member"} param0.role
 * @param {import('pg').PoolClient} dbClient
 */
export const createGroupMembership = async (
  { user_id, group_conversation_id, role },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "GroupMembership" (user_id, group_conversation_id, role) 
    VALUES ($1, $2, $3)`,
    values: [user_id, group_conversation_id, role],
  }

  await dbClient.query(query)
}
export const updateGroupMembership = async ({ user_id, group_conversation_id, updateKVPairs }) => {}

/**
 * @param {object} param0
 * @param {number} param0.sender_id
 * @param {number} param0.conversation_id
 * @param {object} param0.msg_content
 * @param {"text" | "image" | "voice"} param0.msg_content.type
 * @param {string | null} param0.msg_content.text_content Text content
 * @param {string | null} param0.msg_content.image_data_url Image URL
 * @param {string | null} param0.msg_content.voice_data_url Voice data URL
 * @param {string | null} param0.msg_content.image_description Image description. If `type` is Image
 * @param {object} param0.msg_attachment
 * @param {"audio" | "video" | "location" | "document" | "compressed" | "other"} param0.msg_attachment.type
 * @param {string} param0.msg_attachment.file_type A valid MIME file type
 * @param {string} param0.msg_attachment.file_url File URL
 * @param {GeolocationCoordinates} param0.msg_attachment.location_coordinate A valid geolocation coordinate
 * @param {import('pg').PoolClient} dbClient
 */
export const createMessage = async (
  { sender_id, conversation_id, msg_content, msg_attachment },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "Message" (sender_id, conversation_id, msg_content, msg_attachment) 
    VALUES ($1, $2, $3, $4)`,
    values: [sender_id, conversation_id, msg_content, msg_attachment],
  }

  await dbClient.query(query)
}

export const deleteMessage = async ({ message_id }) => {}
export const updateMessage = async ({ message_id, updateKVPairs }) => {}
export const getAllConversationMessages = ({ conversation_id }) => {}

export const createMessageReaction = async (
  { message_id, reactor_user_id, reaction_code_point },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "MessageReaction" (message_id, reactor_user_id, reaction_code_point) 
    VALUES ($1, $2, $3)`,
    values: [message_id, reactor_user_id, reaction_code_point],
  }

  await dbClient.query(query)
}

export const deleteMessageReaction = async ({
  message_id,
  reactor_user_id,
}) => {}

export const updateUserConnectionStatus = async ({ user_id, updateKVPairs }) => {}

export const createBlockedUser = async (
  { blocking_user_id, blocked_user_id },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
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
 * @param {import('pg').PoolClient} dbClient
 */
export const createReportedMessage = async (
  { message_id, reporter_user_id, reported_user_id, reason },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
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
 * @param {import('pg').PoolClient} dbClient
 */
export const createMessageDeletionLog = async (
  { user_id, message_id, deleted_for },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
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
 * @param {import('pg').PoolClient} dbClient
 */
export const createConversationActivityLog = async (
  { conversation_id, subject_user_id, activity_message },
  dbClient
) => {
  /** @type {import('pg').QueryConfig} */
  const query = {
    text: `
    INSERT INTO "ConversationActivityLog" (conversation_id, subject_user_id, activity_message) 
    VALUES ($1, $2, $3)`,
    values: [conversation_id, subject_user_id, activity_message],
  }

  await dbClient.query(query)
}

export const getUsersForChat = () => {}
