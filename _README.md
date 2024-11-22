# i9lyfe-server

i9lyfe-server is an API server for a social media application modelled after Instagram

## Features

### Content Creation & Sharing

- **Create post:** Post types include *Photo*, *Video*, *Story*, and *Reel*; all may include a description, which may include @mentions and #hashtags.

### Interactivity

- **Comment on Post/Comment:** Comments may contain text, media, or both. Its text content may include @mentions and #hashtags.
  > In this API, I've represented the notion of a *reply* as a *comment-on-comment*, as the former — in my experience — complicates API and database design.
- **React to Post/Comment:** Reactions are basically, non-surrogate pair, emojis.
- Repost posts.
- View comments on posts and comments on comments. View users who reacted to posts.

### User profile

- Edit your profile
- Delete your contents
- View contents you saved
- View contents you reacted to or commented on

### Networking

- Follow/Unfollow users
- Content Suggestions/Recommendations
- View user profiles and their contents

### Home feed & Explore section

- Pagination (Infinite-scrolling)
- New posts, relevant to the user, are delivered in real-time.
- Post interactions including reactions count, comments count, reposts count, and saves count, are updated in realtime

### Messaging

- Initiate DM chat with user
- Messages may come in various types including text, voice note, image, video, audio, and file attatchment. image, video, and audio may optionally include a description
- React to messages and view reactions
- Realtime chat experience

### Discovery and Search

- **Search Bar:** Helps users find accounts, hashtags, content, or topics of interest.
- **Hashtags:** Organizes posts under common themes for easy discovery.

### Content Curation

- Save specific posts
- Retrieve saved posts

### Notifications

- **Activity Updates:** Alerts about likes, comments, shares, or mentions.

## Doings

The following sections comprise of notable processes and activities in the API and its development. Each process discusses important aspects of the process, some of which are: *Approach*, *Concepts* applied, *Technologies* used, File *Attachments*, and *Tools* used.

Generally, the API uses a RESTful architecture and is built using the NodeJS's Express.js framework.

## Data modelling

### Tools

- **pgAdmin ERD Tool:** Its canvas and UI objects were used to create the database's entity relationship diagram.
  - Its card UI objects were used to represent tables, whose properties iconically identify the table schema, table name, attributes/columns along with their data types, unique key attributes, primary key attributes, and foreign key attributes.
  - Its relationship line UI objects were used to establish relationships between entities (tables); linking the primary key column of one table to the foreign key column of another or the same (circular).
  - I was able to work with its canvas comfortably with the aid of my drawing tablet.
  
### Attachments

[ER diagram](./i9lyfe_ERD.png) - PNG file

[ER diagram](./i9lyfe_ERD.pgerd) - pgAdmin ERD file. Open with pgAdmin.

## Authentication

### Approach

The Signup process involved three steps:

- In the first step, the user submits a new email (which isn't already registered with the API) for verification, after which the API sends a 6-digit verification code to the email
- In the second step, the user submits the 6-digit code they received via email for validation
- In the third and final step, the user provides their registration information

Each next step is dependent on the success of the previous.

### Concepts

- **OTP Auth:** Used in cases where email verification is required before allowing the user to perform a transaction. Use cases in the API include, the signup transaction, and the forgot password transaction.

- **JWT Auth:** We issue a JWT when user signup is successful or when credentials are valid at user signin. Subsequent requests to protected endpoints attach the JWT in the Authorization header for authetication.
  > *Concern 1:* What if a client performs multiple login
  requests to the API and the API issues a JWT for each, shouldn't we invalidate the JWTs previously issued??
  >
  > *Concern 2:* A client logs out the user only by deleting the JWT issued to it. Is that really the best way? Shouldn't we also invalidate the JWT on the backend?
  >
  > *I plan to settle these concerns and update the API accordinly.*

- **Session Management:** Transactions involving a number of steps or lined-up requests, — like "Signup" and "Password Reset" — need to maintain a session between these requests. I used HTTP cookie with the express session middleware to achieve this.

### Technologies

- **jsonwebtoken:** Used for JWT signing

- **express-jwt:** The express jwt middleware that handles JWT authentication for subsequent requests, performing automatic JWT verification and populating `req.auth` with user data.

  Setting the `credentialsRequired` parameter to `false` for public routes allows us to tailor our response data differently for the case where we have an authenticated user and for the case where we do not. This, for example, allows a client to see a user's profile and followers even when they're not logged in. And if they're logged in, they'll be able to see if they're following a user (*the unfollow button will be displayed instead of the follow button*). If they're not logged in, pressing the follow button will redirect them to the login page, as a JWT is not provided in the Authorization header of the POST request.

  ```js
    export const getFollowers = async (req, res) => {
    try {
      const { username } = req.params

      const { limit = 50, offset = 0 } = req.query

      const userFollowers = await UserService.getFollowers({
        username,
        limit,
        offset,
        client_user_id: req.auth?.client_user_id, // client is optional, data will be tailored accordingly
      })

      res.status(200).send(userFollowers)
    } catch (error) {
      console.error(error)
      res.sendStatus(500)
    }
  }
  ```

- **express-session:** The express session middleware that handles session (cookie) management for "Signup" and "Password Reset". The session store **connect-pg** integrates with it to keep session data.

## Database & DB Management

The API uses **PostgreSQL** as its RDBMS.

### Notable PostgreSQL Objects

#### Tables

Of course, I'll use tables. But the ones that vett my proficiency with the DBMS follow.

#### Views

I used Views to represent specific UI components, for exmaple, the post card component. The View attributes consists of the properties of the UI component it represents. The "PostView", for example, includes the `reactions_count`, `comments_count`, `reposts_count`, and `saves_count` attributes, among others (these attributes are not calculated with every SELECT query, optimizing SELECT's performance).

#### Stored Functions

> I get that the info here is somewhat lengthy, but its worth the read — you'll enjoy it.

One thing you'll observe as you inspect the `*.model.js` files is that, all database queries are small and 99% include function calls. Well, that's because I made a heavy use of stored functions to handle business logic. It completely takes over the big deal of handling database transactions, and it saves the API server multiple round-trips to the database server required to complete a single business task.

I was getting overwhemled by the complexity and weight of database code. Large-size queries were written directly into the `*.model.js` files, buisiness tasks require complex database transaction constructs; it resulted into a javascript code that's difficult to navigate and understand.

I tried using `WITH` queries coupled with "SQL Triggers" with the aim of simplifying things, but even with these I also got an incomprehensible database code. I became stressed and burned-out. You'll be confused if you decided to go through the whole code.

It wasn't until when I was working on i9chat — a chat application modelled after WhatsApp and Messenger, and got to a point where I needed a programmatic solution to achieve a certain behaviour, and found "stored functions and procedures" to be the solution, that I realized that "stored functions" are the solution to all my database manipulation problems.

I had heard about "stored functions and procedures", but I postponed it for later learning seeing it claims to be an advance feature of PostgreSQL. Little did I know that the project I decided to take on as my first is pretty much advance itself.

As I continued to learn and integrate "Stored functions" into aspects of my database and business logic code, I began to realize that PLpgSQL, which is the language you use to write the functions, is no different from an actual programming language, seeing it comprises of all the constructs I already understood in programming languages. On realizing this, my head tingled *"Wooow, database programming is a thing??? Think of all the possibilities!"* I immediately got the idea that I could rewrite all my messy database and business logic code using PLpgSQL, statement-after-statement, in a stored function, passing the necessary input parameters and returning as much result as I need.

It wasn't long after this realization that I suspended "i9chat-server", revisited this application (i9lyfe-server),  and reworked its business logic and database code into "stored functions". One of the reworkings that I love is this portion that implements the business task of "creating a post".

Post model code:

```js
export class Post {
  /**
   * @param {object} post
   * @param {number} post.client_user_id
   * @param {string[]} post.media_urls
   * @param {string[]} post.mentions
   * @param {string[]} post.hashtags
   * @param {"photo" | "video" | "reel" | "story"} post.type
   * @param {string} post.description
   */
  static async create({ client_user_id, media_urls, type, description, mentions, hashtags }) {
    const query = {
      text: "SELECT new_post_id, mention_notifs FROM create_post($1, $2, $3, $4, $5, $6)",
      values: [
        client_user_id,
        [...media_urls],
        type,
        description,
        [...mentions],
        [...hashtags],
      ],
    }

    return (await dbQuery(query)).rows[0]
  }
}
```

`create_post` function definition:

```sql
CREATE FUNCTION public.create_post(OUT new_post_id integer, OUT mention_notifs json[], client_user_id integer, in_media_urls text[], in_type text, in_description text, mentions character varying[], hashtags character varying[]) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  ret_post_id int;
  
  mention_username varchar;
  ment_user_id int;
  
  client_data json;
  
  mention_notifs_acc json[] := ARRAY[]::json[];
  
  hashtag_n varchar;
BEGIN
  INSERT INTO post (user_id, type, media_urls, description)
  VALUES (client_user_id, in_type, in_media_urls, in_description)
  RETURNING id INTO ret_post_id;
  
  -- populate client data
  SELECT json_build_object(
   'id', id,
   'username', username,
   'profile_pic_url', profile_pic_url
  ) INTO client_data FROM i9l_user WHERE id = client_user_id;
  
  
  FOREACH mention_username IN ARRAY mentions
  LOOP
    SELECT id INTO ment_user_id FROM i9l_user WHERE username = mention_username;

        -- skip if mentioned user is not found
    CONTINUE WHEN ment_user_id IS NULL;

    -- create mentions
    INSERT INTO pc_mention (post_id, user_id) 
    VALUES (ret_post_id, ment_user_id);
    
    -- skip mention notification for client user
    CONTINUE WHEN ment_user_id = client_user_id;
    
    -- create mention notifications
    INSERT INTO notification (type, sender_user_id, receiver_user_id, via_post_id)
    VALUES ('mention_in_post', client_user_id, ment_user_id, ret_post_id);
    
    -- accumulate mention notifications
    mention_notifs_acc := array_append(mention_notifs_acc, json_build_object(
      'receiver_user_id', ment_user_id,
      'sender', client_data,
      'type', 'mention_in_post',
      'post_id', ret_post_id
    ));
  END LOOP;
  
  
  -- create hashtags
  FOREACH hashtag_n IN ARRAY hashtags
  LOOP
    INSERT INTO pc_hashtag (post_id, hashtag_name)
    VALUES (ret_post_id, hashtag_n);
  END LOOP;
  
  new_post_id := ret_post_id;
  mention_notifs := mention_notifs_acc;
  
  RETURN;
END;
$$;
```

A lot of PLpgSQL constructs including variable declarations, conditional statements, loops, function input parameters, function return types, function output parameters, and error handling were useful accross stored functions.

#### Types

I used types particulaly as return types of stored functions in order to simplify complex return values — when things got serious, and to represent the data object to be returned to the client from the application server.

Although, for some types such as `ui_post_struct` and `ui_comment_struct`, our Views already contain the properties (attributes) we need. Our Views, however, do not consist results narrowed to a single client user (i.e. the API request user for which we're executing the function), rather, by default, they hold results for all users in our database. Returning types from our stored functions allows us to have results narrowed to a single client user.

### Notable PostgreSQL Features

#### Full-text Search

The API supports its search & filter feature with PostgreSQL's `to_tsquery()` and `to_tsvector()` functions searching through all text-based data (usernames, post descriptions, hashtags etc.) for the query text, and, of course, there's the option to restrict your search to a set of content types.

#### Notable *DML clauses* used with `SELECT`

`GROUP`, `UNION`, `INNER JOIN`, `LEFT JOIN`, `ORDER BY`, `DISTINCT`

#### Some *Aggregate functions* used

`COUNT`, `array_agg()`, and `json_agg()`

#### Some *JSON functions* used

`json_build_object()`

#### Some *Array Functions* used

`array_append()`

### Technologies

- **node-postgres (pg):** The Node.js database driver used for PostgreSQL. Personally, I prefer to write DDL and DML queries purely and I don't like to delegate the job to ORMs. I'm not here to give my own take on the use of ORMs, but I think I know enough database concepts, SQL, and PostgreSQL not to need an ORM. Besides, this project is pretty big and advanced for the use of ORMs. That said, I've had good experience with the Sequelize ORM in small projects.

### Attachments

[SQL Backup file](./i9lyfe_database_backup.sql) - Scan through the schema definitions and object definitions — **function definitions**, in particuler. It also includes sample data for testing.

### Tools

#### postgres client tools

- **psql:** This tool was a swiss army knife throughout the development process. I used it many times to inspect and make mimimal changes to the database. I also used it when setting up PostgreSQL; to CREATE a new USER X, and ALTER database OWNER TO X (after restoring from backup)
- **pg_dump:** Used when backing-up the database.

  I backup my database's state each time I make changes to any  of its definitions, so that when I accidentally damage my database — which happened💀, I can be able to restore it back to its last working state.
  
  The backup is also useful when you'll want to populate an empty database with an existing backup in a new environment. I remember doing this when deploying the database in my GCE VM instance, also when using sample data to run endpoint tests in GitHub Actions.
- **dropdb**, **createdb**, **pg_restore | psql**: Used when restoring the database. The process involves the following steps:
  - Use **dropdb** to drop the existing (faulty) version of the database, IF one EXISTS.
  - Use **createdb** to create a new database with the same name.
  - Use **psql** to ALTER database OWNER TO my_username.
  - Use **pg_restore** to restore from  a ".dump" backup or **psql** to restore from a ".sql" backup.

#### pgAdmin

Alongside psql is pgAdmin, which I use to handle complex, bulky task, particulary in situations where its editor interface makes the task easier and more convenient.

Definitions of objects such as tables, functions, and views involve lenghty lines of SQL code that require careful, algorithmic thinking and multiple changes throughout the definition process.

I also use it when constructing complex multi-table queries that involve multiple JOIN stacks; primarily for the purpose of inspecting how each JOIN layer affects the resulting data, thereby ensuring that the query works correctly.

## Request Validation

Request data such as POST/PUT request bodies, query parameters, and variable segements, need to be validated for correctness, according to the requirements of an API endpoint.

### Technologies

- **express-validator:** An express request validation middleware.

  I particularly love it for the way it helped define structural constraints on complex POST request bodies — specifically those of *Post* and *Message* that come in varying structures depending on post or message type.

## Realtime communication

The API uses WebSockets for realtime communication

### Use cases in the application

#### Home feed new post updates

First, all clients should listen for the event `new post` on the home feed. When a client comes online, he is added to the *"new post"* room of each of the users he follows, so as to receive their *"new post"* whenever it is broadcasted to their *"new post"* room. Now, whenever the client creates a new post, he broadcasts it to his *"new post"* room, and all users active in this room receives the *"new post"*.
> Note: By *"new post"*, we're actually referring to the post data.

Below is the portion of code that implements this functionality. See [postComment.realtime.service.js](./src/services/realtime/postComment.realtime.service.js)

```js
// When a client comes online i.e. when their socket connects
const followeesNewPostRooms = (
  await User.getFolloweesIds(client_user_id)
).map((followee_user_id) => `user_${followee_user_id}_new_post_room`)

// they subsribe to their followees new post rooms, listening for any new post they might publish
socket.join(followeesNewPostRooms)

// when a client creates a new post, they broadcast it to their new post room by invoking this function, so any client who has subscribed above receives it
static sendNewPost(user_id, newPostData) {
  PostCommentRealtimeService.io
    ?.to(`user_${user_id}_new_post_room`)
    .emit("new post", newPostData)
}
```

#### Post/Comment metrics updates

When a *"post/comment card/snippet component"* comes into view in the UI, a client subscribes to *"post/comment metric updates"* for that post/comment. This subscription adds them to the room of *"metric update subscribers"* for that post.

Now, whenever a client reacts to, comments on, or reposts this post/comment, this post/comment broadcasts the update to its room of *"metric update subscribers",* and all clients (client sockets) active in that room will receive the broadcast, updating their *"post/comment card/snippet UI component"* accordingly.

When the *"post/comment card/snippet component"* goes out of view in the UI, the client can then unsubscribe to stop receiving updates for it.

> When to subscribe or unsubscribe for updates is actually the businesses design decision. That, above, is just my suggestion.

Below is the portion of code that implements this functionality. See [postComment.realtime.service.js](./src/services/realtime/postComment.realtime.service.js)

```js
/* To start receiving metrics update for post when in view in UI */
socket.on("subscribe to post metrics update", (post_id) => {
  socket.join(`post_${post_id}_metrics_update_subscribers`)
})

/* To stop receiving metrics update for post when out of view in UI */
socket.on("unsubscribe from post metrics update", (post_id) => {
  socket.leave(`post_${post_id}_metrics_update_subscribers`)
})
```

#### Notification updates

When a client comes online (i.e. socket connection established), they immediately receive an update on their number of unread notifications.

When a client needs to be notified of an activity that involves them — they're followed or someone reacted to their post, a notification is streamed to them via a WebSocket connection.

See [notification.service.js](./src/services/notification.service.js)

#### DM Chat & Messaging

A client handles the data associated with each event accordingly.

```js
/**
 * @param {"new chat" | "new message" | "message delivered" | "message read" | "message reaction" | "message reaction removed" | "message deleted"} event
 * @param {number} partner_user_id
 * @param {object} data
 */
static send(event, partner_user_id, data) {
  ChatRealtimeService.sockClients.get(partner_user_id)?.emit(event, data)
}
```

See [chat.realtime.service.js](./src/services/realtime/chat.realtime.service.js)

### Technologies

- **socket-io:** The WebSocket library used for Realtime Communication.

## Handling user-generated content

User-generated media from user profiles, posts, and chat messages shouldn't be kept on the API server. Using a managed storage service is the best way to go.

Consider the case when a client initiates the changing of their profile picture in the API. The steps below, contained in a middleware dedicated to file upload, are executed.

- The API receives the request body which includes the picture in binary format — specifically an array of unsigned 8-bit integers.
- The API takes the property it expects to contain the picture binary data, uploads the binary data to the GCS bucket dedicated to the API while specifying a location path that ends in the file name.
- This location path combined with google cloud storage bucket's public domain, [https://storage.googleapis.com](https://storage.googleapis.com), to form a full, publicly accessible URL.
- The property holding the binary data is then deleted for a new property that holds the generated URL, in the request body.
- The modified request body is then passed from this middleware on to the request handler.

### Technology: Google Cloud Storage API

#### Bucket name

i9lyfe-bucket

#### Credentials

Credentials authenticate clients into the API, and the ones used are:

- **Service Account:** I consider this option a really sophisticated form of authorization.

  Service accounts are like passes that have a selected number of permissions attached to them. When a client uses a Service Account, they take on the roles specified in the SA, and use the scopes defined in it.

  Google APIs require roles to access them, and/or scopes to tell them what you want. A Google user himself has the highest form of access via their email, he has unrestricted access to any Google API (this isn't secure). SAs are like emails, but their access levels and restrictions can be controlled. You can have an SA that can access Cloud Storage, but not Cloud Run; and you can later modify it to access Cloud Run, but not Cloud Storage.

  One particular role required for GCS is the **"Storage Admin"** role.

- **API Key:** I find this to be the easiest form of credential to use in production.

  It also has its own access level and restriction control. You can restrict its access to one or more APIs, or even one or more domain.

> The fact that you can control their access levels and restrictions doesn't mean you should be lax about storing them securely.

- **ADC** (Application Default Credential): I set this up using the **gcloud CLI**. I think it's best used in a development environment, since you have to set it up on the client machine, which is your local machine or a remote VM instance.

  But if you like setting things up, you can go ahead to set it up in production in your VM instance. The best approach is to:
  - Create one Service Account Credential that has access to all the Google APIs used in the application server.
  - Use the **gcloud CLI** to perform an *application-default* login, as this Service Account.

    ```bash
    gcloud auth application-default activate-service-account
    ```

    Follow the next steps, and you have for yourself a default creadential in your application.

  The API Client Library automatically detects this credential without you having to explicity pass any. That is cool.
  
#### API Client Library

- **@google-cloud/storage:** The Node.js Google Cloud API Client library used to upload files to "i9lyfe-bucket".

## Security knots

### Rate limiting

## Testing

### Technologies

- **Jest/Supertest:** For endpoint testing. Check out the [test](./tests/) folder to see them.

- **Postman:** For functional testing of the application.

  I created a *collection* for this application which includes folders categorizing the endpoints contained in them.

  I've spent hours, days, weeks, and months working with this app for every API I'm working on.
  