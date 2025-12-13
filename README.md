# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe is a full-fledged social media API server built in Go. It supports all of the major social media application features with a scalable, production-grade arcitecture, serving as a foundation for building apps like Instagram, TikTok, or Facebook clones.

## Technologies and Tools

<div style="display: flex; align-items: center;">
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="gofiber" width="40" src="./.attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="postgresql" width="40" src="./.attachments/tech-icons/postgresql-original.svg" />
<img style="margin-right: 10px" alt="redis" width="40" src="./.attachments/tech-icons/redis-original.svg" />
<img style="margin-right: 10px" alt="websocket" width="40" src="./.attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="neo4j" width="40" src="./.attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="jwt" width="40" src="./.attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="googlecloud" width="40" src="./.attachments/tech-icons/googlecloud-original.svg" />
<img style="margin-right: 10px" alt="docker" width="40" src="./.attachments/tech-icons/docker-plain.svg" />
</div>


### Technologies
- **Go** - Programming Language
- **Fiber** - REST API Framework
- **PostgreSQL** - Relational DBMS
- **SQL** - Structured Query Language for Relational Databases
- **PL/pgSQL** - Procedural Language for Database Programming
- **Neo4j** - Graph DBMS
- **CypherQL** - Query Language for a Graph database
- **WebSocket** - Full-duplex, Bi-directional communication protocol
- **Redis Key/Value Store** (Cache)
- **Redis Streams**
- **Redis Pub/Sub**
- **Redis Queue** (via LPOP, RPUSH, RPOP, LPUSH)
- **Google Cloud Storage**

### Tools
- Docker
- Ubuntu Linux
- VSCode
- Git & GitHub Version Control
- GitHub Actions CI

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Upcoming features](#upcoming-features)
- [API Documentation](#api-documentation)

## Features

The following are the features supported by this API. *Visit the API documentation for implementation guides.*

### Content Creation & Sharing

**Create posts** of various types (inspired by Instagram) including *Photo*, *Video*, and *Reel*.
  - Mention users
  - Include hashtags

### User Engagement with Posts and Comments

- **React** to Posts and Comments
- **Comment** on Posts and Comments (replies).
- **Repost:** Re-share posts on your feed.  
- **Access Interactions:** Access comments on posts and replies to comments, and access the list of users who have reacted to a post or comment.
- **Save** posts for later

### User Profile

- Edit your profile.
- Manage your posts.
- Access to saved posts.  
- Access to contents you've engaged with through likes and comments.
- Access to contents you we're mentioned in.
- Access to other user profiles

### Networking

- Follow or unfollow users.

### Chatting and Messaging

- Realtime chatting with users of the application.
- Supports various message types including:
  - Text and voice message
  - Images and videos with caption
  - Audio
  - File attachments
- Realtime user presence status and last seen.


### Realtime Notification and Message Delivery

- Users receive likes, comments, reposts, and mentions notifications relating to them in realtime.
- Chat messages are delivered to target users in realtime.

### Real-Time Updates

- New posts relevant to users are delivered to them in realtime.
- Individual posts receive real-time interaction updates (upon client subscription).
- Clients receive user "presence" and "last seen" updates (upon subscription)

## Upcoming features

The following is a list of features to be supported by this Social Media Backend API.

### Media Processing and Upload Pipeline (In progress)

- Image and Video processing and upload pipeline for user-generated media and post content and chat messages.

### Following Interests

- Users will be able to follow interests (or topics). Content will also be recommended to them based on the interests they follow.

### Search: Full-text Search | Results Ranking | Fuzzy Matching

- Full-text search through content (photos, videos, and reels)
- Search user accounts
- Search hashtaged posts

### Content Recommendation System | User Feed

- A complex recommendation system that pushes relevant content to the user's feed, based on:
  - User following network
  - User's interest followed
  - User engagement stats
  - and more...

### User Follow Recommendation

- App will intelligently recommend users to follow, based on your follow network and content interaction stats.

## Feature design and architecture

<details>
<summary>Post creation</summary>

  * fslek
  
</details>

## API Documentation

REST request/response Communication: [Click Here](./docs/swagger.json)

WebSocket Real-time Communication: [Click Here](./docs/websocketsapi.md)
