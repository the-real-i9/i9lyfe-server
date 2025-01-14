import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class Chat {
  static async sendMessage({
    client_user_id,
    partner_user_id,
    message_content,
    created_at,
  }) {
    created_at = new Date(created_at).toISOString()

    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (partnerUser:User{ id: $partner_user_id })
      MERGE (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })-[:WITH_USER]->(partnerUser)
      MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:Chat{ owner_user_id: $partner_user_id, partner_user_id: $client_user_id })-[:WITH_USER]->(clientUser)
      WITH clientUser, clientChat, partnerUser, partnerChat
      CREATE (message:Message{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: datetime($created_at) }),
        (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat),
        (partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)
      WITH message, toString(message.created_at) AS created_at, clientUser { .id, .username, .profile_pic_url, .connection_status } AS sender
      RETURN { new_msg_id: message.id } AS client_res,
        message { .*, created_at, sender } AS partner_res
      `,
      { client_user_id, partner_user_id, message_content, created_at }
    )

    return records[0].toObject()
  }

  static async delete(client_user_id, partner_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })
      DETACH DELETE clientChat;
      `,
      { client_user_id, partner_user_id }
    )
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

  static async blockUser(client_user_id, to_block_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id })
      MERGE (clientUser)-[:BLOCKS_USER]->(:User{ id: $to_block_user_id })
      `,
      {
        client_user_id,
        to_block_user_id,
      }
    )
  }

  static async unblockUser(client_user_id, blocked_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH (:User{ id: $client_user_id })-[br:BLOCKS_USER]->(:User{ id: $blocked_user_id })
      DELETE br
      `,
      { client_user_id, blocked_user_id }
    )
  }
}

export class Message {
  static async ackDelivered({
    client_user_id,
    partner_user_id,
    message_id,
    delivered_at,
  }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id }),
        ()-[:RECEIVES_MESSAGE]->(message:Message{ id: $message_id, delivery_status: "sent" })-[:IN_CHAT]->(clientChat)
      SET message.delivery_status = "delivered", message.delivered_at = datetime($delivered_at), clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1
      `,
      { client_user_id, partner_user_id, message_id, delivered_at }
    )
  }

  static async ackRead({ client_user_id, partner_user_id, message_id, read_at }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id }),
        ()-[:RECEIVES_MESSAGE]->(message:Message{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])-[:IN_CHAT]->(clientChat)
      WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
      SET message.delivery_status = "seen", message.read_at = datetime($read_at), clientChat.unread_messages_count = unread_messages_count
      `,
      { client_user_id, partner_user_id, message_id, read_at }
    )
  }

  static async reactTo({
    client_user_id,
    partner_user_id,
    message_id,
    reaction,
  }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })<-[:IN_CHAT]-(message:Message{ id: $message_id }),
      MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
      ON CREATE
        SET crxn.reaction = $reaction
        SET crxn.at = datetime()
      `,
      {
        client_user_id,
        partner_user_id,
        message_id,
        reaction,
      }
    )
  }

  static async removeReaction({ client_user_id, partner_user_id, message_id }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (:User{ id: $client_user_id })-[crxn:REACTS_TO_MESSAGE]->(:Message{ id: $message_id })-[:IN_CHAT]->(:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })
      DELETE rr
      `,
      {
        client_user_id,
        partner_user_id,
        message_id,
      }
    )
  }

  static async delete({
    client_user_id,
    partner_user_id,
    message_id,
    delete_for,
  }) {
    if (delete_for === "me") {
      // just remove the message from my client's chat
      await neo4jDriver.executeWrite(
        `
        MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })<-[inr:IN_CHAT]-(message:Message{ id: $message_id })<-[rsmr:SENDS_MESSAGE|RECEIVES_MESSAGE]-(clientUser),
        DELETE inr, rsmr
        `,
        { client_user_id, partner_user_id, message_id }
      )
    }

    // remove the message from both client's and partner's chats
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })<-[:IN_CHAT]-(message:Message{ id: $message_id })
      DETACH DELETE message
      `,
      { client_user_id, partner_user_id, message_id }
    )
  }
}
