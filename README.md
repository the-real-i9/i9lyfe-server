# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe is an API server for a Social Media Application, built with Go and Neo4j. It supports major social media application features that can be used to implement a mordern social media application frontend.

### Open to suggestions

If you need a feature that this API server currently doesn't support, feel free to suggest them in the issues and it will be added as soon as possible.

## Technologies

<div style="display: flex; align-items: center;">
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="neo4j" width="40" src="./.attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="websocket" width="40" src="./.attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="apachekafka" width="40" src="./.attachments/tech-icons/apachekafka-original.svg" />
<img style="margin-right: 10px" alt="javascript" width="40" src="./.attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="javascript" width="40" src="./.attachments/tech-icons/express-validator.svg" />
<img style="margin-right: 10px" alt="googlecloud" width="40" src="./.attachments/tech-icons/googlecloud-original.svg" />
<img style="margin-right: 10px" alt="docker" width="40" src="./.attachments/tech-icons/docker-plain.svg" />
</div>

### Old Technologies

i9lyfe was first built using these technologies, subsequent improvements to the app required alternatives.

<div>
<img style="margin-right: 10px" alt="javascript" width="40" src="./.attachments/tech-icons/javascript-original.svg" />
<img style="margin-right: 10px" alt="nodejs" width="40" src="./.attachments/tech-icons/nodejs-original.svg" />
<img style="margin-right: 10px" alt="express" width="40" src="./.attachments/tech-icons/express-original.svg" />
<img style="margin-right: 10px; alt="postgresql" width="40" src="./.attachments/tech-icons/postgresql-original.svg" />
<img style="margin-right: 10px" alt="socket.io" width="40" src="./.attachments/tech-icons/socketio-original.svg" />
</div>

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation)
- [ToDo List](#todo-list)

## Features

The following is a summary of the features supported by this API. Visit the API documentation to see the full features and their implementation details.

### Content Creation & Sharing

- **Create Post:** Create post of different types (inspired by Instagram) including *Photo*, *Video*, and *Reel*.
  - Mention users
  - Add hashtags

### Interactivity

- **Comment on Posts and Comments:** Write comments on posts and replies to comments.

- **React to Posts and Comments:** Reactions are represented as non-surrogate pair emojis.  
- **Repost:** Share posts on your feed.  
- **View Interactions:** Access comments on posts and replies to comments. View the list of users who have reacted to a post or comment.
- Save posts

### User Profile

- Edit your profile information.
- Manage your posts.
- View saved posts.  
- Access posts you’ve reacted to, posts you've commented on, posts you were mentioned in and more.
- View user profiles

### Networking

- Follow or unfollow users.

### Content Recommendation

#### Home Feed

Get posts feed based on your interests and your follow network *(Work in progress)*.

#### Explore

Explore different post types based on a content recommendation algorithm *(Work in progress)*.

#### Reels

Yes, just like you're thinking. Swipe, swipe and swipe up through an exhautsing list of reel videos.

### Chatting and Messaging

- Chat with users in the application
- Supports message types including:
  - Text and voice message
  - Images, videos, audio files, and file attachments (with description).

### Search

- **Search and Filter:** Find users, hashtags, posts (photo, video, reel), or topics of interest matching a search term.
- **Hashtags:** View top posts with specific hashtags.

### Notifications

- **Activity Updates:** Receive notifications of likes, comments, reposts, and mentions.

### Real-Time Updates

- New posts, relevant to users can be delivered to them in realtime.
- Individual posts can receive real-time interaction updates.

## API Documentation

> Note: The documentation is currently incomplete. Writing it is a cumbersome task, but I'll be done very soon.

For all **HTTP request/response communication**: [Here](./.apidoc/openapi.json)'s a well-written OpenAPI JSON document. Drop or Import it into a [Swagger Editor](https://editor.swagger.io/?_gl=1*1numedn*_gcl_au*MTUxNDUxNjEuMTc0MjY1MTg5Nw..) to access it.

For all **WebSocket real-time communication**: [Here](./.apidoc/websocketsapi.md)'s a written markdown document.

## ToDo List
