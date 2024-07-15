# Notes

## Production ToDos: Docker

- Set all secrets in Github action secrets
- Just before the step that starts the server, write the necessary secrets into a .env file

## Pending Implementations

- Uploading User generated content to Cloud Storage and getting back URL to store in DB
  - Profile pictures, Cover images, Post and Message binary datas
- Design document | API Blueprint
- Architectural diagramming
- Write all tests
- ER diagram
- DB Normalization
- Implementing OWASP Security measures

```js
// bulk
[
  ["kendrick", 10, 4],
  ["starlight", 10, 4],
  ["itz_butcher", 10, 4],

  ["itz_butcher", 11, 9],
  ["johnny", 11, 9],
  ["starlight", 11, 9],

  ["itz_butcher", 11, 8],
  ["johnny", 11, 8],
  ["starlight", 11, 8],

  ["kendrick", 12, 5],
  ["starlight", 12, 5],
  ["johnny", 12, 5],

  ["itz_butcher", 13, 10],
  ["kendrick", 13, 10],
  ["johnny", 13, 10],
]
```
