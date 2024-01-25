## Pending Implementations
- Chat routes and controllers - 23rd
- Post and Comments realtime - 25th
  - New post, new comment
  - Post and Comment data counts
- Explore & Search - 26th
- Create Database Indexes
- Uploading User generated content to Cloud Storage and getting back URL to store in DB - 27th
  - Profile pictures, Post images & videos
- Frontend preparation - 28th



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

### Hints
- For every conversation (direct or group), create a room socket, and join particpants to it ($2$ or $N > 1$ respectively). Every event that happens in a conversation is triggered on the target room socket and sent to all its participants.


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