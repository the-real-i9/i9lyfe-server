import { generateJwt } from "../../utils/helpers.js"


/**
 * @param {import("socket.io").Socket} socket
 */
export const renewJwtToken = (socket) => {
  const { client_user_id, client_username } = socket.jwt_payload

  const newJwtToken = generateJwt({ client_user_id, client_username })

  socket.emit("renewed jwt", newJwtToken)
}
