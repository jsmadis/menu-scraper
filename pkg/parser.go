package pkg

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	dateRegex = `\d{2}\.\d{2}\.\d{4}`
	dayRegex  = `^(pondělí|úterý|středa|čtvrtek|pátek|sobota|neděle)$`
)

type Line string

type ScrapedRestaurant struct {
	RestaurantName string
	Raw            [][]Line
	Err            error
	DailyMenus     []*Menu
}

type Date string

type Menu struct {
	Day   string
	Date  Date
	Lines []Line
}

// we need to follow time.WeekDay format
var czechDayNames = []string{
	"pondělí",
	"úterý",
	"středa",
	"čtvrtek",
	"pátek",
	"sobota",
	"neděle",
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
func findText(node *html.Node, out []Line) []Line {
	if node == nil {
		return out
	}

	if node.Type == html.TextNode {
		if data := strings.TrimSpace(node.Data); data != "" {
			out = append(out, Line(data))
		}
	}
	out = findText(node.FirstChild, out)
	return findText(node.NextSibling, out)
}

// process processes scraped text and fills it to the ScrapedRestaurant struct
func (l *ScrapedRestaurant) process() {

	// Bauman has the menu in one dimension
	if len(l.Raw) > 7 {
		l.splitRawData()
	}

	for _, dayRawData := range l.Raw {
		menu := &Menu{}

		rawData := menu.parseDate(dayRawData)
		menu.processLines(rawData)

		l.DailyMenus = append(l.DailyMenus, menu)
	}
}

// splitRawData splits the raw data in format where each array of string represents menu for one day
func (l *ScrapedRestaurant) splitRawData() {
	dayRe := regexp.MustCompile(dayRegex)
	var splitRawData [][]Line
	var lines []Line
	for _, o := range l.Raw {

		for _, line := range o {
			if line == "" {
				continue
			}
			if dayRe.MatchString(line.ReplaceAllToLower(" ", "")) && len(lines) != 0 {
				splitRawData = append(splitRawData, lines)
				lines = make([]Line, 0)
			}
			lines = append(lines, line)
		}
	}
	l.Raw = append(splitRawData, lines)
}

// parseDate parses date and day from the data
func (m *Menu) parseDate(rawData []Line) []Line {
	dateRe := regexp.MustCompile(dateRegex)
	dayRe := regexp.MustCompile(dayRegex)

	// day and date are in the same string
	if dateRe.MatchString(rawData[0].String()) {
		m.Date = Date(dateRe.FindString(rawData[0].String()))
		m.Day = rawData[0].Split(" ")[0]
		return rawData[1:]
	} else if dateRe.MatchString(rawData[1].String()) {
		// date is the second string
		m.Day = rawData[0].String()
		// pivnice u capa has unnecessary whitespaces
		m.Date = Date(strings.ReplaceAll(dateRe.FindString(rawData[1].String()), " ", ""))
		return rawData[2:]
	} else if dayRe.MatchString(rawData[0].ToLower()) {
		m.Day = rawData[0].ReplaceAll(" ", "")
		m.Date.getDateFromWeekDay(m.Day)
		return rawData[1:]
	} else {
		log.Print("Unable to parse date and day from menu.")
		return rawData
	}
}

// processLines processes lines to better format so each meal is for 1 line
func (m *Menu) processLines(data []Line) {
	var line Line

	// Menu is missing
	if len(data) < 5 {
		m.Lines = data[:]
		return
	}

	for i, rawLine := range data {
		line.append(" " + rawLine.TrimSpace())

		if len(rawLine) > bigLineLimit && mustSplit(i+1, data) || rawLine.containsPrice() {
			m.Lines = append(m.Lines, line[:])
			line = ""
		}
	}
}

// mustSplit Split based on length of line or position of line with price (kc)
func mustSplit(start int, data []Line) bool {
	for i := start; i < len(data); i++ {
		line := data[i]

		if line.containsPrice() {
			return false
		}

		if len(line) > bigLineLimit {
			return true
		}
	}
	return false
}

func (d *Date) getDateFromWeekDay(weekDay string) {
	dayToday := int(time.Now().Weekday()) - 1
	// move sunday to last day
	if dayToday == -1 {
		dayToday = 6
	}

	for i, day := range czechDayNames {
		if strings.ToLower(weekDay) == day {
			*d = Date(time.Now().AddDate(0, 0, i-dayToday).Format("02.01.2006"))
			return
		}
	}
	return
}

func (d *Date) String() string {
	return string(*d)
}

func (l *Line) String() string {
	return string(*l)
}

func (l *Line) Split(char string) []string {
	return strings.Split(l.String(), char)
}

func (l *Line) ToLower() string {
	return strings.ToLower(l.String())
}

func(l *Line) ReplaceAll(old, new string) string {
	return strings.ReplaceAll(l.String(), old, new)
}

func (l *Line) ReplaceAllToLower(old, new string) string {
	return strings.ToLower(l.ReplaceAll(old, new))
}

func (l *Line) TrimSpace() string {
	return strings.TrimSpace(l.String())
}

// containsPrice checks the line if it contains a czech currency kc
func (l *Line) containsPrice() bool {
	return strings.Contains(l.ToLower(), "kč")
}

func (l *Line) append(suffix string) {
	*l = Line(l.String() + suffix)
}
