package pkg

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
)

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

// parse loads, parses and processes restaurants data about their daily menus
func (rc *RestaurantConfig) parse(resultChan chan<- *ScrapedRestaurant) {
	var lunchData = &ScrapedRestaurant{}
	lunchData.RestaurantName = rc.Name

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

	lunchData.process()

	resultChan <- lunchData
}

// findText finds TextNodes inside selected html snippet
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

// process processes scraped text and fills it to the ScrapedRestaurant struct
func (l *ScrapedRestaurant) process() {
	for _, dayRawData := range l.Raw {
		menu := &Menu{}

		rawData := menu.parseDate(dayRawData)
		menu.processLines(rawData)

		l.DailyMenus = append(l.DailyMenus, menu)
	}
}

// parseDate parses date and day from the data
func (m *Menu) parseDate(rawData []string) []string {
	dateRe := regexp.MustCompile(dateRegex)

	// day and date are in the same string
	if dateRe.MatchString(rawData[0]) {
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

// processLines processes lines to better format so each meal is for 1 line
func (m *Menu) processLines(data []string) {
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

		if containsPrice(line) {
			return false
		}

		if len(line) > bigLineLimit {
			return true
		}
	}
	return false
}

// containsPrice checks the line if it contains a czech currency kc
func containsPrice(line string) bool {
	return strings.Contains(strings.ToLower(line), "kč")
}
