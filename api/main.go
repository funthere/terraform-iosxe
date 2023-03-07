package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/meirizal/terraform-experiment/api/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	seed := flag.String("seed", "", "a file location with some data in JSON form to seed the server content")
	flag.Parse()

	items := map[string]server.Item{}

	if *seed != "" {
		seedData, err := ioutil.ReadFile(*seed)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(seedData, &items)
		if err != nil {
			log.Fatal(err)
		}
	}

	// --------- db no-sql ---------
	mongoCtx := context.TODO()
	opts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(mongoCtx, opts)
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(mongoCtx)

	if err = client.Ping(mongoCtx, readpref.Primary()); err != nil { // ping connection
		panic(err)
	}
	db := client.Database("dd_ios") // connect to db name

	itemService := server.NewService("localhost:3001", items, db)

	err = itemService.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
