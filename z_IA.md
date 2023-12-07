# Description
Social media with lotta cool features

# Affordances
**What can users do?** | **User actions?**
- Create an Account
- Login/Logout
- Change Password
- Email verification
---
- Create post or comment
- React (emojis) on post or comment
- Comment on post
- Reply on comment
- Bookmark post
- Follow/Unfollow users
---
- Edit profile
- Delete its post
---
- Chat: Start a conversation with users
  - Send text messages optionally with media attachments
- Leave conversation
- React on messages
- Reply messages

**What app can do?**
- Post feed: Fetch posts according to interests and follows, *algorithmically {AI}*
- Gather analytics on user posts and profiles
- Display analytics: Allow users to view their analytics both on their posts and profile
---
- Explore and Search: *Algorithmically {AI}* fetch all kinds of content from the app, and provide a sophisticated search system
---
- Notifications & Alert: Gets you notifications of activities pertaining to your account.
---
- User's active status in coversation
- When last a user was active

# Design Patterns
- Singleton
- Factory method
- Observer
- Strategy
- Command

# Architectural patterns
- REST
- Client-Server
- Model-{API Reponse}-Controller
- Event-driven
- Microservices | SOA ***

# Services and Integrations
- CDN
- Storage
- Message brokers for microservices ***

# App Integrations
- API Gateway (Proxy request router) for microservices *** (A unified entry point for frontend)
- Database
- Validator and Sanitizers
- AI Services
- WebSockets
- Analytics

# App features based on Learnt concepts
- Authentication - MFA
  - Password-based
  - Passwordless
- Rate limiting
- Security
  - CORS