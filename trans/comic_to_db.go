package trans

import (
	"fmt"

	"github.com/tenkoh/japanese-webcomic/scraper"
)

// DynamoComic : suitable struct for storing Comic to DynamoDB
type DynamoComic struct {
	CID        string
	DataType   string
	DataValue  string
	Publisher  string `dynamo:",omitempty"`
	Title      string `dynamo:",omitempty"`
	Link       string `dynamo:",omitempty"`
	Latest     string `dynamo:",omitempty"`
	LastUpdate string `dynamo:",omitempty"`
	NextUpdate string `dynamo:",omitempty"`
}

// DynamoMarshal : convert Comic to DynamoComic struct
func DynamoMarshal(c *scraper.Comic) []*DynamoComic {
	dcomics := []*DynamoComic{}

	// Required for detail of each content
	// Required for user fav page
	// Required for TopPage
	dcomics = append(dcomics, &DynamoComic{
		CID:       c.CID,
		DataType:  "content#fav#mainlist",
		DataValue: fmt.Sprintf("%s %s %s", c.LastUpdate, c.LatestRecord, c.NextUpdate),
		Publisher: c.Publisher,
		Title:     c.Title,
		Link:      c.Link,
		Latest:    c.Latest,
	})

	// Required for detail of each content
	dcomics = append(dcomics, &DynamoComic{
		CID:       c.CID,
		DataType:  "content#introduction",
		DataValue: c.Introduction,
	})

	// Required for detail of each content
	// Required for search by author
	if len(c.Authors) > 0 {
		ath := []*DynamoComic{}
		for _, a := range c.Authors {
			ath = append(ath, &DynamoComic{
				CID:       c.CID,
				DataType:  fmt.Sprintf("content#ath_%s", a),
				DataValue: a,
				Title:     c.Title,
			})
		}
		dcomics = append(dcomics, ath...)
	}

	// Required for detail of each content
	// Required for search by genre
	if len(c.Genres) > 0 {
		gen := []*DynamoComic{}
		for _, g := range c.Genres {
			gen = append(gen, &DynamoComic{
				CID:       c.CID,
				DataType:  fmt.Sprintf("content#gen_%s", g),
				DataValue: g,
				Title:     c.Title,
			})
		}
		dcomics = append(dcomics, gen...)
	}

	// Just required for search by publisher
	dcomics = append(dcomics, &DynamoComic{
		CID:       c.CID,
		DataType:  "publisher",
		DataValue: c.Publisher,
		Title:     c.Title,
	})

	return dcomics
}
