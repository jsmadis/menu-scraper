package main

import (
	"flag"
	. "github.com/jsmadis/menu-scraper/pkg"
	"log"
	"sort"
)

const (
	DefaultRestaurantsPath = "config/restaurants.yml"
)

func main() {
	var restaurantPath string
	var today bool
	var preFetchFilter = &PreFetchFilter{}

	flag.BoolVar(&today, "today", false, "Prints only today's menus.")
	flag.StringVar(&restaurantPath, "restaurantPath", DefaultRestaurantsPath, "Path that contains a file with configuration of restaurants.")
	flag.Var(&preFetchFilter.Tags, "tag", "Tag to filter restaurants")
	flag.Var(&preFetchFilter.RestaurantName, "name", "Name of the restaurant you want to fetch.")
	flag.Parse()

	var restaurants Restaurants

	err := restaurants.Load(restaurantPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	restaurants.Filter(preFetchFilter)
	scrapedRestaurants := restaurants.Scrape()

	if today {
		scrapedRestaurants.FilterTodayMenus()
	}

	//sort to get always the same order of restaurants
	sort.Sort(scrapedRestaurants)

	err = scrapedRestaurants.Print()

	if err != nil {
		log.Fatal(err)
		return
	}

}
