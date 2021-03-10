package trans

import "strings"

// Content : Comic + AdLink
type Content struct {
	Site         string   `json:"site"` //just used to generate CID (hash of site + title)
	Title        string   `json:"title"`
	CID          string   `json:"cid"`
	Introduction string   `json:"introduction"`
	Link         string   `json:"link"`
	Latest       string   `json:"latest"` //ex: story1-1
	LastUpdate   string   `json:"lastUpdate"`
	NextUpdate   string   `json:"nextUpdate,omitempty"`
	LatestRecord string   `json:"latestRecord"`
	Authors      []string `json:"author"`
	Publisher    string   `json:"publisher"`
	Genres       []string `json:"genre"`
	AdLink       []string `json:"adLink"`
}

// DynamoUnmarshal : translate dynamo record to readable comic struct
func DynamoUnmarshal(ds []*DynamoComic) *Content {
	c := Content{}
	for _, d := range ds {
		switch d.DataType {
		case "content#fav#mainlinst":
			// d.DataValue = LastUpdate LastRecord(date) LastRecord(time) NextUpdate
			records := strings.Split(d.DataValue, " ")
			c.LastUpdate = records[0]
			c.NextUpdate = records[3]
			if c.NextUpdate == "1990-01-01" {
				c.NextUpdate = ""
			} // default value of scraping result
			c.LatestRecord = strings.Join(records[1:3], " ")
		}
	}
	return &c
}
