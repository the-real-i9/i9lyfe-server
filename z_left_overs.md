## Pending Implementations
- Chatting & Messaging | Conversations & Messages | Real-time communication
- Explore & Search
- Uploading User generated content to CDN
  - Profile pictures, Post images & videos
- Real time post data counts

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
        <li><u>Conversation snippet:</u> To increment the unread messages count <code>AND</code> change the last message</li>
        <li><u>Conversation history page:</u> To add new message to the chat history.</li>
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
</table>