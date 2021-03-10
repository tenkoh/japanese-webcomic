package scraper

import (
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const comiWalkerEndpoint = "https://comic-walker.com/"

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

// ComicWalkerScraper : special scraper for ComicWalker
func ComicWalkerScraper() []*Comic {
	doc, err := getDoc(comiWalkerEndpoint)
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
