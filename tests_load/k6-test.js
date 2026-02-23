import { sleep } from "k6"
import http from "k6/http"

export const options = {
  vus: 10,
  duration: "30s",
}

export default function () {
  const HOST_URL = "http://localhost:8000"
  const signupPath = HOST_URL + "/api/auth/signup"

  http.get("https://k6.io")

  sleep(1)
}
