package scraper

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tenkoh/japanese-webcomic/localtime"
)

var genreTable = map[string][]string{
	"COMICフルール":     {"BL", "女性向け"},
	"FLOS COMIC":    {"女性向け"},
	"コミックNewtype":   {"男性向け"},
	"コミックアライブ":      {"男性向け"},
	"コミックウォーカー":     {"男性向け"},
	"コミックジーン":       {"少女向け"},
	"コミックブリッジ":      {"女性向け"},
	"コンプエース":        {"男性向け", "美少女"},
	"コンプティーク":       {"男性向け", "美少女"},
	"シルフ":           {"女性向け"},
	"ドラゴンエイジ":       {"少年向け"},
	"少年エース":         {"少年向け"},
	"月刊ブシロード":       {"少年向け"},
	"電撃だいおうじ":       {"4コマ"},
	"電撃マオウ":         {"男性向け", "美少女"},
	"電撃大王":          {"少年向け", "美少女"},
	"魔法のiらんどCOMICS": {"女性向け"},
}

// Comic : Medium format of scraping result, which is easy to understand for programmer
type Comic struct {
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
}

// DynamoComic : Final format of scraping result, which is suitable for Querying on DynamoDB
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
func (c *Comic) DynamoMarshal() []*DynamoComic {
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
		DataType:  "content#intro",
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

func idGenerator(s string) string {
	md5 := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", md5)
}

func getDoc(url string) (*goquery.Document, error) {
	// Search main content
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return doc, nil
}

func timeInJST(inputLayout, outputLayout, date string) string {
	var undefinedDate string = localtime.Date(1990, 1, 1, 0, 0, 0).Format(outputLayout)
	jst, err := localtime.Parse(date, inputLayout)
	if err != nil {
		return undefinedDate
	}
	return jst.Format(outputLayout)
}

func removeDuplicateKeepFirst(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Scraper : contentGetter shall be unique for each web site.
func Scraper(links []string, contentParser func(string) (*Comic, error)) []*Comic {
	comics := []*Comic{}
	comicChan := make(chan *Comic, len(links))
	var wg sync.WaitGroup

	// get each content
	for _, link := range links {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			comic, err := contentParser(l)
			if err != nil {
				log.Fatalln(l, err)
				return
			}
			comicChan <- comic
			return
		}(link)
	}

	go func() {
		wg.Wait()
		close(comicChan)
	}()

	for c := range comicChan {
		comics = append(comics, c)
	}

	return comics
}

// SequentialScraper : Scraper for a site with crawl-delay limit
func SequentialScraper(links []string, contentParser func(string) (*Comic, error), waitSec int) []*Comic {
	comics := []*Comic{}

	for i, link := range links {
		if i != 0 {
			// sec -> nano sec
			time.Sleep(time.Duration(waitSec * 1000000000))
		}
		comic, err := contentParser(link)
		if err != nil {
			log.Fatalln(link, err)
			continue
		}
		fmt.Println(link)
		comics = append(comics, comic)
	}

	return comics
}

// FromJSON : for debug.
func FromJSON(file string) ([]*Comic, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Could not read %s\n", file)
		return nil, err
	}

	var comics []*Comic
	if err := json.Unmarshal(b, &comics); err != nil {
		log.Fatal("Could not unmarshal json")
		return nil, err
	}

	return comics, nil
}
