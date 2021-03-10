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

// Comic : common struct of scraping result
type Comic struct {
	Site          string    `json:"site"`
	Title         string    `json:"title"`
	CID           string    `json:"cid"`
	Introduction  string    `json:"introduction"`
	Link          string    `json:"link"`
	Latest        string    `json:"latest"`
	LastUpdate    time.Time `json:"lastUpdate"`
	NextUpdate    time.Time `json:"nextUpdate,omitempty" dynamo:",omitempty"`
	LatestRecord  time.Time `json:"latestRecord"`
	SortTimeStamp string    `json:"sortTimeStamp"` //LastUpdate#LatestRecord yyyy-mm-dd#yyyy-mm-ddTHH:MM:SS
	Author        []string  `json:"author" dynamo:",set"`
	Publisher     string    `json:"publisher"`
	Genre         []string  `json:"genre" dynamo:",set"`
}

func idGenerator(s string) string {
	md5 := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", md5)
}

func sortTimeStampGenerator(t1, t2 time.Time) string {
	layout1 := "2006-01-02"
	layout2 := "2006-01-02T15:04:05"
	return fmt.Sprintf("%s#%s", t1.Format(layout1), t2.Format(layout2))
}

func getDoc(endpoint string) (*goquery.Document, error) {
	// Search main content
	res, err := http.Get(endpoint)
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

func timeInJST(layout, date string) time.Time {
	loc := time.FixedZone("Asia/Tokyo", 9*60*60)
	var undefinedDate time.Time = time.Date(1990, 1, 1, 0, 0, 0, 0, loc)
	jst, err := time.ParseInLocation(layout, date, loc)
	if err != nil {
		return undefinedDate
	}
	return jst
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
	lastUpdate := timeInJST("2006/01/02 15:04:05", strings.Split(lu, " ")[0]+" 00:00:00")
	nu := doc.Find("#detailIndex > div > div > div > div > span").Text()
	nextUpdate := timeInJST("2006/01/02 15:04:05", strings.Split(nu, "】")[1]+" 00:00:00")
	latestRecord := time.Now().In(time.FixedZone("Asia/Tokyo", 9*60*60))

	c := Comic{
		Site:          "ComicWalker",
		Title:         title,
		CID:           idGenerator("ComicWalker" + title),
		Introduction:  introduction,
		Link:          parent + link,
		Latest:        latest,
		LastUpdate:    lastUpdate,
		NextUpdate:    nextUpdate,
		LatestRecord:  latestRecord,
		SortTimeStamp: sortTimeStampGenerator(lastUpdate, latestRecord),
		Author:        authors,
		Publisher:     publisher,
		Genre:         genres,
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
	comics := Scraper(links, parseComicWalker)
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
