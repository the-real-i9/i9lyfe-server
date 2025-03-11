import cookieSession from "cookie-session"
import dotenv from "dotenv"


dotenv.config()

export const expressSession = () =>
  cookieSession({
    keys: [process.env.COOKIE_SECRET_KEY_1, process.env.COOKIE_SECRET_KEY_2],
    domain: process.env.SERVER_HOST,
  })
