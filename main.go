package main

import (
	"flag"
	. "github.com/jsmadis/menu-scraper/pkg"
	"log"
)

const(
	DefaultRestaurantsPath = "config/restaurants.yml"
)

func main() {
	var restaurantPath string
	var today bool

	flag.BoolVar(&today,"today", false,  "Prints only today's menus.")
	flag.StringVar(&restaurantPath, "restaurantPath", DefaultRestaurantsPath, "Path that contains a file with configuration of restaurants.")
	flag.Parse()

	var restaurants Restaurants

	err := restaurants.Load(restaurantPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := restaurants.Scrape()

	if today {
		result.FilterTodayMenus()
	}

	err = result.Print()

	if err != nil {
		log.Fatal(err)
		return
	}

}
