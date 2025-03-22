# API Documentation

This document provides a comprehensive usage guide for this API.

This API follows the REST architecture, and HTTP and WebSockets (Socket.io) for communication

## Table of Contents

- [HTTTP Communication](#http-communication)
- [WebSocket(Socket.io) Communication](#websocketsocketio-communication)

## HTTP Communication

| Action | Endpoint | Request | Response |
| - | - | - | - |
| **Signup: Step 1** of 3: *Request New Account* | `/api/auth/signup/request_new_account` | Method: `POST`<br><br>Body:<br>- `email`: A valid email address | Status: `200 OK`<br><br>Headers:<br>- `Set-Cookie`: `...`<br><br>Body:<br>- `msg`: Enter the 6-digit code sent to `${email}` to verify your email. <hr> Status: `400 BadRequest`<br><br>Body:<br>- `msg`: A user with this email already exists. |
| **Signup: Step 2** of 3: *Verify your Email* | `/api/auth/signup/verify_email` | Method: `POST`<br><br>Headers:<br>- `Cookie`: `...`<br><br>Body:<br>- `code`: The 6-digit code sent to your email. | Status: `200 OK`<br><br>Headers:<br>- `Set-Cookie`: `...`<br><br>Body:<br>- `msg`: Your `${email}` has been verified. <hr> Status: `400 BadRequest`<br><br>Body:<br>- `msg`: Incorrect verification code! Check or Re-submit your email. <hr> Status: `400 BadRequest`<br><br>Body:<br>- `msg`: Verification code expired! Re-submit your email. |
| **Signup: Step 3** of 3: *Register User* | `/api/auth/signup/register_user` | Method: `POST`<br><br>Headers:<br>- `Cookie`: `...`<br><br>Body:<br>- `username`: Minimum of 3 alphanumeric characters. | |

---
---

## WebSocket(Socket.io) Communication

**WebSocket Endpoint:** $HOST_DOMAIN/socket.io

### Server Events

### Client Events
