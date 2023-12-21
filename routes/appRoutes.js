import express from "express"
import { postCreationController } from "../controllers/appControllers"

const router = express.Router()

router.post("/create_post", postCreationController)
