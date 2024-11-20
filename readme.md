# i9lyfe (API Server)

## Overview

i9lyfe is an API server for a social media platform, designed as a portfolio project to showcase my backend development skills. It is aimed at fellow backend engineers and hiring managers seeking highly skilled developers.

## Table of Contents

- [Overview](#overview)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Diagrams](#diagrams)
- [Technologies Used](#technologies-used)
- [Code examples (Code explained)](#code-examples-code-exaplained)

## Features

Coming up with a completely new social media platform idea is challenging, and deciding which features to include or exclude is just as difficult. So, I chose to model my platform after a popular one—Instagram. While it may not yet include all of Instagram's features, my goal is to eventually incorporate them. This approach presents more problems to solve, making the journey even more rewarding.

### Content Creation & Sharing

- **Create Post:** Supports post types such as *Photo*, *Video*, *Story*, and *Reel*. Each post can include a description with optional @mentions and #hashtags.

### Interactivity

- **Comment on Posts or Comments:** Comments can include text, media, or both, with optional @mentions and #hashtags.  
  > In this API, the concept of a *reply* is modeled as a *comment-on-comment* to simplify API and database design.  
- **React to Posts or Comments:** Reactions are represented as non-surrogate pair emojis.  
- **Repost:** Share posts on your feed.  
- **View Interactions:** Access comments on posts and replies to comments. View a list of users who reacted to a post or comment.

### User Profile

- Edit your profile information.  
- Delete your contents.
- View saved posts.  
- Access posts you’ve reacted to or commented on.

### Networking

- Follow or unfollow other users.  
- Explore content suggestions and recommendations tailored to your interests.  
- View user profiles and their associated content.

### Home Feed & Explore

- **Infinite Scrolling:** Seamless pagination for browsing.  
- **Real-Time Updates:** New, relevant posts and interactions, including reactions, comments, reposts, and saves, are delivered and updated in real-time.

### Messaging

- **Direct Messaging (DM):** Initiate private chats with other users.  
- **Message Types:** Supports text, voice notes, images, videos, audio files, and file attachments. Media files (images, videos, and audio) can include optional descriptions.  
- **Reactions:** React to messages and view reactions.  
- **Real-Time Experience:** Ensures seamless and instant communication.

### Discovery & Search

- **Search and Filter:** Find accounts, hashtags, content, or topics of interest.  
- **Hashtags:** Categorize and organize posts under themes for easy discovery.

### Content Curation

- Save specific posts for future reference.  
- Retrieve and view your saved posts.

### Notifications

- **Activity Updates:** Receive alerts about likes, comments, shares, and mentions.

## Diagrams

- [ER diagram](./attachments/i9lyfe_ERD.png) - Created using the pgAdmin ERD tool.
- [Architectural Diagram](./attachments/i9lyfe_ARCHD.png) - A component-level diagram based on the c4 model. The API itself is a c4 "container" type interacting with other container types such as Databases, and Message queues. Created using Draw.io
- [Sequence Diagrams] - showing the flow of operations in the API for each endpoint accessed *(Coming soon...)*

## Technologies Used

### Core

- **Language/Runtime:** JavaScript/Node.js
- **Database System:** PostgreSQL.
  - Database Driver: node-postgres
- **Blob Storage:** Google Cloud Storage
- **Realtime Communication:** WebSockets
- **Messaging System:** Apache Kafka

#### PostgreSQL (Features used)

- Objects: Tables, Views, Sequences, Stored Functions, and Types.
- Full-text Search
- Cursor-based data fetching

### Frameworks & Libraries

- API Framework
  - Express.js
- Authentication
  - express-jwt:
  - jsonwebtoken:
- Session Management
  - express-session
  - connect-pg-simple
- E-Mailing
  - nodemailer
- Validation
  - express-validator
- Database Driver
  - node-postgres (pg)
- Blob storage
  - @google-cloud/storage
  - file-type
- Realtime Communication
  - socket.io
- Event Streaming
  - kafka-node
- Security
  - cors
  - bcrypt
- Environment variables
  - dotenv
- Testing
  - jest+supertest

### Tools

- Database Management
  - pgAdmin
  - psql
  - pg_dump, createdb, dropdp, pg_restore
- Functional Testing
  - Postman
  - Hoppscotch.io
- Cloud Platform
  - Google Cloud Platform (GCP)
- Cloud Platform Management
  - gcloud CLI
  - Google Cloud Console
- Version Control
  - Git & GitHub
  - Github Desktop
  - VSCode's "Source Control" Feature
- Deployment
  - GitHub Actions
  - Docker
  - Google Compute Engine
  - SSH
  - gcloud compute scp
- API Documentation
  - Open API Specification
  - API Blueprint
- Workflow Speed-up
  - OpenAI's ChatGPT
  - GCP's integrated AI, Gemini
  - VSCode Extensions
  - Microsoft Bing Copilot
- Development
  - VSCode
  - Ubuntu Linux
  - Bash Script

#### Google Cloud Platform (Features used)

- **APIs & Services:** Google Compute Engine (GCE), Google Cloud Storage (GCS)
- **Credentials:** Service Accounts, API Tokens, Application default credentials, Workload Identity Federation

## Code Examples (Code Exaplained)

The following are code examples with explanations of notable functionalities and solutions.

### Problem: User creates a post

#### Solution

The API server handles this `POST` request on this endpoint: `/api/post_comment/new_post`

Before handling this request, we validate the request body using the express-validator middleware:

```js
export const createNewPost = [
  checkExact(
    checkSchema(
      {
        medias_data_list: {
          isArray: {
            options: { min: 1 },
            errorMessage: "value must be an array of at least one item",
          },
        },
        "media_data_list.*": {
          isArray: {
            options: { min: 1, max: 10 * 1024 ** 2 },
            errorMessage:
              "item must be an array of uint8 integers containing a maximum of 10mb",
          },
        },
        type: {
          notEmpty: true,
          isIn: {
            options: [["photo", "video", "reel", "story"]],
            errorMessage: "invalid post type",
          },
        },
        description: {
          optional: true,
          notEmpty: true,
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
```

The `checkExact` function validates that the fields provided in the are no more than the ones specified.

- The `media_data_list` field is a list of media items of a specific type, according to post's `type`, selected for the post. `checkSchema` validates that at least one media item is provided.
- The `media_data_list.*` field is any media item in the `media_data_list` array, represented as a binary data&#x2014;specifically an array of unsigned 8-bit integers. `checkSchema` validates that each media data size is no more than 10mb.
- The `type` field is the type of post selected, which must be one of photo, video, story, or reel. `checkSchema` validates that the field is one of these types.
- The `description` field is the optional description text associated with this post. `checkSchema` validates that it is either not provided or must not be an empty string.
