package pkg

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
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
	Restaurants []ScrapedRestaurant
}

type ScrapedRestaurant struct {
	RestaurantName string
	Raw            [][]string
	Err            error
	DailyMenus     []*Menu
}

type Menu struct {
	Day   string
	Date  string
	Lines []string
}

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

func (r *Restaurants) Scrape() ScrapedRestaurants {
	var resultsChan = make(chan ScrapedRestaurant, len(r.Restaurants))
	var resultArray = make([]ScrapedRestaurant, len(r.Restaurants))

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

func (rc *RestaurantConfig) parse(resultChan chan<- ScrapedRestaurant) {
	var lunchData ScrapedRestaurant

	response, err := http.Get(rc.Url)
	if err != nil {
		lunchData.Err = err
		resultChan <- lunchData
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		lunchData.Err = fmt.Errorf("status code error: %d %s", response.StatusCode, response.Status)
		resultChan <- lunchData
		return
	}

	body, err := DecodeHTMLBody(response.Body)

	if err != nil {
		lunchData.Err = err
		resultChan <- lunchData
		return
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		lunchData.Err = err
		resultChan <- lunchData
		return
	}

	selection := doc.Find(rc.Selector)

	for _, node := range selection.Nodes {
		lunchData.Raw = append(lunchData.Raw, findText(node.FirstChild, nil))
	}

	lunchData.RestaurantName = rc.Name
	lunchData.process()

	resultChan <- lunchData
}

func findText(node *html.Node, out []string) []string {
	if node == nil {
		return out
	}

	if node.Type == html.TextNode {
		if data := strings.TrimSpace(node.Data); data != "" {
			out = append(out, data)
		}
	}
	out = findText(node.FirstChild, out)
	return findText(node.NextSibling, out)
}

func (l *ScrapedRestaurant) process() {
	for _, dayRawData := range l.Raw {
		menu := &Menu{}

		rawData := menu.parseDate(dayRawData)
		menu.parseLines(rawData)

		l.DailyMenus = append(l.DailyMenus, menu)
	}
}
func (m *Menu) parseDate(rawData []string) []string {
	dateRe := regexp.MustCompile(dateRegex)

	// day and date are in the same string
	if dateRe.MatchString(rawData[0]){
		m.Date = dateRe.FindString(rawData[0])
		m.Day = strings.Split(rawData[0], " ")[0]
		return rawData[1:]
	} else {
		// date is the second string
		m.Day = rawData[0]
		// pivnice u capa has unnecessary whitespaces
		m.Date = strings.ReplaceAll(rawData[1], " ", "")
		return rawData[2:]
	}
}

func (m *Menu) parseLines(data []string)  {
	var line string

	// Menu is missing
	if len(data) < 5 {
		m.Lines = data[:]
		return
	}

	for i, rawLine := range data {
		line = line + " " + rawLine

		if len(rawLine) > bigLineLimit && mustSplit(i+1, data) || containsPrice(rawLine) {
			m.Lines = append(m.Lines, line[:])
			line = ""
		}
	}
}

// mustSplit Split based on length of line or position of line with price (kc)
func mustSplit(start int, data []string) bool {
	for i := start; i < len(data); i++ {
		line := data[i]

		if containsPrice(line){
			return false
		}

		if len(line) > bigLineLimit {
			return true
		}
	}
	return false
}

func containsPrice(line string) bool {
	return strings.Contains(strings.ToLower(line), "kč")
}

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