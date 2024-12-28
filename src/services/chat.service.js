import { Chat, Message } from "../graph_models/chat.model.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as messageBrokerService from "../services/messageBroker.service.js"

export const createChat = async ({
  partner_user_id,
  client_user_id,
  init_message,
}) => {
  let { media_data, ...init_msg } = init_message

  init_msg.media_url = await mediaUploadService.upload({
    media_data,
    extension: init_msg.extension,
    path_to_dest_folder: `message_medias/user-${client_user_id}`,
  })

  const init_msg_json = JSON.stringify(init_msg)

  const { client_res, partner_res } = await Chat.create({
    client_user_id,
    partner_user_id,
    init_message: init_msg_json,
  })

  messageBrokerService.sendChatEvent("new chat", partner_user_id, partner_res)

  return {
    data: client_res,
  }
}

export const getMyChats = async (client_user_id) => {
  const chats = await Chat.getAll(client_user_id)

  return {
    data: chats,
  }
}

export const deleteChat = async (client_user_id, chat_id) => {
  await Chat.delete(client_user_id, chat_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const getChatHistory = async ({ chat_id, limit, offset }) => {
  const chatHistory = await Chat.getHistory({
    chat_id,
    limit,
    offset,
  })

  return {
    data: chatHistory,
  }
}

export const sendMessage = async ({
  client_user_id,
  chat_id,
  partner_user_id,
  msg_content,
}) => {
  const { media_data, ...message_content } = msg_content

  message_content.media_url = await mediaUploadService.upload({
    media_data,
    extension: msg_content.extension,
  })

  const message_content_json = JSON.stringify(message_content)

  const { client_res, partner_res } = await Chat.sendMessage({
    client_user_id,
    chat_id,
    message_content: message_content_json,
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
  chat_id,
  message_id,
  delivery_time,
}) => {
  await Message.isDelivered({
    client_user_id,
    chat_id,
    message_id,
    delivery_time,
  })

  messageBrokerService.sendChatEvent("message delivered", partner_user_id, {
    chat_id,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const ackMessageRead = async ({
  client_user_id,
  partner_user_id,
  chat_id,
  message_id,
}) => {
  await Message.isRead({
    client_user_id,
    chat_id,
    message_id,
  })

  messageBrokerService.sendChatEvent("message read", partner_user_id, {
    chat_id,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const reactToMessage = async ({
  client,
  chat_id,
  partner_user_id,
  message_id,
  reaction,
}) => {
  const reaction_code_point = reaction.codePointAt()

  await Message.reactTo({
    client_user_id: client.user_id,
    chat_id,
    message_id,
    reaction_code_point,
  })

  messageBrokerService.sendChatEvent("message reaction", partner_user_id, {
    chat_id,
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
  chat_id,
  partner_user_id,
  message_id,
}) => {
  await Message.removeReaction(message_id, client.user_id)

  messageBrokerService.sendChatEvent(
    "message reaction removed",
    partner_user_id,
    {
      reactor: client,
      chat_id,
      message_id,
    }
  )

  return {
    data: { msg: "operation successful" },
  }
}

export const deleteMessage = async ({
  client,
  chat_id,
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
      chat_id,
      deleter_username: client.username,
      message_id,
    })
  }

  return {
    data: { msg: "operation successful" },
  }
}
