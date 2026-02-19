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
- **Fiber v3** - HTTP (REST) API Framework
- **PostgreSQL** - Relational DBMS
- **SQL** - Structured Query Language for Relational Databases
- **PL/pgSQL** - Procedural Language for PostgreSQL. Stored functions.
- **Neo4j** - Graph DBMS
- **CypherQL** - Query Language for Neo4j
- **WebSocket** - Full-duplex, Bi-directional communications protocol. Realtime communication.
- **Redis Key/Value Store** - Cache. Fast data structures. Pagination. Aggregation.
- **Redis Streams** - Event-based messaging system. Background tasks queue.
- **Redis Pub/Sub** - PubSub pattern messaging system
- **Google Cloud Storage** - Cloud object storage
---
- **JWT** - User authentication. Token signing and verification.
- **MessagePack** - Object serializer and deserializer (major use)
- **JSON** - Object serializer and deserializer (minor use)

### Tools
- **Swagger** - HTTP API Documentation
- **AsyncAPI** - Websockets API Documention
- **Docker** - Container running Postgres and Redis instances
- **Git & GitHub** - Repository & Version Control
- **GitHub Actions** - Continuous Integration. Unit & Integration Testing
- VSCode
- Ubuntu Linux

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Upcoming features](#upcoming-features)
- [API Documentation](#api-documentation-)
- [API Diagrams](#api-diagrams-)
  - [Architecture Diagram](#architecture-diagram)
  - [ER Diagram](#entity-relationship-diagram)
  - [Sequence Diagrams](#sequence-diagrams)
- [Articles](#articles-)

<!-- - [Technical Highlights](#technical-notes)
  - [Why I serve all READ requests from Redis (cache).](#why-i-serve-all-read-requests-from-redis-cache)
  - [Why I offload content media processing to client-side.](#why-i-offload-content-media-processing-to-client-side)
  - [Why event streaming and background workers?](#why-event-streaming-and-background-workers) -->

## Features

The following are the features supported by this API. *Visit the API documentation for implementation guides.*

### Content Creation & Sharing

**Create posts** of various types (inspired by Instagram) including *Photo*, *Video*, and *Reel*.
  - Mention users
  - Include hashtags

### Feed

- Browse posts from people you follow
- Receive new posts in real-time.

### Content Engagement and Interaction

- **React** to Posts and Comments
- **Comment** on Posts and Comments (replies).
- **Repost:** Re-share posts on your feed.
- **Access Interactions:** Access comments on posts and replies to comments, and access the list of users who have reacted to a post or comment.
- **Save** posts for later

### User Profile Management

- View your profile
- Access to other user profiles.
- Edit your profile.
- Manage your posts.
- Access to saved posts.
- Access to posts you've engaged with through likes and comments.
- Access to posts you we're mentioned in.

### Networking

- Follow or unfollow users.

### Chatting and Direct Messaging

- Realtime chat with users of the application.
- Supports various message types including:
  - Text and voice message
  - Images and videos with caption
  - Audio
  - File attachments (Documents)
- Realtime user presence status and last seen.


### Realtime Notification and Message Delivery

- Users receive likes, comments, reposts, and mentions notifications involving them in realtime.
- Chat messages are delivered to target users in realtime.
- New posts are delivered to user's feed in realtime.

### Realtime Updates

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

### Discover: Content Recommendation System

- A content recommendation system that pushes relevant content to the user's feed, based on:
  - User following network
  - User's interest followed
  - User engagement stats
  - and more recommendation parameters...

### User Follow Recommendation

- App will recommend users to follow, based on your follow network and content interaction stats.


## API Documentation &#x1f4d6;

HTTP API (REST): [Here](./docs/swagger.json). Open in [Swagger Editor](editor.swagger.io)

WebSockets API: [Here](./docs/asyncapi.json) Open in [AsyncAPI Editor](studio.asyncapi.com).


## API Diagrams &#x1f3a8;

### Architecture Diagram
API (C4) Component Level Diagram: [Here](./arch.pu). (Open in [PlantUML Editor](editor.platuml.com))

<details>
  <summary>Show Diagram</summary>

  ![i9lyfe_api_arch](https://www.plantuml.com/plantuml/png/nLdVR-Is4NxtNp7rGxC6j3RG106A5aKkwsjwpHdmUNrrtWuKDTOofaY5fDRU_puOaf98hNRjvhReaoroE9mVtpSSSd-mZXdNDHB4N-Nk4R-mLdnXXc_E_YGKbrs5yFVLTrUhxxwiTDDgXShzu-pC2ISHXX6u5gmsh857RQ8UU7Vx0TPejcZEphw1bLQE2OLcmLfHmFcUq7a1vp_fyukpi_NEEco-l7zn2soOaE6xWFFtz0NXCuZhMYAiR-vlpo_YiWMNkglMJXjMux1oHOCGH_SrPDDamc3jTRcec6CbirZ0-NNybvRUQkjgW-jF7yylClcdhNNQaFZFk5l-XyaElMg4TFaTLiAcApzrsE42jwZSsXbatc7wMePFk45hMIrKdTZljK0GcD7MTgNjofircTkLL-kkFA8bIWJtOrxpQABP2exxuGpZFfwhDVoaWKj1e92YDCpiQ6AbbHNMMNZi57T2A-kXGl9bnv_GGSCKgx51vGhWMctHEB9IbxyYT-3j0CKQjE0qLFfPISqg-7ptaVo-GpblwYmQFbkq4pSPttZ7wEyGaW9glJG-KVCkk0th3GU22GW5VMAzIK7jY8WxP9MzUZ32JIsmo3iZt2vQRWiuNyUXFi0s89QiaapLor-VNG4bDxlMPVRrgYQn_lV3FLXHAw7ggoqQyIWueza2QbHec0_tzyjArCA1qq-ehhYk49w5sm2-jC9uSK37BptKUjCzHmTit7YE9f3m1fFCvVOl1NXpzPMJUKpPOAotrQAgo1pFVqfuG5IsRKALSNVFc4Y8VIj2Cc7noxkgkZW2oz-8FNv_KcT1AATXoZLASEPwNgmYo4XgpjASw9wkusuvD0mowKIJyEykZamzj3l5BszOc_fwsvLIs4q1wwwqt8WIBmfunj9gejsH87XSKn0OTsABq5aq45RP0cpKHW8XEPwORLTUsQu4lc5AeGoQUWUWQolc2DjjC8eMDxuW1Lhtgue10DQV3I4oewWKJmWAdvTymrpX_ujpuQFLDWuhxGQWU-HcOABax4KtAPeh9kKOmginWdc2A39KUOsM4eqDoJUG356O_obADB3zjeWr66yPVyhYR_2t3gs33LCLhVB1ys5lZNuK4hCY-QYbrCy-GQg2JetVjCGMy8YER_flu0QPmmeeDhQ0D32se1JnPBP16X-_WsuoYVOQTh8P8McR4QC-11a-ZnNHrkYjgD27G1Yi8jfYuQEa1IVIA9CfsctoOGXDYyUEg5NN5lHNg20-875S0OiaocXHSVGuh7sMS-zkT8Mozyr_a4_hJnyBQEkNjnqNg3XhRIU927xRfK7fCzCt6l1mz-7-FKXT2nwjTfrHOD1sqbTfVzOGa9cKGuBxO5t2eQtBJiWgt_Fp1-2zYnOUZMxWxlsxQr9gKEAMAGV-58E5htyNWCePCPIsFUZd8F04IaI7cJIQ1uKED1yJyQWcQTCcf-n4S3ph3soxNsyFYiwUGRrqUhfxAEOrOkt7JbGbt0RXswLrs4Om1EMjqHojXON2Ps2SwqwvNa4yi23l1Jo_XM8ow35O2-l2oN_X9wDcYzPIJHqhKiQM83N693hxn-j1MWgY2Gdh_nQinHTXNMGmyYSgP4nuXgHEWczhl2_eF9_ymDZY55CdFT_wBedcMsRjipOLiPaQ6Z_eA-n0nnZj0bZdz9B8hC94kUlRC62-AVEzLaJN352ceIHuPT8PgUP1En6SQJQFIEVztG5q2xMFRg6-6jtZ4HWssGEPK3d8X3e2igJBp516x9aYZO16us6J16m7iJQ0ec7nuFKpnPZaFl59d2TncxeqJKioTppRYU2HR2V1xAs_Guikp-LjxvYV-nzLh25wncyoLAx9-AHuSQqUXMdsYjUmAXxbVcgYwfKIbWLa3FGa6AUfQ1FBiieM79t0RQniQGGYojBI5aRmHRHxjIkFmjjgd6yra_8s3NH4u64m7OXg2jzvaSDmdvT_fSRDBtgrr54_lD8DWMN24MqT_1dYt8aumgJq7RWFEu-gmp53oH3AeM5ypCmX98z69pZE3JuDZYTA4npl6THtVtx9ANffKdg2NheDBlJHVcjNdUt-DLrryYWLJ1ZQv6ZHAM8LLV__mZKru6ll8OTuEGlzh0xC2AJmV4rscgJNn5-zqiojEdhDw1UzwHwK2CxTWt9lZglAPQVQKXnlsAHOhWW8_tM9_DVz_UrG9mcS3rYkTNWy3alsgoVjxMKhOUWTkHcUltKRJxe9wJOiHGejFqrfzAeE5KJQTO0G2Pw4v1js2eRx4BwPr-cUNqlgFU1CafC2Dvt9edb8wFWjFnDzbNxvxgzIFB5lbeJZz646phmFZw5AduZ7qOCiAqOcaSmxph6bsvAYk-i6bQEBCrRntVGRI-Bn9PWGhJ2s8FvN3kix-_WBq3Zh6TbgQqKfCJxhBfrUHi7yWVSIlBS3sOU7Nu-XMVq5L4ZNUBha3M_1rEIX1Vzx1ZXcQamQBBhKJJW9siGNHI_LFq27zf78ykp33mPRkUilY7kcN4vUbjEypDJ29FDpAmxtlNFIXr6T0APa_AaskbFLWpPFmOmVXp68WzEVJPAw0rAhUcawvLyn9eM0QsMx1i667s66t2qgdFok4cvb-8AySmYZLU2OVKfVSeabGjMZUv6J7PD3wQzGYYsQNVekZ2yjAYksT1sWmanqMZPCYPPA9A4oEmL-51JMcUZaFzPS1p_wdsi8ap4ZCqid30u4FYm-Lmzcn8wa-GTKLTV8tm40 "i9lyfe_api_arch")

</details>

### Entity-Relationship Diagram

DBDiagram.io version: [Here](./diagrams/ER.dbml). View in [DBDiagram Editor](dbdiagram.io).

Mermaid version: [Here](./diagrams/ER.mermaid).

### Sequence Diagrams

API sequence diagrams: [Here](./diagrams/sequence-diagrams.md)

## Articles &#x1f4f0;
*Coming Soon...*

- [Design Decisions behind a High-performance API]()

<!-- 

---
---

## Technical Notes

### Why I serve all READ requests from Redis (cache).

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

Each table's tree index will undergo scanning in logarithmic `O(log n)` time complexity, even though some of these tables contribute relatively low UI part compare to others. And, table JOINs at this level brings about low query performance.\
In addition to this is the cost of aggregation. Counting is a sequential operation. A COUNT aggregate data on a post card UI component, for example, can range from tens to hundreds of thousands.

This solution scales badly, and latency is high compared to a Redis solution.

The relatively dramatic reduction in cost that a key-value store like Redis provides comes from the fact that we can have hash entries that map directly to the data required to build the same UI component, allowing us to "GET" each of them in constant time `O(1)` complexity; no table scan involved, just one direct lookup to an already prepared data.

> Redis provides several data structures that makes it look almost like it's already predicting what kind of data we might need, in that, we can store our data in one or more Redis data structures based on the kinds of accesses we want on that data, allowing us to choose the most efficient data structure for the kind of access we want.

Even for aggregate (count) data, a `ZCard` or `SCard` on a Redis Set data structure *(e.g. the set of users who reacted to or saved a post, the set of comments on a post, etc.)* executes in constant `O(1)` time complexity.

This solution scales well, with a relatively low latency. Overall performance is better for READ requests.

Another benefit derived from Redis is that I can use the same hash entry with a single data structure to serve multiple requests. `ZRANGE` on `ZSET` `(user:X:followers)` allows me to get a collection of followers `O(n)`, while `ZCARD` on the same entry gives me the total number of followers `O(1)`

#### A simpler, and more efficient pagination

Redis allows me to solve cursor-based pagination in a remarkably simpler way compared to how I would have done it with Postgres (by using timestamps for cursor, combined with a WHERE clause that does inequality check).

This solution originates from **event streaming** using Redis Streams.

After every user action/request like post creation, comment, reaction, save etc., an event for that action is added (queued in order of arrival) to the target event stream for that user action.

The goldmine here is that, Redis provides us with a **stream message ID** that is guaranteed to be different for each stream message (one added event) even to the microsecond.\
This stream message ID `string` is converted to `float64` and used as the `score` value in a Redis sorted set (ZSET) data structure for collection data.

> What if there's a collision in stream message ID?\
Well, that's a question I'd like to address in a job interview.

Now, when a client sends a READ request for a collection data, we access the target Redis sorted set (ZSET) data structure using the `ZRANGE` class of commands, to which we can specify a `limit` and a `score`. On each item returned in the collection, we attach a `cursor` key that holds the `score` value, so that the next collection of N items can be fetched with the cursor (`score`) of the last item in previous collection.

### Why I offload content media processing to client-side.

Creating a post, adding comment a with attachment, sending a photo, video, audio, file message, and changing profile picture, all involve binary multimedia processing.

These WRITE requests are hot, frequently made to the API server, and as such should be as cheap as possible.

Media processing is a heavy work requiring hardware resource time and quality. If we do media processing on the server-side, hundreds to thousands of requests will compete for server resources, slowing down response time even API requests that don't involve media processing. This becomes a source of performance bottleneck for the API server. This is at the far end of "cheap".

Even if we choose to offload media processing to a dedicated server, that's additional infrastructure which comes at a high price, a price that'll be unnecessary when today's client's hardware resources are built for high performance media processing. And client requests sent to this dedicated server will still compete for hardware resources&#x2014;still, the requests aren't cheap.

Client hardware resources today are running tools like Figma, Adobe graphics processing products, and, some, AI models. I don't think it's expensive if we also allow our client-side to take advantage of the client's hardware capabilities.

The level of media processing our social media system requires isn't something even many old hardwares with just a CPU and a graphics card can't handle. Cropping, blurring, compressing, audio enhancement, etc., all these are media processing operations that WhatsApp executes fast on a mobile device without an internet connection.

So, as media processing is offloaded to client-side, a client only experiences the slight delay caused by their own hardware capabilities, rather than the heavy delay caused by the competition between many client WRITE requests involving media processing.

### Why event streaming and background workers?

In the client-side's reality, the task that the client asks the server to perform by sending a WRITE request is usually small and cheap. But, in the server-side reality, no task is actually cheap, even if it appears small to the client. This is because, on the server-side, there are two sides to the task, one is the main, cheap task that the client is asking the server to perform, and the other is the side effect of performing that main task, which mosttimes can be expensive.

A high performance, scalable system separates the main task (client task) of a WRITE request from its side effects (server task). A WRITE request that the client expects to perform task X shouldn't perform task X, Y, and Z. A client request that says "create post", isn't saying "create post", then "notify the users I mentioned", then "update cache A, B, C", then "show my post to users D, E, F".

On a WRITE request path we should only perform the client task and resulting side effects (server task) should be performed off this path, this way we ensure the client's request is cheap. How we achieve this is through the **event-sourcing pattern**, using event streaming with background workers.

For a WRITE request, when client task is done, an event that communicates "X has been done" is added to a target event stream for that particular task along with necessary parameters, and we immediately return to the client.\
Then, a dedicated background worker watching and consuming this stream, performs the necessary side effects (server task) in the background, for each added event.

> Event stream goes by several names including event stream, event queue, or background task queue. All these are the same thing.

In this API, background tasks (side effects) include cache management, message delivery, expensive database operations, and more. -->

<!-- - I store JWT and session data in encrypted cookie for authentication and stateless session management, ensuring security and scalability.

- I make use of the event-sourcing pattern; client request handlers queue events (e.g. user reacted to post) into Redis streams, from which dedicated background workers dequeue and execute necessary background tasks like caching for future READs, notifying post owners, performing expensive database operations etc. This allows client requests to be as cheap as possible, thereby delivering fast user experience.

- I use Redis's sorted set data structure to serve cursor-based, paginated results (e.g. post comments, user chats, chat messages, notifications etc.) to the client. Each result item includes a cursor data that can be supplied on the next request for a new chunk of N items.

- Client requests for aggregate data (e.g. reactions count on post) are computed in constant time from Redis's set data structure using ZCard (sorted) or SCard (unsorted). No linear-time aggregate function is executed, reducing latency and enhancing scalability.

- I serve all READ requests from the cache—practically, whole data is built from parts in relevant cache entries. This offers fast user experience. Meanwhile, the cache is dynamically made fresh by relevant WRITE requests. This is backed up by the database for possible cache misses, however, for collection or aggregate data results, “eventual consistency ” reasonably holds. -->

