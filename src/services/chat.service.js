import { Conversation, Message } from "../models/chat.model.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as messageBrokerService from "../services/messageBroker.service.js"

export const createConversation = async ({
  partner_user_id,
  client_user_id,
  init_message,
}) => {
  const { media_data, ...init_msg } = init_message

  init_msg.media_url = await mediaUploadService.upload({
    media_data,
    extension: init_msg.extension,
    pathToDestFolder: `message_medias/user-${client_user_id}`,
  })

  const { client_res, partner_res } = await Conversation.create({
    client_user_id,
    partner_user_id,
    init_message: init_msg,
  })

  messageBrokerService.sendChatEvent(
    "new conversation",
    partner_user_id,
    partner_res
  )

  return {
    data: client_res,
  }
}

export const getMyConversations = async (client_user_id) => {
  const conversations = await Conversation.getAll(client_user_id)

  return {
    data: conversations,
  }
}

export const deleteConversation = async (client_user_id, conversation_id) => {
  await Conversation.delete(client_user_id, conversation_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const getConversationHistory = async ({
  conversation_id,
  limit,
  offset,
}) => {
  const conversationHistory = await Conversation.getHistory({
    conversation_id,
    limit,
    offset,
  })

  return {
    data: conversationHistory,
  }
}

export const sendMessage = async ({
  client_user_id,
  conversation_id,
  partner_user_id,
  msg_content,
}) => {
  const { media_data, ...message_content } = msg_content

  message_content.media_url = await mediaUploadService.upload({
    media_data,
    extension: msg_content.extension,
  })

  const { client_res, partner_res } = await Conversation.sendMessage({
    client_user_id,
    conversation_id,
    message_content,
  })

  // replace with
  messageBrokerService.sendChatEvent(
    "new message",
    partner_user_id,
    partner_res
  )

  return {
    data: client_res,
  }
}

export const ackMessageDelivered = async ({
  client_user_id,
  partner_user_id,
  conversation_id,
  message_id,
  delivery_time,
}) => {
  await Message.isDelivered({
    client_user_id,
    conversation_id,
    message_id,
    delivery_time,
  })

  messageBrokerService.sendChatEvent("message delivered", partner_user_id, {
    conversation_id,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const ackMessageRead = async ({
  client_user_id,
  partner_user_id,
  conversation_id,
  message_id,
}) => {
  await Message.isRead({
    client_user_id,
    conversation_id,
    message_id,
  })

  messageBrokerService.sendChatEvent("message read", partner_user_id, {
    conversation_id,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const reactToMessage = async ({
  client,
  conversation_id,
  partner_user_id,
  message_id,
  reaction,
}) => {
  const reaction_code_point = reaction.codePointAt()

  await Message.reactTo({
    reactor_user_id: client.user_id,
    message_id,
    reaction_code_point,
  })

  messageBrokerService.sendChatEvent("message reaction", partner_user_id, {
    conversation_id,
    reactor: client,
    message_id,
    reaction_code_point,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const removeReactionToMessage = async ({
  client,
  conversation_id,
  partner_user_id,
  message_id,
}) => {
  await Message.removeReaction(message_id, client.user_id)

  messageBrokerService.sendChatEvent(
    "message reaction removed",
    partner_user_id,
    {
      reactor: client,
      conversation_id,
      message_id,
    }
  )

  return {
    data: { msg: "operation successful" },
  }
}

export const deleteMessage = async ({
  client,
  conversation_id,
  partner_user_id,
  message_id,
  delete_for,
}) => {
  await Message.delete({
    deleter_user_id: client.user_id,
    message_id,
    deleted_for: delete_for,
  })

  if (delete_for === "everyone") {
    messageBrokerService.sendChatEvent("message deleted", partner_user_id, {
      conversation_id,
      deleter_username: client.username,
      message_id,
    })
  }

  return {
    data: { msg: "operation successful" },
  }
}
