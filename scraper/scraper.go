package scraper

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	loc := time.FixedZone("Asia/Tokyo", 9*60*60)
	var undefinedDate string = time.Date(1990, 1, 1, 0, 0, 0, 0, loc).Format(outputLayout)
	jst, err := time.ParseInLocation(inputLayout, date, loc)
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

func parseComicWalker(link string) (*Comic, error) {
	parent := "https://comic-walker.com"
	doc, err := getDoc(parent + link)
	if err != nil {
		log.Fatal(parent+link, err)
		return nil, err
	}

	title := doc.Find("#detailIndex > div > h1").Text()
	introduction := doc.Find("#summaryIntroduction > div.ac-content.ac-hidden > div > div.inner > p").Text()
	introduction = strings.Replace(introduction, "<br>", "", -1)
	introduction = strings.Replace(introduction, "\n", "", -1)
	var authors []string
	doc.Find("#summaryIntroduction > div.ac-content.ac-hidden > div > div.inner > div > a").Each(func(i int, s *goquery.Selection) {
		author := strings.Split(s.Text(), "(")[0]
		authors = append(authors, author)
	})
	var genres []string
	var publisher string
	doc.Find("#genre > div.container > div > h3").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			publisher = s.Text()
		default:
			genres = append(genres, s.Text())
		}
	})
	latest := doc.Find("#detailIndex > div > div > div > div > div.titleBox > p.comicIndex-title").Text()
	lu := doc.Find("#detailIndex > div > div > div > div > div.dateBox > span.comicIndex-date").Text()
	lastUpdate := timeInJST("2006/01/02 15:04:05", "2006-01-02", strings.Split(lu, " ")[0]+" 00:00:00")
	nu := doc.Find("#detailIndex > div > div > div > div > span").Text()
	nextUpdate := timeInJST("2006/01/02 15:04:05", "2006-01-02", strings.Split(nu, "】")[1]+" 00:00:00")
	latestRecord := time.Now().In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006-01-02 15:04:05")

	c := Comic{
		Site:         "ComicWalker",
		Title:        title,
		CID:          idGenerator("ComicWalker" + title),
		Introduction: introduction,
		Link:         parent + link,
		Latest:       latest,
		LastUpdate:   lastUpdate,
		NextUpdate:   nextUpdate,
		LatestRecord: latestRecord,
		Authors:      authors,
		Publisher:    publisher,
		Genres:       genres,
	}

	return &c, nil
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

// ComicWalkerScraper : special scraper for ComicWalker
func ComicWalkerScraper(endpoint string) []*Comic {
	doc, err := getDoc(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	links := []string{}

	// get content's link
	// [MUST] When deploy, search range must be i=0 only
	doc.Find(".tileList.clearfix").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			s.Find("li").Each(func(i int, ss *goquery.Selection) {
				if ss.Find(".icon-latest").Nodes != nil {
					link, _ := ss.Find("a").Attr("href")
					links = append(links, link)
				}
			})
		}
	})

	links = removeDuplicateKeepFirst(links)
	comics := SequentialScraper(links, parseComicWalker, 10)
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
