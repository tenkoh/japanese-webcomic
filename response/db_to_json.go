package response

import (
	"log"
	"strings"

	"github.com/tenkoh/japanese-webcomic/database"
	"github.com/tenkoh/japanese-webcomic/scraper"
)

// QueryDateError : Error in putting itemes to DynamoDB
type QueryDateError struct{}

func (e *QueryDateError) Error() string {
	return "Fail to query items by date"
}

// Content : Comic + AdLink
type Content struct {
	Site         string   `json:"site,omitempty"` //just used to generate CID (hash of site + title)
	Title        string   `json:"title"`
	CID          string   `json:"cid"`
	Introduction string   `json:"introduction"`
	Link         string   `json:"link"`
	Latest       string   `json:"latest"` //ex: story1-1
	LastUpdate   string   `json:"lastUpdate"`
	NextUpdate   string   `json:"nextUpdate"`
	LatestRecord string   `json:"latestRecord"`
	Authors      []string `json:"author"`
	Publisher    string   `json:"publisher"`
	Genres       []string `json:"genre"`
	AdLink       []string `json:"adLink"`
}

func dynamoRespToContent(d *scraper.DynamoComic) *Content {
	c := new(Content)
	c.CID = d.CID
	c.Publisher = d.Publisher
	c.Title = d.Title
	c.Link = d.Link
	c.Latest = d.Latest
	c.LastUpdate = d.DataValue

	dates := strings.Split(d.DataValue, " ")
	c.LastUpdate = dates[0]
	c.LatestRecord = strings.Join(dates[1:3], " ")
	if dates[3] == "1990-01-01" {
		c.NextUpdate = ""
	} else {
		c.NextUpdate = dates[3]
	}

	return c
}

// ContentsList : query items by date with range
// ex) If you want to get 2021-03-10, begin = 2020-03-10, end = 2020-03-11 (+1Day)
func ContentsList(begin, end string, limit int64) ([]*Content, error) {
	var resp []scraper.DynamoComic
	if err := database.QueryDateWithLimit(begin, end, limit).All(&resp); err != nil {
		log.Fatalln(err)
		return nil, &QueryDateError{}
	}
	contents := make([]*Content, len(resp))
	for i, r := range resp {
		contents[i] = dynamoRespToContent(&r)
	}
	return contents, nil
}
