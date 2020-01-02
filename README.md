<p align="center"><img src="https://github.com/hewiefreeman/GopherDB/raw/master/logo.png" width="25%" height="25%"></p>
<h1 align="center" color="#69a2ff">
  
<?xml version="1.0" encoding="utf-8"?><svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100" height="50"><text font-size="16" x="0" y="0"><tspan fill="#69a2ff">GopherDB</tspan></text></svg>

GopherDB aims at easing the creation, retrieval, and manipulation of data through a robust, yet simple query format. Gopher DB will feature many table types tailored for specific use-cases including secure user authentication, leaderboards for online games, keystores, ordered tables, and more to come. It also features a powerful schema validation mechanism to keep your data well organized, while being extremely flexible so you can fine-tune to your exact needs.

Much like MongoDB, GopherDB uses JSON as it's query/response language and means of storing data on the disk. Where GDB excels is the simplification of the query process, allowing you to target and manipulate any piece of data in an entry as a hierarchy of `Object`, `Array`, and `Map`. This not only makes building queries easier, but they're also more expandable and effective than any other JSON query format.
<br>
<br>
<br>
<p align="center">:construction: <b>PROJECT IN DEVELOPMENT</b> :construction:</p>
<br>
<hr>
<br>

## Main Features
  - In-depth schema validation
  - Standardized format across insert, update, and get queries
  - Many useful methods for arithmetic, comparisons, list append/prepend, etc.
  - Wide selection of data types and settings
  - User authentication tables (single select queries only)
  - Key-value tables (multi & single select queries)
  - Ordered list tables (multi & single select queries)
  - Leaderboards (multi & single select queries)
  
> **Recommendations**: All feature recommendations will be taken into consideration. This includes new security features, data types, methods, table types, etc. (*Security feature recommendations will be dealt with at the highest priority*)
  
### Data Types
When creating a table in GopherDB, you will need to make a schema that describes what types of data the database will store, and how. These are all the data types available in GDB, and one or more must be used when creating a database schema:

  - Boolean
  - Unsigned Integer (8, 16, 32, and 64 bit)
  - Integer (8, 16, 32, and 64 bit)
  - Float (32 & 64 bit)
  - String
  - Array
  - Map
  - Object (AKA Schema)
  - Time (AKA Date)
  
## Installing
Binaries will be created when project is considered stable. For now, you must download and use the Go source with:

  ```go get github.com/hewiefreeman/GopherDB```

And the dependencies:

 `go get github.com/json-iterator/go` ([JSON-iterator](https://github.com/json-iterator/go))

`keystore` is the only stable package as of right now. You can test all functionalities of the keystore package with this command from the `keystore` directory:

 ```go test -v keystore_test.go```

<hr>

<h6>GopherDB and all of it's contents Copyright 2020 Dominique Debergue
<h6>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at:

  `http://www.apache.org/licenses/LICENSE-2.0`

<h6>Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.</h6>
