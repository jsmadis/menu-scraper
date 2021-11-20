package pkg

type PreFetchFilter struct {
	Tags Tags
}

type Tags []string

func (p *PreFetchFilter) Filter(restaurantConfigs []*RestaurantConfig) []*RestaurantConfig {
	var filteredRestaurants []*RestaurantConfig
	for _, restaurantConfig := range restaurantConfigs {
		if p.filterRestaurantConfig(restaurantConfig) {
			filteredRestaurants = append(filteredRestaurants, restaurantConfig)
		}
	}
	return filteredRestaurants
}

func (p * PreFetchFilter) filterRestaurantConfig(config *RestaurantConfig) bool {
	filterFunctions := []func(*RestaurantConfig) bool {
		p.Tags.contains,
	}

	filtered := true
	for _, filterFunc := range filterFunctions {
		filtered = filtered && filterFunc(config)
	}
	return filtered
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

func (t *Tags) contains(config *RestaurantConfig) bool {
	// don't filter out if filter is not used
	if len(*t) == 0 {
		return true
	}
	for _, tag := range *t {
		for _, rTag := range config.Tags {
			if tag == rTag {
				return true
			}
		}
	}
	return false
}
