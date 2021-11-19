package pkg

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const (
	dateRegex    = `\d{2}\.\d{2}\.\d{4}`
	bigLineLimit = 25
	printingTemplate = `

{{ range .Restaurants }}
{{ .RestaurantName }}

{{ range .DailyMenus }}
	{{ .Day }} ({{ .Date }})
	{{ range .Lines }}
		{{ . }}
	{{ end }}
{{ end }}
	================================================================================================================
{{ end }}
`)

type Restaurants struct {
	Restaurants []*RestaurantConfig `yaml:"restaurants"`
}

type RestaurantConfig struct {
	Name     string `yaml:"name"`
	Url      string `yaml:"url"`
	Selector string `yaml:"selector"`
}

type ScrapedRestaurants struct {
	Restaurants []*ScrapedRestaurant
}

// Load loads restaurants configuration
func (r *Restaurants) Load(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fileData, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(fileData, r)
	if err != nil {
		return err
	}

	return nil
}

// Scrape scrapes menus and saves them to the ScrapeRestaurants struct
func (r *Restaurants) Scrape() ScrapedRestaurants {
	var resultsChan = make(chan *ScrapedRestaurant, len(r.Restaurants))
	var resultArray = make([]*ScrapedRestaurant, len(r.Restaurants))

	for _, restaurant := range r.Restaurants {
		go restaurant.parse(resultsChan)
	}

	for i, _ := range r.Restaurants {
		resultArray[i] = <-resultsChan

		if resultArray[i].Err != nil {
			log.Print(resultArray[i].Err)
		}
	}
	return ScrapedRestaurants{Restaurants: resultArray}
}

// FilterTodayMenus dummy way how to remove menus for other days
func (sr *ScrapedRestaurants) FilterTodayMenus() {
	dateToday := time.Now().Format("02.01.2006")

	for _, restaurant := range sr.Restaurants {
		var menuToday []*Menu
		for _, menu := range restaurant.DailyMenus {
			if menu.Date == dateToday {
				menuToday = append(menuToday, menu)
			}
		}
		restaurant.DailyMenus = menuToday
	}
}

// Print prints scrapedRestaurants in pretty format
func (sr *ScrapedRestaurants) Print() error {
	restaurantTemplate, err := template.New("ScrapedRestaurants").Parse(printingTemplate)

	if err != nil {
		return err
	}

	err = restaurantTemplate.Execute(os.Stdout, sr)

	if err != nil {
		return err
	}
	return nil
}