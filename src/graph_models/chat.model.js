import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class Chat {
  static async create({ client_user_id, partner_user_id, init_message }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (partnerUser:User{ id: $partner_user_id })
      MERGE (clientChat:Chat{ id: $cli_to_par, updated_at: datetime() }), 
        (partnerChat:Chat{ id: $par_to_cli, updated_at: datetime() }),
        (message:Message{ id: randomUUID(), msg_content: $init_message, delivery_status: "sent", created_at: datetime() }),
        (clientUser)-[:HAS_CHAT]->(clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),
        (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat)-[:TO_USER]->(partnerUser),
        (partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)-[:FROM_USER]->(clientUser)
      WITH clientChat.id AS ccid, partnerChat.id AS pcid, message, clientUser { .id, .username, .profile_pic_url } AS clientUserView, partnerUser { .id, .username, .profile_pic_url } AS partnerUserView
      RETURN { chat: { id: ccid, partner: partnerUserView }, init_message: message { .*, sender: clientUserView } } AS client_res,
        { chat: { id: pcid, partner: clientUserView }, init_message: message { .*, sender: clientUserView } } AS partner_res,
      `,
      {
        client_user_id,
        partner_user_id,
        init_message,
        cli_to_par: `${client_user_id}_${partner_user_id}`,
        par_to_cli: `${partner_user_id}_${client_user_id}`,
      }
    )

    return records[0].toObject()
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

  static async sendMessage({ client_user_id, chat_id, message_content }) {
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
