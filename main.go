package main

import (
	. "github.com/jsmadis/menu-scraper/pkg"
	"log"
)

const(
	DefaultRestaurantsPath = "config/restaurants.yml"
)

func main() {
	var restaurants Restaurants

	err := restaurants.Load(DefaultRestaurantsPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := restaurants.Scrape()

	err = result.Print()

	if err != nil {
		log.Fatal(err)
		return
	}

}
