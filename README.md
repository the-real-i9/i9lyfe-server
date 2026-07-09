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
- [API Tests](#api-tests-)

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
API (C4) Component Level Diagram: [Here](./diagrams/arch.pu). (Open in [PlantUML Editor](editor.platuml.com))

<details>
  <summary>Show Diagram</summary>

  ![i9lyfe_api_arch](https://www.plantuml.com/plantuml/png/nLdVR-Is4NxtNp7rGxC6j3RG106A5aKkwsjwpHdmUNrrtWuKDTOofaY5fDRU_puOaf98hNRjvhReaoroE9mVtpSSSd-mZXdNDHB4N-Nk4R-mLdnXXc_E_YGKbrs5yFVLTrUhxxwiTDDgXShzu-pC2ISHXX6u5gmsh857RQ8UU7Vx0TPejcZEphw1bLQE2OLcmLfHmFcUq7a1vp_fyukpi_NEEco-l7zn2soOaE6xWFFtz0NXCuZhMYAiR-vlpo_YiWMNkglMJXjMux1oHOCGH_SrPDDamc3jTRcec6CbirZ0-NNybvRUQkjgW-jF7yylClcdhNNQaFZFk5l-XyaElMg4TFaTLiAcApzrsE42jwZSsXbatc7wMePFk45hMIrKdTZljK0GcD7MTgNjofircTkLL-kkFA8bIWJtOrxpQABP2exxuGpZFfwhDVoaWKj1e92YDCpiQ6AbbHNMMNZi57T2A-kXGl9bnv_GGSCKgx51vGhWMctHEB9IbxyYT-3j0CKQjE0qLFfPISqg-7ptaVo-GpblwYmQFbkq4pSPttZ7wEyGaW9glJG-KVCkk0th3GU22GW5VMAzIK7jY8WxP9MzUZ32JIsmo3iZt2vQRWiuNyUXFi0s89QiaapLor-VNG4bDxlMPVRrgYQn_lV3FLXHAw7ggoqQyIWueza2QbHec0_tzyjArCA1qq-ehhYk49w5sm2-jC9uSK37BptKUjCzHmTit7YE9f3m1fFCvVOl1NXpzPMJUKpPOAotrQAgo1pFVqfuG5IsRKALSNVFc4Y8VIj2Cc7noxkgkZW2oz-8FNv_KcT1AATXoZLASEPwNgmYo4XgpjASw9wkusuvD0mowKIJyEykZamzj3l5BszOc_fwsvLIs4q1wwwqt8WIBmfunj9gejsH87XSKn0OTsABq5aq45RP0cpKHW8XEPwORLTUsQu4lc5AeGoQUWUWQolc2DjjC8eMDxuW1Lhtgue10DQV3I4oewWKJmWAdvTymrpX_ujpuQFLDWuhxGQWU-HcOABax4KtAPeh9kKOmginWdc2A39KUOsM4eqDoJUG356O_obADB3zjeWr66yPVyhYR_2t3gs33LCLhVB1ys5lZNuK4hCY-QYbrCy-GQg2JetVjCGMy8YER_flu0QPmmeeDhQ0D32se1JnPBP16X-_WsuoYVOQTh8P8McR4QC-11a-ZnNHrkYjgD27G1Yi8jfYuQEa1IVIA9CfsctoOGXDYyUEg5NN5lHNg20-875S0OiaocXHSVGuh7sMS-zkT8Mozyr_a4_hJnyBQEkNjnqNg3XhRIU927xRfK7fCzCt6l1mz-7-FKXT2nwjTfrHOD1sqbTfVzOGa9cKGuBxO5t2eQtBJiWgt_Fp1-2zYnOUZMxWxlsxQr9gKEAMAGV-58E5htyNWCePCPIsFUZd8F04IaI7cJIQ1uKED1yJyQWcQTCcf-n4S3ph3soxNsyFYiwUGRrqUhfxAEOrOkt7JbGbt0RXswLrs4Om1EMjqHojXON2Ps2SwqwvNa4yi23l1Jo_XM8ow35O2-l2oN_X9wDcYzPIJHqhKiQM83N693hxn-j1MWgY2Gdh_nQinHTXNMGmyYSgP4nuXgHEWczhl2_eF9_ymDZY55CdFT_wBedcMsRjipOLiPaQ6Z_eA-n0nnZj0bZdz9B8hC94kUlRC62-AVEzLaJN352ceIHuPT8PgUP1En6SQJQFIEVztG5q2xMFRg6-6jtZ4HWssGEPK3d8X3e2igJBp516x9aYZO16us6J16m7iJQ0ec7nuFKpnPZaFl59d2TncxeqJKioTppRYU2HR2V1xAs_Guikp-LjxvYV-nzLh25wncyoLAx9-AHuSQqUXMdsYjUmAXxbVcgYwfKIbWLa3FGa6AUfQ1FBiieM79t0RQniQGGYojBI5aRmHRHxjIkFmjjgd6yra_8s3NH4u64m7OXg2jzvaSDmdvT_fSRDBtgrr54_lD8DWMN24MqT_1dYt8aumgJq7RWFEu-gmp53oH3AeM5ypCmX98z69pZE3JuDZYTA4npl6THtVtx9ANffKdg2NheDBlJHVcjNdUt-DLrryYWLJ1ZQv6ZHAM8LLV__mZKru6ll8OTuEGlzh0xC2AJmV4rscgJNn5-zqiojEdhDw1UzwHwK2CxTWt9lZglAPQVQKXnlsAHOhWW8_tM9_DVz_UrG9mcS3rYkTNWy3alsgoVjxMKhOUWTkHcUltKRJxe9wJOiHGejFqrfzAeE5KJQTO0G2Pw4v1js2eRx4BwPr-cUNqlgFU1CafC2Dvt9edb8wFWjFnDzbNxvxgzIFB5lbeJZz646phmFZw5AduZ7qOCiAqOcaSmxph6bsvAYk-i6bQEBCrRntVGRI-Bn9PWGhJ2s8FvN3kix-_WBq3Zh6TbgQqKfCJxhBfrUHi7yWVSIlBS3sOU7Nu-XMVq5L4ZNUBha3M_1rEIX1Vzx1ZXcQamQBBhKJJW9siGNHI_LFq27zf78ykp33mPRkUilY7kcN4vUbjEypDJ29FDpAmxtlNFIXr6T0APa_AaskbFLWpPFmOmVXp68WzEVJPAw0rAhUcawvLyn9eM0QsMx1i667s66t2qgdFok4cvb-8AySmYZLU2OVKfVSeabGjMZUv6J7PD3wQzGYYsQNVekZ2yjAYksT1sWmanqMZPCYPPA9A4oEmL-51JMcUZaFzPS1p_wdsi8ap4ZCqid30u4FYm-Lmzcn8wa-GTKLTV8tm40 "i9lyfe_api_arch")

</details>

### Entity-Relationship Diagram

DBDiagram.io version: [Here](./diagrams/ER.dbml). View in [DBDiagram Editor](dbdiagram.io).

Mermaid version: [Here](./diagrams/ER.mermaid).

### Sequence Diagrams

API sequence diagrams: [Here](./diagrams/sequence-diagrams.md)

## Noteworthy Design Decisions

### Realtime message delivery -- (notifications, chat messages, etc.)

If a client isn't connected to receive realtime delivery of messages, particularly those that are "collection items", those messages are lost. i.e there's no dedicated queue for undelivered messages.

**Why?** Imagine a user who goes offline for a very long time, but we keep queueing messages for this user. This will accumulate load on our infrastructure, taking space that's needed by other users. This isn't an efficient user of server resources.
- Another is that, suppose we queue undelivered messages, by the time the user comes back online, we'll have to spend a lot of time dequeueing messages to the client. This will stress our WebSocket stream.

**So how do we ensure users who come online are kept up to date?** When a client is reconnected, it first requests (HTTP) for new items that might have been added to the collection since it went offline using the last item as a cursor. After this, he then connects to the realtime stream for realtime delivery of messages again.
- Even if there's an enormous amount of data to be fetched, we can fetch it in chunks (pagination) as the user browses towards the latest item in the collection.

**Tradeoff?** No messages are returned on this HTTP request, which may seem like a wasted effort. But, queueing undelivered messages plus long realtime streaming of those messages is very costly in the worst case scenario.

### Feed posts delivery

When a post is needed to be delivered to target users, we add the `postId` to each user's feed cache, then we signal each client of new feed posts. At its convenient time (based on UX decisions), the client requests for available new posts (with the cursor value of the latest post on their end as a starting point for the server).

Apart from this, the user can also forcibly "pull refresh" for available new posts.

### Aggregration problem: Post and Comment metrics, User follows count

### Chat History

**What's in a chat history?** To the normal viewer's eyes, a chat history only contains a ordered sequence of messages. But that's not true.

A chat history actually contains everything that happens in a chat including messages, message reactions, chat activities (for group chats), etc.