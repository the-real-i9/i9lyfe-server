# API Documentation

This API uses the REST API architecture. It is detailed, and easy to implement.

## Table of Contents

- [HTTTP Request/Response](#http-requestresponse)
- [WebSocket Events](#websocket-events)

## HTTP Request/Response

| Action | Endpoint | Description | Request | Response |
| - | - | - | - | - |
| **Signup: Step 1** of 3: *Request New Account* | `POST /api/public/auth/signup/request_new_account` | Allows you to Request new Account | | |
| **Signup: Step 2** of 3: *Verify your Email* | `POST /api/public/auth/signup/verify_email` | Allows you to Verify your email | | |
| **Signup: Step 3** of 3: *Register User* | `POST /api/public/auth/signup/register_user` | Allows | | |

---
---

## WebSocket Events

**WebSocket Endpoint:** $HOST_DOMAIN

### Server Events

### Client Events
