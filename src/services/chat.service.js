import { Chat, Message } from "../graph_models/chat.model.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import * as messageBrokerService from "../services/messageBroker.service.js"

export const getMyChats = async (client_user_id) => {
  const chats = await Chat.getAll(client_user_id)

  return {
    data: chats,
  }
}

export const deleteChat = async (client_user_id, partner_user_id) => {
  await Chat.delete(client_user_id, partner_user_id)

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
  partner_user_id,
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
    client_user_id,
    partner_user_id,
    message_content: message_content_json,
    created_at,
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
  message_id,
  delivered_at,
}) => {
  await Message.ackDelivered({
    partner_user_id,
    client_user_id,
    message_id,
    delivered_at,
  })

  // to mark message with a double-tick on the partner's side, whose own partner is the client_user
  messageBrokerService.sendChatEvent("message delivered", partner_user_id, {
    partner_user_id: client_user_id,
    delivered_at,
    message_id,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const ackMessageRead = async ({
  client_user_id,
  partner_user_id,
  message_id,
  read_at,
}) => {
  await Message.ackRead({
    client_user_id,
    partner_user_id,
    message_id,
    read_at,
  })

  messageBrokerService.sendChatEvent("message read", partner_user_id, {
    partner_user_id: client_user_id, // client_user is the partner of partner_user
    message_id,
    read_at,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const reactToMessage = async ({
  client,
  partner_user_id,
  message_id,
  reaction,
}) => {
  await Message.reactTo({
    client_user_id: client.user_id,
    partner_user_id,
    message_id,
    reaction,
  })

  messageBrokerService.sendChatEvent("message reaction", partner_user_id, {
    partner: client,
    message_id,
    reaction,
  })

  return {
    data: { msg: "operation sucessful" },
  }
}

export const removeReactionToMessage = async ({
  client,
  partner_user_id,
  message_id,
}) => {
  await Message.removeReaction({
    client_user_id: client.user_id,
    partner_user_id,
    message_id,
  })

  messageBrokerService.sendChatEvent(
    "message reaction removed",
    partner_user_id,
    {
      partner: client,
      message_id,
    }
  )

  return {
    data: { msg: "operation successful" },
  }
}

export const deleteMessage = async ({
  client,
  partner_user_id,
  message_id,
  delete_for,
}) => {
  await Message.delete({
    client_user_id: client.user_id,
    partner_user_id,
    message_id,
    delete_for,
  })

  if (delete_for === "everyone") {
    messageBrokerService.sendChatEvent("message deleted", partner_user_id, {
      partner: client,
      message_id,
    })
  }

  return {
    data: { msg: "operation successful" },
  }
}
