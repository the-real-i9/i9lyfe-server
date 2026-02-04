# i9lyfe (API Server)

[![Test i9lyfe](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/the-real-i9/i9lyfe-server/actions/workflows/test.yml)

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
- **WebSocket** - Full-duplex, Bi-directional communications protocol
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
- [API Diagrams](#api-diagrams)
- [✨Technical Highlights✨](#technical-highlights)

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

### Following Interests

- Users will be able to follow interests (or topics). Content will also be recommended to them based on the interests they follow.

### Search: Full-text Search | Results Ranking | Fuzzy Matching

- Full-text search through content (photos, videos, and reels)
- Search user accounts
- Search hashtaged posts

### Content Recommendation System | User Feed

- A content recommendation system that pushes relevant content to the user's feed, based on:
  - User following network
  - User's interest followed
  - User engagement stats
  - and more recommendation parameters...

### User Follow Recommendation

- App will recommend users to follow, based on your follow network and content interaction stats.


## API Documentation

HTTP (REST) API: [Open Swagger JSON](./docs/swagger.json)

WebSockets API: [Open AsyncAPI JSON](./docs/asyncapi.json)


## API Diagrams

Architecture Diagrams: [See here](./diagrams/arch-diags.md)

ER diagram: [Open DBML source](./diagrams/ER.dbml). *(Open with the dbdiagram.io visualizer.)*

---
---

## ✨Technical Highlights✨

### Why I choose Redis over PostgreSQL to serve READ requests.

#### The cost of serving a UI component

The API implements READ requests that serve the whole data required to render a specific UI component. But, in reality, a UI component composes of multiple entities.

Hence, from a relational database, it is built from multiple related tables. For example, a post card UI component is built from: 
- the `users` table - for the post author info, 
- the `posts` table - for the post content,
- `post_reactions` - for the post count,
- `user_post_reactions` - for the end-user's reaction,
- `comments` table - for comments count,
- `reposts` table - for reposts count,
- `user_reposts` table - for the end-user's repost highlight,
- `post_saves` table - for post saves count, and, finally, 
-  the `user_post_saves` table for the end-user's post save check.

Each table's tree index will undergo scanning in logarithmic `O(log N)` time complexity, even though some of these tables contribute relatively low UI part compare to others. And, table JOINs at this level brings about low query performance.\
In addition to this is the cost of aggregation. Counting is a sequential operation. A COUNT aggregate data on a post card UI component, for example, can range from tens to hundreds of thousands.

This solution scales badly, and latency is high compared to a Redis solution.

The relatively dramatic reduction in cost that a key-value store like Redis provides comes from the fact that we can have hash entries that map directly to the data required to build the same UI component, allowing us to "GET" each of them in constant time `O(1)` complexity; no table scan involved, just one direct lookup to an already prepared data.

> Redis provides several data structures that makes it look almost like it's already predicting what kind of data we might need in that, we can store our data in one or more Redis data structures based on the kinds of accesses we want on that data, allowing us to choose the most efficient data structure for the kind of access we want.

Even for aggregate (count) data, a `ZCard` or `SCard` on a Redis Set data structure *(e.g. the set of users who reacted to or saved a post, the set of comments on a post, etc.)* executes in constant `O(1)` time complexity.

This solution scales well, with a relatively low latency. Overall performance is better for READ requests.

#### A highly simple, and more efficient pagination solution

Redis allows me to solve cursor-based pagination in a remarkably simpler way compared to how I would have done it with Postgres (by using timestamps for cursor, combined with a WHERE clause that does inequality check).

This solution originates from **event streaming** using Redis Streams.

After every user action/request like post creation, comment, reaction, save etc., an event for that action is added (queued in order of arrival) to the target event stream for that user action.

The goldmine here is that, Redis provides us with a **stream message ID** that is guaranteed to be different for each stream message (one added event) even to the microsecond.\
This stream message ID `string` is converted to `float64` and used as the `score` value in a Redis sorted set (ZSET) data structure for collection data.

> What if there's a collision in stream message ID?\
Well, that's a question I'd like to address in a job interview.

Now, when a client sends a READ request for a collection data, we access the target Redis sorted set (ZSET) data structure using the `ZRANGE` class of commands, to which we can specify a `limit` and a `score`. On each item returned in the collection, we attach a `cursor` key that holds the `score` value, so that the next collection of N items can be fetched with the cursor (`score`) of the last item in previous collection.

<!-- - I store JWT and session data in encrypted cookie for authentication and stateless session management, ensuring security and scalability.

- I employ the event-sourcing pattern; client request handlers queue events (e.g. users’ reaction to post) into Redis streams, from which dedicated background workers dequeue and execute background tasks (e.g. incrementing reactions count on post in Redis cache, notifying post owners, performing expensive operations in the primary database). This allows client requests to spend the smallest, inevitable processing time, delivering fast user experience.

- I use Redis's sorted set data structure to serve cursor-based, paginated results (e.g. post comments, user chats, chat messages, notifications etc.) to the client. Each result item includes a cursor data that can be supplied on the next request for a new chunk of N items.

- Client requests for aggregate data (e.g. reactions count on post) are computed in constant time from Redis's set data structure using ZCard (sorted) or SCard (unsorted). No linear-time aggregate function is executed, reducing latency and enhancing scalability.

- I serve all READ requests from the cache; practically, whole data is built from parts in relevant cache entries. This offers fast user experience. Meanwhile, cached data are dynamically made fresh by relevant WRITE requests, zeroing the chances of having a stale cache at anytime. However, for certain data results, “eventual consistency” reasonably holds. -->

