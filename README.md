# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe is an API server for Social Media Application, built with Node.js and Neo4j. It supports major social media application features that can be used to implement a mordern social media application frontend.

### Target Audience

#### Frontend Developers

If you're a frontend developer looking to build a Social Media Application, not just to have it static UI, but to also make it function.

The API documentation provides a detailed usage guide, following the OpenAPI specification.

#### HRMs, Startup Founders, Project Teams, Hiring Managers etc

If you're in need of a passionate, highly-skilled, expert-level backend engineer/developer, you've found the right one for the job.

The codebase is easily accessible, as it follows a (Routes)-->(Controllers)-->(Services)-->((Services) | (Model)) pattern. Nonetheless, I provide an easy-to-follow graphical map for navigating the codebase in [this section](#codebase-map).

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

Initially, I wrote i9lyfe's database entirely with Postgres, a relational database.

Later on, when I was about implementing more advanced features, I found out that a Graph database better suits these features. This led me exploring deeper into the world of Graph databases, and that's when I realized that the whole database is even better written in a Graph database, and chose Neo4j as a replacement.

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation)
- [Codebase Map](#codebase-map)
- [Feature Building Tutorials (Blog Posts)](#feature-building-tutorials-blog-posts)
- [Building & Running i9lyfe Locally](#building--running-i9lyfe-locally)
- [Deploying & Running i9lyfe Remotely](#deploying--running-i9lyfe-remotely)

## Features

The following is a summary of the features supported by this API. Visit the API documentation to see the full features and their implementation details.

### Content Creation & Sharing

- **Create Post:** Create post of different types (inspired by Instagram) including *Photo*, *Video*, and *Reel*.

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
- View saved posts

### Networking

- Follow or unfollow other users.  
- Explore content suggestions and recommendations tailored to your interests.  
- View user profiles and their content.

### Home Feed

Get posts feed based on your interests and your follow network.

### Explore

Explore different post types based on a content recommendation algorithm.

### Reels

Yes, just like you're thinking. Swipe, swipe and swipe up through an exhautsing list of reel videos.

### Chatting and Messaging

- Chat with users in the application
- Supported message types include text, voice notes, images, videos, audio files, and file attachments (with description).

### Search

- **Search and Filter:** Find users, hashtags, posts (photo, video, reel), or topics of interest.  
- **Hashtags:** View top posts with specific hashtags

### Notifications

- **Activity Updates:** Receive notifications about likes, comments, shares, and mentions.

### Real-Time Updates

- New posts arrive on top of your home feed in realtime.
- Individual posts receive real-time interaction updates.

## API Documentation

For all **HTTP request/response communication**: [Here](./apidoc/openapi.json)'s a well-written OpenAPI JSON document. Drop or Import it into a [Swagger Editor](https://editor.swagger.io/?_gl=1*1numedn*_gcl_au*MTUxNDUxNjEuMTc0MjY1MTg5Nw..) to access it.

For all **WebSocket real-time communication**: [Here](./apidoc/websocketsapi.md)'s a written markdown document.

## Codebase Map

Here's a graphical map of the code base for easy navigation.

## Feature Building Tutorials (Blog Posts)

The following are links to blog posts discussing and walking you through the build process of major features of the application.

## Building & Running i9lyfe Locally

## Deploying & Running i9lyfe Remotely
