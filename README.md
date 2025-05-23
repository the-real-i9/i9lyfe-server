# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe is an API server for a Social Media Application, built with Go and Neo4j. It supports major social media application features that can be used to implement a mordern social media application frontend.

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

i9lyfe was initially built using these technologies, subsequent improvements to the app required alternatives.

<div>
<img style="margin-right: 10px" alt="javascript" width="40" src="./.attachments/tech-icons/javascript-original.svg" />
<img style="margin-right: 10px" alt="nodejs" width="40" src="./.attachments/tech-icons/nodejs-original.svg" />
<img style="margin-right: 10px" alt="express" width="40" src="./.attachments/tech-icons/express-original.svg" />
<img style="margin-right: 10px" alt="postgresql" width="40" src="./.attachments/tech-icons/postgresql-original.svg" />
<img style="margin-right: 10px" alt="socket.io" width="40" src="./.attachments/tech-icons/socketio-original.svg" />
</div>

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation)
- [My Content Recommendation Algorithm](#my-content-recommendation-algorithm)

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

For all **REST request/response Communication**: [Click Here](./.apidoc/restapi.md)

For all **WebSocket Real-time Communication**: [Click Here](./.apidoc/websocketsapi.md)

## My Content Recommendation Algorithm

### User Home Feed Population Re-Implementation

- When a user comes online (i.e establishes a WebSocket connection), he joins the list of users listening for recommended posts in a `sync.Map`.
  - When a user goes offline (i.e tears-down a WebSocket connection), he is removed from the list of users listening for recommended posts in a `sync.Map`.

- In the `sync.Map`, each user's `username` is mapped to a`[2]any`, the first item holds the user's socket pipe `*websocket.Conn`, while the second hosts a recommended posts queue `[]string`, where recommended posts are kept (by `id`) until the user pulls them.
  - When the post queue contains at least one post, the user is notified of the availability of new posts.
  - When the user pulls them, we get all posts whose `id` is present in the recommended posts queue, sort them according to time created in descending order, and send them to the user.

- If the user is offline (i.e. its key isn't found in the list of users listening for recommended posts, `sync.Map`), the post is queued in the database instead.
  - When the user comes online, we get all posts whose `id` is present in the user recommended posts list, sort them according to time created in descending order, and send them to the user.

### Expolre Feed Content Recommedation Algorithm

### User Home Feed Content Recommendation Algorithm

We recommend posts to users based on their follow network and their interests.

Determining a user's follow network is easy. Determining their interests requires technicality. That's why this algorithm is about determining user's interests.
