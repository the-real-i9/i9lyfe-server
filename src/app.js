import express from "express"
import cors from "cors"

import dotenv from "dotenv"

dotenv.config()

import PrivateRoutes from "./routes/private.routes.js"
import PublicRoutes from "./routes/public.routes.js"


const app = express()

app.use(cors())

app.use(express.json({ limit: "10mb" }))

app.use("/api/private", PrivateRoutes)
app.use("/api/public", PublicRoutes)

export default app
