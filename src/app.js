import express from "express"
import cors from "cors"
import helmet from "helmet"

import PrivateRoutes from "./routes/private.routes.js"
import PublicRoutes from "./routes/public.routes.js"


const app = express()

app.use(helmet())

app.use(cors())

app.use(express.json({ limit: "10mb" }))

app.use("/api/auth", )

app.use("/api/app/private", PrivateRoutes)
app.use("/api/app/public", PublicRoutes)

export default app
