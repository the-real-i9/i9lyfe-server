import { dbQuery } from "../configs/db.js"
import { neo4jDriver } from "../configs/graph_db.js"

export class Chat {
  static async create({ client_user_id, partner_user_id, init_message }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id }), (partnerUser:User{ id: $partner_user_id })
      MERGE (clientChat:Chat{ id: $cli_to_par, unread_messages_count: 0, updated_at: datetime() }), 
        (partnerChat:Chat{ id: $par_to_cli, unread_messages_count: 0, updated_at: datetime() })
      CREATE
        (message:Message{ id: randomUUID(), msg_content: $init_message, delivery_status: "sent", created_at: datetime() }),
        (clientUser)-[:HAS_CHAT]->(clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),
        (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat),
        (partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)
      WITH clientChat.id AS ccid, partnerChat.id AS pcid, message, clientUser { .id, .username, .profile_pic_url, .connection_status } AS clientUserView, partnerUser { .id, .username, .profile_pic_url, .connection_status } AS partnerUserView
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
    await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(chat:Chat{ id: $chat_id })
      DETACH DELETE chat;
      `,
      { client_user_id, chat_id }
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

  static async sendMessage({
    client_user_id,
    client_chat_id,
    message_content,
  }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      CREATE (message:Message{ id: randomUUID(), msg_content: $message_content, delivery_status: "sent", created_at: datetime() }),
        (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat),
        (partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)
      WITH partnerChat.id AS pcid, message.id AS msgid, clientUser { .id, .username, .profile_pic_url, .connection_status } AS clientUserView, partnerUser.id AS puid
      RETURN { new_msg_id: msgid } AS client_res,
        { chat_id: pcid, new_message: message { .id, msg_content, sender: clientUserView } } AS partner_res,
        puid AS partner_user_id
      `,
      { client_user_id, client_chat_id, message_content }
    )

    return records[0].toObject()
  }

  static async blockUser(client_user_id, to_block_user_id) {
    await neo4jDriver.executeQuery(
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
    await neo4jDriver.executeQuery(
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
    client_chat_id,
    message_id,
    delivery_time,
  }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })<-[:IN_CHAT]-(message:Message{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-(clientUser),
        (clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      SET message.delivery_status = "delivered", clientChat.unread_messages_count = clientChat.unread_messages_count + 1, clientChat.updated_at = datetime($delivery_time)
      RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
      `,
      { client_user_id, client_chat_id, message_id, delivery_time }
    )

    return records[0].toObject()
  }

  static async ackRead({ client_user_id, client_chat_id, message_id }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })<-[:IN_CHAT]-(message:Message{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-(clientUser),
        (clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      SET message.delivery_status = "seen", clientChat.unread_messages_count = clientChat.unread_messages_count - 1
      RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
      `,
      { client_user_id, client_chat_id, message_id }
    )

    return records[0].toObject()
  }

  static async reactTo({
    client_user_id,
    client_chat_id,
    message_id,
    reaction_code_point,
  }) {
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })<-[:IN_CHAT]-(message:Message{ id: $message_id }),
        (clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      CREATE (clientUser)-[:REACTS_TO_MESSAGE { user_to_message: $user_to_message, reaction_code_point: $reaction_code_point }]->(message)
      RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
      `,
      {
        client_user_id,
        client_chat_id,
        message_id,
        reaction_code_point,
        user_to_message: `user-${client_user_id}_to_message-${message_id}`,
      }
    )

    return records[0].toObject()
  }

  static async removeReaction({ client_user_id, client_chat_id, message_id }) {
    await neo4jDriver.executeQuery(
      `
      MATCH ()-[rr:REACTS_TO_MESSAGE { user_to_message: $user_to_message }]->()-[:IN_CHAT]->(clientChat:Chat{ id: $client_chat_id }),
        (clientUser)-[:HAS_CHAT]->(clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      DELETE rr
      RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
      `,
      {
        client_chat_id,
        user_to_message: `user-${client_user_id}_to_message-${message_id}`,
      }
    )
  }

  static async delete({
    client_user_id,
    client_chat_id,
    message_id,
    delete_for,
  }) {
    if (delete_for === "me") {
      // just remove the message from my client's chat
      const { records } = await neo4jDriver.executeQuery(
        `
        MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })<-[inr:IN_CHAT]-(message:Message{ id: $message_id })<-[rsmr:SENDS_MESSAGE|RECEIVES_MESSAGE]-(clientUser),
          (clientChat)-[:WITH_USER]->(partnerUser),
          (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
        DELETE inr, rsmr
        RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
        `,
        { client_user_id, client_chat_id, message_id }
      )

      return records[0].toObject()
    }

    // remove the message from both client's and partner's chats
    const { records } = await neo4jDriver.executeQuery(
      `
      MATCH (clientUser:User{ id: $client_user_id })-[:HAS_CHAT]->(clientChat:Chat{ id: $client_chat_id })<-[:IN_CHAT]-(message:Message{ id: $message_id })
        (clientChat)-[:WITH_USER]->(partnerUser),
        (partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
      DETACH DELETE message
      RETURN partnerUser.id AS partner_user_id, partnerChat.id AS partner_chat_id
      `,
      { client_user_id, client_chat_id, message_id }
    )

    return records[0].toObject()

  }
}
