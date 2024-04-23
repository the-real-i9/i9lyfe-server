### Session User JWT
`eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MywiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzEzOTAyMTA2fQ.myi_rjVxUwzNf-bkMSQI8fXaroP4MRufIo5fzs7F-to`

## Pending Implementations
- Uploading User generated content to Cloud Storage and getting back URL to store in DB - 27th
  - Profile pictures, Cover images, Post and Message binary datas
- Design document | API Blueprint
- Architectural diagramming
- Write all tests
- ER diagram
- DB Normalization
- Implementing OWASP Security measures

### Realtime Chat Events & Listeners
<table>
  <tr>
    <th>Event</th>
    <th>Listeners</th>
  </tr>
  <tr>
    <td>New Group Conversation</td>
    <td>
      <ul style=" padding-left: 20px;">
        <li><u>Conversations list page:</u> To display a new Group conversation.</li>
      </ul>
    </td>
  </tr>
  <tr>
    <td>New Message</td>
    <td>
      <ul style=" padding-left: 20px;">
        <li><u>Client App's Push-NS</u> To notify the user of a new message</li>
        <li><u>Conversations list page:</u> To display a new DM conversation <em>provided this message is a first.</em></li>
        <li><u>Conversation snippet:</u> To increment the unread messages count <code>AND</code> change the last message <code>AND</code> set and <code>POST</code> message's delivery status to delivered.</li>
        <li><u>Conversation history page:</u> To add new message to the chat history <code>AND</code> set message's delivery status to read.</li>
        <li><u>Message snippet:</u> Set and <code>POST</code> message's delivery status to read.</li>
      </ul>
    </td>
  </tr>
  <tr>
    <td>New Group Activity</td>
    <td>
      <ul style=" padding-left: 20px;">
      <li><u>Conversation snippet:</u> To change the last message</li>
      <li><u>Group conversation history page:</u> To add the new activity to the group chat history.</li>
      </ul>
    </td>
  </tr>
  <tr>
    <td>Message (Reaction | Deletion)</td>
    <td>
      <ul style=" padding-left: 20px;">
        <li><u>Client App's Push-NS:</u> To notify the user of a reation on the message. <em>(Reactions only).</em></li>
        <li><u>Message Snippet:</u> in Chat history page</li>
        <li><u>Conversation Snippet:</u> <em>provided this message is the last.</em></li>
      </ul>
    </td>
  </tr>
  <tr>
    <td>Message delivery status change</td>
    <td>
      <ul style=" padding-left: 20px;">
        <li><u>Message snippet:</u> Set message's delivery status accordingly.</li>
      </ul>
    </td>
  </tr>
</table>

## Brainstorming
### Trigger function: `message_delivery_acknowledgement()`
When a user inserts a new message -- `BEFORE INSERT`
- Include its `NEW`'s `delivered_to[]` and `read_by[]` the `NEW.sender_user_id`
- Set the user's `last_read_message_id` to `message_id`

When a user acknowledges `delivered` -- `BEFORE UPDATE`
- Subtract the old acknowledgement from the new to get the newly acknowledge user
- Set this user's `unread_messages_count + 1`
- if the number of `delivered_to` is equal to the number of users in that conversation, set `delivery_status` to 'delivered'.

When a user acknowledges `read` -- `BEFORE UPDATE`
- Subtract the old acknowledgement from the new to get the newly acknowledge user
- Set this user's `unread_messages_count - 1`
- Set this user's `last_read_message` to `message_id`
- if the number of `delivered_to` is equal to the number of users in that conversation, set `delivery_status` to 'delivered'.

### Home feed realtime new posts. How?
- When client connects, get the `user_id`s of all they follow. With this ID create a room through which clients subscribe to post updates from this user.
  ```js
  const followees_new_post_rooms = followees_user_ids.map((user_id) => `user_{user_id}_new_post_room`)
  socket.join(followees_new_post_rooms)
  ```
- When a user creates a new post, send the `newPostData` to its room to all its members
  ```js
  io.to(`user_{user_id}_new_post_room`).emit("new post", newPostData)
  ```

### Realtime updates of comments list. How?
- The implementation is similar to the subscribe-unsubscribe pattern discussed below. But the difference here is that it's going to be `subscribe to post new comments`.
- You should get the idea by now.

### Posts & Comments realtime data counts. How?
- When client connects, attach an event listener on their socket for the event `subscribe to post updates` while accepting the `post_id` of the post in subject. In the event handler, make socket join the room of subscribers to updates for that post
  ```js
  socket.on("subscribe to post data updates", (post_id) => {
    socket.join(`post_${post_id}_updates_subscribers`)
  })
  ```
  - Now when an update is received for this post
  ```js
  io.to(`post_{post_id}_updates_subscribers`).emit("post_data_update", updatedPostDataCounts)
  ```
  - All subscribing clients listening for `post_data_update` on this particular post will then receive updates accordingly.
- In addition, attach another event listener on their socket for the event `unsubscribe from post updates` while accepting the `post_id` of the post in subject.
  - In the event handler, make socket leave the room of subscribers to updates for that post.
  ```js
  socket.on("unsubscribe from post data updates", (post_id) => {
    socket.leave(`post_${post_id}_updates_subscribers`)
  })
  ```
  
