# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

A Social Media API Server

## Intro

i9lyfe-server is an API server for Social Media Application, built with Node.js and Neo4j. It supports major social media application features that can be used to implement a mordern social media application frontend.

### Target Audience

#### Frontend Developers

If you're a frontend developer looking to build a Social Media Application, not just to have it static UI, but to also make it function.

The API documentation provides a detailed usage guide, following the OpenAPI specification.

#### HRMs, Startup Founders, Project Teams, Hiring Managers etc

If you're in need of a passionate, highly-skilled, expert-level backend engineer/developer. The codebase is easily accessible, as it follows a `(Routes)-->(Controllers)-->(Services)-->((Services) | (Model))` pattern.

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

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](./API%20doc.md)
- [Features Showcase (UI*)](#features-showcase-ui)
- [Building & Running the Application (Locally)](#building--running-the-application-locally)
- [Deploying the Application](#deploying-the-application)

## Features

The following is a summary of the features supported by this API. Visit the API documentation to see the full features and their implementation details.

### Content Creation & Sharing

- **Create Post:** Create post types inspired by Instagram including *Photo*, *Video*, *Story*, and *Reel*.

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

### Stories

Get updates of moments shared by those you follow in stories.

### Explore

Explore top (trending) content of different post types

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

## Features Showcase (UI*)

## Building & Running the Application (Locally)

## Deploying the Application
