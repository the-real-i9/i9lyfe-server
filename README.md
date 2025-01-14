# i9lyfe (API Server)

Build your Social Network Application

## Intro

i9lyfe-server is an API server for a Social Network Application, built with Node.js. It supports major social network application features that can be used to implement a mordern social network application.

### Who is this project for?

If you're a frontend developer looking to build a Social Network Application, not just to have it static, but to also make it function.

The goal of this API server is to support as many Social Network features as possible.

The API documentation provides a detailed usage guide. It doesn't follow the Open API specification, rather it follows Google's API documentation style sturcured in a simple markdown table, which I consider easier to work with.

### Open to suggestions

If your needs require more features than this API server currently supports, feel free to suggest them, and it will be added as soon as possible.

## Technologies

<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/nodejs-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/express-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/javascript-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/postgresql-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/socketio-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/apachekafka-original.svg" />
<img style="margin-right: 5px" alt="nodejs" width="50" src="./attachments/tech-icons/googlecloud-original.svg" />

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

- **Activity Updates:** Receive notifications about likes, comments, shares, and mentions.

## API Documentation

## Notable Features and their Algorithms

## Building & Running the Application (Locally)

## Deploying the Application