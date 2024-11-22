import { dbQuery } from "../configs/db.js"

export class Chat {
  static async create({ client_user_id, partner_user_id, init_message }) {
    const query = {
      text: "SELECT client_res, partner_res FROM create_chat($1, $2, $3)",
      values: [client_user_id, partner_user_id, init_message],
    }

    // return needed details
    return (await dbQuery(query)).rows[0]
  }

  static async delete(client_user_id, chat_id) {
    const query = {
      text: `
    UPDATE user_chat
    SET deleted = true
    WHERE user_id = $1 AND chat_id = $2`,
      values: [client_user_id, chat_id],
    }

    await dbQuery(query)
  }

  static async getAll(client_user_id) {
    const query = {
      text: "SELECT * FROM get_user_chats($1)",
      values: [client_user_id],
    }

    return (await dbQuery(query)).rows
  }

  static async getHistory({ chat_id, limit, offset }) {
    const query = {
      text: `
    SELECT * FROM get_chat_history($1, $2, $3)
    `,
      values: [chat_id, limit, offset],
    }

    return (await dbQuery(query)).rows
  }

  static async sendMessage({
    client_user_id,
    chat_id,
    message_content,
  }) {
    const query = {
      text: "SELECT client_res, partner_res FROM create_message($1, $2, $3)",
      values: [chat_id, client_user_id, message_content],
    }

    return (await dbQuery(query)).rows[0]
  }

  static async blockUser(blocking_user_id, blocked_user_id) {
    const query = {
      text: `
    INSERT INTO blocked_user (blocking_user_id, blocked_user_id) 
    VALUES ($1, $2)`,
      values: [blocking_user_id, blocked_user_id],
    }

    await dbQuery(query)
  }

  static async unblockUser(blocking_user_id, blocked_user_id) {
    const query = {
      text: "DELETE FROM blocked_user WHERE blocking_user_id = $1 AND blocked_user_id = $2",
      values: [blocking_user_id, blocked_user_id],
    }

    await dbQuery(query)
  }
}

export class Message {
  static async isDelivered({
    client_user_id,
    chat_id,
    message_id,
    delivery_time,
  }) {
    const query = {
      text: "SELECT ack_msg_delivered($1, $2, $3, $4)",
      values: [client_user_id, chat_id, message_id, delivery_time],
    }

    await dbQuery(query)
  }

  static async isRead({ client_user_id, chat_id, message_id }) {
    const query = {
      text: "SELECT ack_msg_read($1, $2, $3)",
      values: [client_user_id, chat_id, message_id],
    }

    return await dbQuery(query)
  }

  static async reactTo({ message_id, reactor_user_id, reaction_code_point }) {
    const query = {
      text: `
    INSERT INTO message_reaction (message_id, reactor_user_id, reaction_code_point) 
    VALUES ($1, $2, $3)`,
      values: [message_id, reactor_user_id, reaction_code_point],
    }

    await dbQuery(query)
  }

  static async removeReaction(message_id, reactor_user_id) {
    const query = {
      text: `
    DELETE FROM message_reaction WHERE message_id = $1 AND reactor_user_id = $2`,
      values: [message_id, reactor_user_id],
    }

    await dbQuery(query)
  }

  static async report({
    message_id,
    reporter_user_id,
    reported_user_id,
    reason,
  }) {
    const query = {
      text: `
    INSERT INTO reported_message (message_id, reporter_user_id, reported_user_id, reason) 
    VALUES ($1, $2, $3, $4)`,
      values: [message_id, reporter_user_id, reported_user_id, reason],
    }

    await dbQuery(query)
  }

  static async delete({ deleter_user_id, message_id, deleted_for }) {
    const query = {
      text: `
  INSERT INTO message_deletion_log (deleter_user_id, message_id, deleted_for) 
  VALUES ($1, $2, $3)`,
      values: [deleter_user_id, message_id, deleted_for],
    }

    await dbQuery(query)
  }
}
