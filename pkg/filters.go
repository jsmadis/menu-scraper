package pkg

type PreFetchFilter struct {
	Tags Tags
	RestaurantName RestaurantName
}

type RestaurantName []string
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
		p.RestaurantName.contains,
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

func (rn *RestaurantName) String() string {
	if *rn != nil {
		return rn.String()
	}
	return ""
}

func (rn *RestaurantName) Set(value string) error {
	*rn = append(*rn, value)
	return nil
}

func (rn *RestaurantName) contains(config *RestaurantConfig) bool {
	// don't filter out if filter is not used
	if len(*rn) == 0 {
		return true
	}

	for _, rName := range *rn {
		if config.Name == rName {
			return true
		}
	}
	return false

}
