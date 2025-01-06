import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class Chat {
  static async sendMessage({
    client_user_id,
    partner_user_id,
    message_content,
    created_at,
  }) {
    const { records } = await neo4jDriver.executeWrite(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (partnerUser:User{ id: $partner_user_id })
      MERGE (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })-[:WITH_USER]->(partnerUser)
        ON MATCH
          SET updated_at = datetime($created_at)
      MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:Chat{ owner_user_id: $partner_user_id, partner_user_id: $client_user_id })-[:WITH_USER]->(clientUser)
      WITH clientUser, clientChat, partnerUser, partnerChat
      CREATE (message:Message{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: datetime($created_at) }),
        (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat),
        (partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)
      WITH message, clientUser { .id, .username, .profile_pic_url, .connection_status } AS clientUserView
      RETURN { new_msg_id: message.id } AS client_res,
        { message: message { .id, .content }, sender: clientUserView } AS partner_res
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
      MATCH (clientUser:User{ id: $client_user_id }), (toblockUser:User{ id: to_block_user_id })
      CREATE (clientUser)-[:BLOCKS_USER { user_to_user: $user_to_user }]->(toblockUser)
      `,
      {
        client_user_id,
        to_block_user_id,
        user_to_user: `user-${client_user_id}_to_user-${to_block_user_id}`,
      }
    )
  }

  static async unblockUser(client_user_id, blocked_user_id) {
    await neo4jDriver.executeWrite(
      `
      MATCH ()-[br:BLOCKS_USER { user_to_user: $user_to_user }]->()
      DELETE br
      `,
      { user_to_user: `user-${client_user_id}_to_user-${blocked_user_id}` }
    )
  }
}

export class Message {
  static async ackDelivered({
    client_user_id,
    partner_user_id,
    message_id,
    delivery_time,
  }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id }),
        ()-[:RECEIVES_MESSAGE]->(message:Message{ id: $message_id, delivery_status: "sent" })-[:IN_CHAT]->(clientChat)
      SET message.delivery_status = "delivered", clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1, clientChat.updated_at = datetime($delivery_time)
      `,
      { client_user_id, partner_user_id, message_id, delivery_time }
    )
  }

  static async ackRead({ client_user_id, partner_user_id, message_id }) {
    await neo4jDriver.executeWrite(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_chat_id, partner_user_id: $partner_user_id }),
        ()-[:RECEIVES_MESSAGE]->(message:Message{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])-[:IN_CHAT]->(clientChat)
      SET message.delivery_status = "seen", clientChat.unread_messages_count = CASE WHEN coalesce(clientChat.unread_messages_count, 0) <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END
      `,
      { client_user_id, partner_user_id, message_id }
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
      CREATE (clientUser)-[:REACTS_TO_MESSAGE { user_to_message: $user_to_message, reaction: $reaction }]->(message)
      `,
      {
        client_user_id,
        partner_user_id,
        message_id,
        reaction,
        user_to_message: `user-${client_user_id}_to_message-${message_id}`,
      }
    )
  }

  static async removeReaction({ client_user_id, partner_user_id, message_id }) {
    await neo4jDriver.executeWrite(
      `
      MATCH ()-[rr:REACTS_TO_MESSAGE { user_to_message: $user_to_message }]->()-[:IN_CHAT]->(:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })
      DELETE rr
      `,
      {
        client_user_id,
        partner_user_id,
        user_to_message: `user-${client_user_id}_to_message-${message_id}`,
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
