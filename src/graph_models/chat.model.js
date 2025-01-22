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
      SET clientChat.last_activity_type = "message", 
        partnerChat.last_activity_type = "message",
        clientChat.last_message_at = datetime($created_at), 
        partnerChat.last_message_at = datetime($created_at)
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
      DETACH DELETE clientChat
      `,
      { client_user_id, partner_user_id }
    )
  }

  static async getAll(client_user_id) {
    // I'm here
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id })-[:WITH_USER]->(partnerUser),
        (clientChat)<-[:IN_CHAT]-(lmsg:Message WHERE lmsg.created_at = clientChat.last_message_at),
        (clientChat)<-[:IN_CHAT]-(:Message)<-[lrxn:REACTS_TO_MESSAGE WHERE lrxn.at = clientChat.last_reaction_at]-(reactor)
      WITH clientChat, toString(clientChat.last_message_at) AS last_message_at, partnerUser { .id, .username, .profile_pic_url, .connection_status } AS partner,
        CASE clientChat.last_activity_type 
          WHEN "message" THEN lmsg { type: "message", .content, .delivery_status }
          WHEN "reaction" THEN lrxn { type: "reaction", .reaction, reactor: reactor.username }
        END AS last_activity
      ORDER BY clientChat.last_message_at DESC
      RETURN collect(clientChat { partner, .unread_messages_count, last_message_at, last_activity }) AS my_chats
      `,
      { client_user_id }
    )

    return records[0].get("my_chats")
  }

  static async getHistory({ client_user_id, partner_user_id, limit, offset }) {
    const { records } = await neo4jDriver.executeRead(
      `
      MATCH (clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })<-[:IN_CHAT]-(message:Message)<-[rxn:REACTS_TO_MESSAGE]-(reactor)
      WITH message, toString(message.created_at) AS created_at, collect({ user: reactor { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
      ORDER BY message.created_at DESC
      OFFSET toInteger($offset)
      LIMIT toInteger($limit)
      RETURN collect(message { .*, created_at, reactions }) AS chat_history
      `,
      { client_user_id, partner_user_id, limit, offset }
    )

    return records[0].get("chat_history")
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

  static async ackRead({
    client_user_id,
    partner_user_id,
    message_id,
    read_at,
  }) {
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
    const reaction_at = new Date().toISOString()

    await neo4jDriver.executeWrite(
      `
      MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_user_id: $client_user_id, partner_user_id: $partner_user_id })<-[:IN_CHAT]-(message:Message{ id: $message_id }),
        (clientChat)-[:WITH_USER]->(partnerChat)
      MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
      ON CREATE
        SET crxn.reaction = $reaction, 
          crxn.at = datetime($reaction_at),
          clientChat.last_activity_type = "reaction", 
          partnerChat.last_activity_type = "reaction",
          clientChat.last_reaction_at = datetime($reaction_at),
          partnerChat.last_reaction_at = datetime($reaction_at)
      `,
      {
        client_user_id,
        partner_user_id,
        message_id,
        reaction,
        reaction_at,
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

      return
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
