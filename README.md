# i9lyfe (API server)

## Intro

i9lyfe-server is an API server for a social media application modelled after Instagram

## Features

- Create post: Post types include *Photo*, *Video*, *Story*, and *Reel*; all accompanied by an optional description. Post may include @mentions and #hashtags.
- Comment on Post/Comment: Comments may contain text, media, or both. Its text content may include @mentions and #hashtags.
  > In this API, I've represented the notion of a *reply* as a *comment-on-comment*, as the former — in my experience — complicates API and database design.
- React to Post/Comment: Reactions are basically, non-surrogate pair, emojis.
- Repost posts. Save posts
- See comments on posts and comments on comments. See users who reacted to posts.
- See your posts. Delete your posts or comments
- See posts and comments you reacted to. See posts you saved
- Follow users. View user profiles
- Edit your profile

---

- Home feed
  - Contents may be fetched in chunks, with the limit and offset request query parameters, to allow UI pagination/infinite-scrolling
  - New posts, relevant to a specific user, are delivered in realtime
  - Post stats including reactions count, comments count, reposts count, and saves count, are updated in realtime

---

- Chat and Messaging
  - Initiate DM chat with user
  - Messages may come in various types including text, voice note, image, video, audio, and file attatchment. image, video, and audio may optionally include a description
  - React to messages and view reactions
  - Realtime chat experience

---
---

The following sections comprise of notable processes and activities in the API and its development. Each process discusses important aspects of the process, some of which are: *Approach*, *Concepts* applied, *Technologies* used, File *Attachments*, and *Tools* used.

Generally, the API uses a RESTful architecture and is built using the NodeJS's Express.js framework.

## Data modelling

### Tools

- **pgAdmin ERD Tool:** Its canvas and UI objects were used to create the database's entity relationship diagram.
  - Its card UI objects were used to represent tables, whose properties iconically identify the table schema, table name, attributes/columns along with their data types, unique key attributes, primary key attributes, and foreign key attributes.
  - Its relationship line UI objects were used to establish relationships between entities (tables); linking the primary key column of one table to the foreign key column of another or the same (circular).
  - I was able to work with its canvas comfortably with the aid of my drawing tablet.
  
### Attachments

[ER diagram](./i9lyfe_ERD.png) - PNG file

[ER diagram](./i9lyfe_ERD.pgerd) - pgAdmin ERD file. Open with pgAdmin.


## Authentication

### Approach

The Signup process involved three steps:

- In the first step, the user submits a new email (which isn't already registered with the API) for verification, after which the API sends a 6-digit verification code to the email
- In the second step, the user submits the 6-digit code they received via email for validation
- In the third and final step, the user provides their registration information

Each next step is dependent on the success of the previous.

### Concepts

- **OTP Auth:** Used in cases where email verification is required before allowing the user to perform a transaction. Use cases in the API include, the signup transaction, and the forgot password transaction.

- **JWT Auth:** We issue a JWT when user signup is successful or when credentials are valid at user signin. Subsequent requests to protected endpoints attach the JWT in the Authorization header for authetication.
  > *Concern 1:* What if a client performs multiple login
  requests to the API and the API issues a JWT for each, shouldn't we invalidate the JWTs previously issued??
  >
  > *Concern 2:* A client logs out the user only by deleting the JWT issued to it. Is that really the best way? Shouldn't we also invalidate the JWT on the backend?
  >
  > *I plan to settle these concerns and update the API accordinly.*

- **Session (Cookie) Auth:** Transactions involving a number of steps or lined-up requests, — like "Signup" and "Password Reset" — need to maintain a session between these requests.

### Technologies

- **jsonwebtoken:** Used for JWT signing

- **express-jwt:** The express jwt middleware that handles JWT authentication for subsequent requests, performing automatic JWT verification and populating `req.auth` with user data.

  Setting the `credentialsRequired` parameter to `false` for public routes allows us to tailor our response data differently for the case where we have an authenticated user and for the case where we do not. This, for example, allows a client to see a user's profile and followers even when they're not logged in. And if they're logged in, they'll be able to see if they're following a user (*the unfollow button will be displayed instead of the follow button*). If they're not logged in, pressing the follow button will redirect them to the login page, as a JWT is not provided in the Authorization header of the POST request.

  ```js
    export const getFollowers = async (req, res) => {
    try {
      const { username } = req.params

      const { limit = 50, offset = 0 } = req.query

      const userFollowers = await UserService.getFollowers({
        username,
        limit,
        offset,
        client_user_id: req.auth?.client_user_id, // client is optional, data will be tailored accordinly
      })

      res.status(200).send(userFollowers)
    } catch (error) {
      console.error(error)
      res.sendStatus(500)
    }
  }
  ```

- **express-session:** The express session middleware that handles session (cookie) management for "Signup" and "Password Reset". The session store **connect-pg** integrates with it to keep session data.

## Database & DB Management

The API uses **PostgreSQL** as its RDBMS.

PostgreSQL Objects used:

- **Tables:** Of course, I'll use tables. But the ones that prove my proficiency with databases are.

- **Views:** I used Views to represent specific UI components, for exmaple, the post card component. The View attributes consists of the properties of the UI component it represents. The "PostView", for example, includes the `reactions_count`, `comments_count`, `reposts_count`, and `saves_count` attributes, among others (these attributes are not calculated with every SELECT query, optimizing SELECT's performance).

- **Types:**

- **Stored Functions:**

- **Full-text Search:** With its `ts_query()` and `ts_vector()` functions, the API supports its searching through all text data contents (usernames, post descriptions, hashtags etc.) for a text, and of course there's the option to retrict your search to a particular set of content.

Notable **DML clauses** used with `SELECT`:

- `GROUP`, `UNION`, `INNER JOIN`, `LEFT JOIN`, `DISTINCT`

Some **Aggregate functions** used:

- `COUNT`, `array_agg()`, and `json_agg()`

Some **JSON functions** used:

- `json_build_object()`

Some **Array Functions** used:

- `array_append()`

### Technologies

- node-postgres (pg):

### Attachments

[SQL Backup file](./i9lyfe_database_backup.sql) - Scan through the schema definitions and object definitions — **function definitions**, in particuler. It also includes sample data for testing.

### Tools

- psql CLI

- pgAdmin

## Request Validation

- express-validator library:

## Realtime communication

- socket-io

## Handling user-generated content

- Google Cloud Storage:

  - Service Account Credential:

  - API Token Credential:

- @google-cloud/storage

## Security knots

### Rate limiting

## Testing

- Jest/Supertest

- Postman

## Deployment

- Google Compute Engine

  - VM instance

  - SSH login

- gcloud CLI

  - gcloud compute scp

  - gcloud compute ssh

- docker CLI

- Docker Hub

- CI/CD: GitHub Actions

  - .yaml script

  - Self Hosted Runner

## API documentation

- Swagger Open API

- API Blueprint

## Version control

- Git & GitHub

## Development tools

- Bash

- Linux

- VSCode

## Support tools

- PostgreSQL Official documentation

- ChatGPT AI

- Microsoft Bing Copilot

- GCP Gemini AI
