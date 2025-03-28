# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe is an API server for a Social Media Application, built with Node.js and Neo4j. It supports major social media application features that can be used to implement a mordern social media application frontend.

### Target Audience

#### Frontend Developers

If you're a frontend developer looking to build a Social Media Application, not just to have it static UI, but to also make it function.

The API documentation provides a detailed usage guide, following the OpenAPI specification.

#### HRMs, Startup Founders, Project Teams, Hiring Managers etc

If you're in need of a passionate, highly-skilled, expert-level backend engineer/developer, you've found the right one for the job.

The codebase is easily accessible, as it follows a (Routes)-->(Controllers)-->(Services)-->((Services) | (Model)) pattern.

### Open to suggestions

If you need a feature that this API server currently doesn't support, feel free to suggest them in the issues and it will be added as soon as possible.

## Technologies

<div style="display: flex; align-items: center;">
<img style="margin-right: 10px" alt="nodejs" width="50" src="./attachments/tech-icons/nodejs-original.svg" />
<img style="margin-right: 10px" alt="express" width="50" src="./attachments/tech-icons/express-original.svg" />
<img style="margin-right: 10px" alt="javascript" width="50" src="./attachments/tech-icons/javascript-original.svg" />
<img style="margin-right: 10px" alt="neo4j" width="50" src="./attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="socket.io" width="50" src="./attachments/tech-icons/socketio-original.svg" />
<img style="margin-right: 10px" alt="apachekafka" width="50" src="./attachments/tech-icons/apachekafka-original.svg" />
<img style="margin-right: 10px" alt="javascript" width="50" src="./attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="javascript" width="50" src="./attachments/tech-icons/express-validator.svg" />
<img style="margin-right: 10px" alt="googlecloud" width="50" src="./attachments/tech-icons/googlecloud-original.svg" />
<img style="margin-right: 10px; alt="postgresql" width="50" src="./attachments/tech-icons/postgresql-original.svg" /> ❌ (old)
</div>
<br>

### A note on the "❌ (old)" label on Postgres

Initially, I wrote the entirety i9lyfe's database in Postgres, a relational database.

Later on, when I was about implementing more advanced features, I found out that a Graph database better suits these features. This led me exploring deeper into the world of Graph databases, and that's when I realized that the whole database is even better written in a Graph database, and chose Neo4j as a replacement.

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation)
- [Feature Building Tutorials (Ref. Links)](#feature-building-tutorials-ref-links)
- [Get i9lyfe Up and Running (Local)](#get-i9lyfe-up-and-running-local)
- [ToDo List](#todo-list)

## Features

The following is a summary of the features supported by this API. Visit the API documentation to see the full features and their implementation details.

### Content Creation & Sharing

- **Create Post:** Create post of different types (inspired by Instagram) including *Photo*, *Video*, and *Reel*.
  - Mention users
  - Add hashtags

### Interactivity

- **Comment on Posts or Comments:** Write comments on posts and replies to comments.

- **React to Posts or Comments:** Reactions are represented as non-surrogate pair emojis.  
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

Get posts feed based on your interests and your follow network.

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

For all **HTTP request/response communication**: [Here](./apidoc/openapi.json)'s a well-written OpenAPI JSON document. Drop or Import it into a [Swagger Editor](https://editor.swagger.io/?_gl=1*1numedn*_gcl_au*MTUxNDUxNjEuMTc0MjY1MTg5Nw..) to access it.

For all **WebSocket real-time communication**: [Here](./apidoc/websocketsapi.md)'s a written markdown document.

## Feature Building Tutorials (Ref. Links)

The following are links to blog posts discussing and walking you through the build process of major features of the application.

Upcoming...

## Get "i9lyfe" Up and Running (Local)

### Install & Setup Node

### Install & Setup Neo4j

### Install & Setup Kafka

### Clone the Repository

### Environment Variables

### Install Project Dependencies

### Start Neo4j Service

### Start Kafka

### Start i9lyfe Server

---

# ToDo List

- Integrate an AI that observes the contents of a post and determines the "Interest" it best aligns with among the many interests in the social world.
  - This is a better way to implement the feature in which users follow interests, and posts are recommended to them based on the interests they follow.
  - For now, i9lyfe actually uses the "Hashtags" on posts to imply "Interests".
