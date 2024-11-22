import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import PrivateRoutes from "./routes/private.routes.js"
import PublicRoutes from "./routes/public.routes.js"

dotenv.config()

const app = express()

app.use(cors())

app.use(express.json({ limit: "10mb" }))

app.use("/api/private", PrivateRoutes)
app.use("/api/public", PublicRoutes)

export default app
