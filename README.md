# i9lyfe (API server)

## Intro

i9lyfe-server is an API server for a social media application modelled after Instagram

## Features

- Create post: Post types include *Photo*, *Video*, *Story*, and *Reel*; all accompanied by an optional description. Post may include @mentions and #hashtags.
- Comment on Post/Comment: Comments may contain text, media, or both. Its text content may include @mentions and #hashtags.
  > In this API, I've represented the notion of a *reply* as a *comment-on-comment*, as the former $-$ in my experience $-$ complicates API and database design.
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

## Processes/Activities and Technologies used

Generally, the API uses a RESTful architecture and is built using the NodeJS's Express.js framework.

### Data modelling

[ER diagram](./i9lyfe_ERD.png) - PNG file

[ER diagram](./i9lyfe_ERD.pgerd) - pgAdmin ERD file. Open using pgAdmin.

#### Tools

- **pgAdmin ERD Tool:** Its canvas and UI objects were used to create the database's entity relationship diagram.
  - Its card UI objects were used to represent tables, whose properties iconically identify the table schema, table name, attributes/columns along with their data types, unique key attributes, primary key attributes, and foreign key attributes.
  - Its relationship line UI objects were used to establish relationships between entities (tables); linking the primary key column of one table to the foreign key column of another or the same (circular).
  - I was able to work with its canvas comfortably with the aid of my drawing tablet.

### Authentication

#### Approach

#### Concepts

- OTP Auth:

- JWT Auth:

- Session (Cookie) Auth:

#### Technologies

- express-jwt

- express-session

### Database & Management

#### Technologies

- PostgreSQL
  - Tables
  - Views
  - Types
  - Functions

- node-postgres (pg):

#### Tools

- psql CLI

- pgAdmin

### Request Validation

- express-validator library:

### Realtime communication

- socket-io

### Handling user-generated content

- Google Cloud Storage:

  - Service Account Credential:

  - API Token Credential:

- @google-cloud/storage

### Security knots

#### Rate limiting

### Testing

- Jest/Supertest

- Postman

### Deployment

- GCP Gemini AI Support

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

### API documentation

- Swagger Open API

- API Blueprint

### Version control

- Git & GitHub

### Development tools

- Bash

- Linux

- ChatGPT AI

## Process story

### Data Modelling

### Deployment
