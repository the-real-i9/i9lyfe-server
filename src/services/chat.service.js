import { Chat, Message } from "../models/chat.model.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as messageBrokerService from "../services/messageBroker.service.js"

export const getMyChats = async (client_username) => {
  const chats = await Chat.getAll(client_username)

  return {
    data: chats,
  }
}

export const deleteChat = async (client_username, partner_username) => {
  await Chat.delete(client_username, partner_username)

  return {
    data: { msg: "operation successful" },
  }
}

export const getChatHistory = async ({
  client_username,
  partner_username,
  limit,
  offset,
}) => {
  const chatHistory = await Chat.getHistory({
    client_username,
    partner_username,
    limit,
    offset,
  })

  return {
    data: chatHistory,
  }
}

export const sendMessage = async ({
  client_username,
  partner_username,
  msg_content,
  created_at,
}) => {
  const { media_data = null, ...message_content } = msg_content

  message_content.media_url = media_data
    ? await mediaUploadService.upload({
        media_data,
        extension: msg_content.extension,
      })
    : ""

  const message_content_json = JSON.stringify(message_content)

  const { client_res, partner_res } = await Chat.sendMessage({
    client_username,
    partner_username,
    message_content: message_content_json,
    created_at,
  })

  // replace with
  messageBrokerService.sendChatEvent(
    "new message",
    partner_username,
    partner_res
  )

  return {
    data: client_res,
  }
}

export const ackMessageDelivered = async ({
  client_username,
  partner_username,
  message_id,
  delivered_at,
}) => {
  await Message.ackDelivered({
    partner_username,
    client_username,
    message_id,
    delivered_at,
  })

  // to mark message with a double-tick on the partner's side, whose own partner is the client_user
  messageBrokerService.sendChatEvent("message delivered", partner_username, {
    partner_username: client_username,
    delivered_at,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const ackMessageRead = async ({
  client_username,
  partner_username,
  message_id,
  read_at,
}) => {
  await Message.ackRead({
    client_username,
    partner_username,
    message_id,
    read_at,
  })

  messageBrokerService.sendChatEvent("message read", partner_username, {
    partner_username: client_username, // client_user is the partner of partner_user
    message_id,
    read_at,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const reactToMessage = async ({
  client_username,
  partner_username,
  message_id,
  reaction,
}) => {
  await Message.reactTo({
    client_username,
    partner_username,
    message_id,
    reaction,
  })

  messageBrokerService.sendChatEvent("message reaction", partner_username, {
    partner_username: client_username,
    message_id,
    reaction,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const removeReactionToMessage = async ({
  client_username,
  partner_username,
  message_id,
}) => {
  await Message.removeReaction({
    client_username,
    partner_username,
    message_id,
  })

  messageBrokerService.sendChatEvent(
    "message reaction removed",
    partner_username,
    {
      partner_username: client_username,
      message_id,
    }
  )

  return {
    data: { msg: "operation successful" },
  }
}

export const deleteMessage = async ({
  client_username,
  partner_username,
  message_id,
  delete_for,
}) => {
  await Message.delete({
    client_username,
    partner_username,
    message_id,
    delete_for,
  })

  if (delete_for === "everyone") {
    messageBrokerService.sendChatEvent("message deleted", partner_username, {
      partner_username: client_username,
      message_id,
    })
  }

  return {
    data: { msg: "operation successful" },
  }
}
