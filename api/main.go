package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/meirizal/terraform-experiment/api/server"
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

	itemService := server.NewService("localhost:3001", items)
	err := itemService.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
