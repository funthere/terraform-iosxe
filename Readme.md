# Building a Terraform Provider

Consists of several components

*  A main.go which serves as the entry point to the provider
*  A provider package which implments the provider and is consumed by main.go
*  An api package which contains of a main.go which is the entry point to the server. This would not usually live within the same repository as the provider code, it's just here so that all the code for this example lives with in a single repository
    *  The api consists of two packages:
        *  server, which is the implementation of the webserver
        *  client, which is a client that can be used to programatically interact with the server.

## Requirements

* go => 1.18

### Routes

All Items are stored in memeory in a `map[string]Item`, where the key is the name of the Item.

The server has five routes:

*  POST /item  - Create an item
*  GET /item - Retrive all of the items
*  GET /item/{name} - Retrieve a single item by name
*  PUT /item/{name} - Update a single item by name
*  DELETE /item/{name} - Delete a single item by name

### Starting the Server

You can start the server by running `go run api/main.go` or `make startapi` from the root of the repository. This will start the server on `localhost:3001`

### Authentication

An non-empty `Authorization` header must be provided with all requests. The server will reject any requests without this.

## Client

The client can be used to programatically interact with the Server and is what the provider will use.

There is a `NewClient` function that will return a `*Client`. The function takes a hostname, port and token (The token can be anything that is not an empty string).

There are then 5 methods, GetAll, GetItem, NewItem, UpdateItem and DeleteItem, which map to the api endpoints of the server.