package main

import (
	"log"

	"github.com/gen1us2k/dbaas-proxy/api"
)

func main() {
	a, err := api.New()
	if err != nil {
		log.Fatal(err)
	}
	a.Run()
}
