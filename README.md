# i9lyfe (API Server)

Build your Social Network Application

## Intro

i9lyfe-server is an API server for a Social Network Application, built with Node.js. It supports major social network application features that can be used to implement a mordern social network application.

### Who is this project for?

If you're a frontend developer looking to build a Social Network Application, not just to have it static, but to also make it function.

The goal of this API server is to support as many Social Network features as possible.

The API documentation provides a detailed usage guide. It doesn't follow the Open API specification, rather it follows Google's API documentation style sturcured in a simple markdown table, which I consider easier to work with.

### Open to suggestions

If you need a feature this API server does not currently support, feel free to suggest them, and it will be added as soon as possible.

## Technologies

<div style="display: flex;">
<img style="margin-right: 10px" alt="nodejs" width="50" src="./attachments/tech-icons/nodejs-original.svg" />
<img style="margin-right: 10px" alt="express" width="50" src="./attachments/tech-icons/express-original.svg" />
<img style="margin-right: 10px" alt="javascript" width="50" src="./attachments/tech-icons/javascript-original.svg" />
<img style="margin-right: 10px" alt="neo4j" width="50" src="./attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="socket.io" width="50" src="./attachments/tech-icons/socketio-original.svg" />
<img style="margin-right: 10px" alt="apachekafka" width="50" src="./attachments/tech-icons/apachekafka-original.svg" />
<img style="margin-right: 10px" alt="googlecloud" width="50" src="./attachments/tech-icons/googlecloud-original.svg" />
<span style="margin-right: 10px">
<img alt="postgresql" width="50" src="./attachments/tech-icons/postgresql-original.svg" /> ❌ (old)
</span>
</div>

### More

- JWT
- express-validator

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation)
- [Notable Features and their Algorithms](#notable-features-and-their-algorithms)
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

## Notable Features and their Algorithms

## Building & Running the Application (Locally)

## Deploying the Application
