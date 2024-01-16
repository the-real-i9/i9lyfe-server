## Pending Implementations
- Chatting & Messaging | Conversations & Messages | Real-time communication
- Explore & Search
- Uploading User generated content to CDN
  - Profile pictures, Post images & videos
- Real time post data counts

### Chatting and messaging
- On pages that include entry points for chatting, return `conversation_id` as part of the data that renders that page, the value for `conversation_id` is either `null` or a valid `conversation_id` value.
- On the chat page where you can `findUsersToChat` from the search, return users with `conversation_id` as part of the data also, with either `null` or a valid `conversation_id` value.
- Now, to create a new or get an existing conversation, if the `conversation_id` is `null` for the returned user, then create a new conversation. Else, get the existing conversation with `conversation_id`
---
- On the conversations page, get all existing conversations involving the client_user in which one or more messages exist.
- Accessing a conversation, gets the following:
  - Participants involved
  - All Messages belonging to the conversation
---
- When you create a conversation, return the conversation data to the user, so the user doesn't need another request to get it.
- The UX is to chat with a user, whether a conversation with it exists or not.


## Chats & Messaging Entities
### Conversation
Contains groups and direct conversations

### User conversations
All the conversations involved in by a user, be it group or direct

When you want to list the conversatons for a particular user

### User memberships
Members of a groups and the group converation they belong to with their roles

### Message

### Message reaction

### Reports/Flags

### Blocked users

### User connection status

### Message deletion logs

### Notification Settings