# Create Account
## Client | Web Application -- {Container}
- Sends `POST` request to `/api/auth/signup` 
  ```js
  axios.post("/api/auth/signup", { firstname, lastname, username, email, password })
  ```

## Server | API - {Container}
### `signupInputValidator` Middleware -- {Component}
- parse request body
- Check input fields. If validation fails, return error status with appropriate error message:
  ```js
  res
  .status(403)
  .json({ err: { field: "field_name", msg: "e.g. password too short"} })
  ```
- If validation succeeds: call `next()`, which fowards control to the `signupController`

### `signupController` -- {Component}
- parse the request JSON body
---
- Check if the username or email is already registered. If `true`, a `403 Bad request` response accompanied by a message containing the reason for the error, is `return`ed.
---
- hash the password provided by the user for security
- `INSERT` a record in the `User{firstname, lastname, username, email, password}` table containing the data provided by the user.
---
- To establish a session for verification, we use the JWT token technique, that'll be sent in an `Authorization` header in the next request
  ```js
  jwt.sign({ email: "user@gmail.com", usage: "verification" }, process.env.EMAIL_VERIFICATION_JWT_SECRET)
  // This JWT token isn't for authentication, but rather just to carry session data for the next request, an alternative to using a cookie. Have a specific `SECRET CODE` for this purpose.
  ```
- send a `201 Created` response along with the JWT token
  ```js
  res.status(201).json({ jwtToken, msg: "Proceed to email verification" })
  ```
---
- generate a 6-digit time-bound token which expires in 12hrs (may be different)
- `UPSERT` a record in the `pending_verifications{email, verificationToken, tokenExpirationTime}` table.
- construct and send a verification email including the 6-digit token to the user's `email`
  > The above 3 steps should be in a reuseable function since it'll be reused for resending verification email.

# Email Verification
## Client | App
> On the client side, prevent this page from being accessed except via a redirect from `/signup`, or `/signin`.
- User sends a `POST` request to `/api/auth/verify_email` with code and jwtToken in the `Authorization` header.
  ```js
  axios.post('/api/auth/verify_email', { verfToken: "6_digit_token" }, {
    headers: {
      Authorization: jwtToken
    }
  })
  ```

## Server | API
### `emailVerificationController`
- Verify the jwtToken: If verification fails
```js
res.status(403).json{{ msg: "Invalid JWT token" }}
```
- Get the user verification data with the email in the `pending_verifications` table
- Check the provided token against the original token and check the token expiration against current time
- If we have an error, in either case, return a `403 Bad request` along with "invalid token" or "token expired" message respectively
- The user is likely to "Resend code"

# Resend Verification

