# GopherDB
GopherDB aims at easing the creation, retrieval, and manipulation of data for secure user authentication, leaderboards, and generic database storage. GDB uses a powerful schema validation mechanism to keep your data well organized, while being extremely flexible, so you can fine-tune to your exact needs.

Much like MongoDB, GopherDB uses JSON as it's query/response language and means of storing data on the disk. Where GDB excels is the simplification of the query process, allowing you to target and manipulate any piece of data in an entry as a hierarchy of `Object`, `Array`, and `Map`. This not only makes building queries and schemas easier, but they're also more readable than ever. On top of that, GDB has built-in number type arithmetic, and `String`, `Object`, `Array`, and `Map` manipulation methods (eg: append, prepend, delete, etc) using the same simple query format.

:warning: **PROJECT IN DEVELOPMENT** :warning:
<br>
<br>
## Main Features
  - In-depth schema validation
  - Easy, simple query format
  - Wide selection of data types and settings
  - User Authentication Tables (single select queries only)
  - Key-value Tables (multi & single select queries)
  - List Tables (multi & single select queries)
  - Leaderboards (multi & single select queries)
  
### Data Types
When creating a table in GopherDB, you will need to make a schema that describes what types of data the database will store, and how. These are all the data types available in GDB, and one or more must be used when creating a database schema:

  - Boolean
  - Unsigned Integer (8, 16, 32, and 64 bit)
  - Integer (8, 16, 32, and 64 bit)
  - Float (32 & 64 bit)
  - String
  - Array
  - Object
  - Map
  - Time/Date

