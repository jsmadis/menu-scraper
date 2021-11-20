package pkg

type PreFetchFilter struct {
	Tags Tags
}

type Tags []string

func (p *PreFetchFilter) Filter(restaurants *Restaurants) {
	p.Tags.filter(restaurants)
}

func (t *Tags) String() string {
	if *t != nil {
		return t.String()
	}
	return ""
}

func (t *Tags) Set(value string) error {
	*t = append(*t, value)
	return nil
}

func (t *Tags) filter(restaurants *Restaurants) {
	if len(*t) == 0 {
		return
	}

	var filteredRestaurants []*RestaurantConfig
	for _, restaurant := range restaurants.Restaurants {
		if t.contains(restaurant.Tags) {
			filteredRestaurants = append(filteredRestaurants, restaurant)
		}
	}
	restaurants.Restaurants = filteredRestaurants
}

func (t *Tags) contains(restaurantTags []string) bool {
	for _, tag := range *t {
		for _, rTag := range restaurantTags {
			if tag == rTag {
				return true
			}
		}
	}
	return false
}
