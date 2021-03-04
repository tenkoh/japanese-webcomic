package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Comic : common struct of scraping result
type Comic struct {
	Site         string    `json:"site"`
	Title        string    `json:"title"`
	Introduction string    `json:"introduction"`
	Link         string    `json:"link"`
	Latest       string    `json:"latest"`
	LastUpdate   time.Time `json:"lastUpdate"`
	NextUpdate   time.Time `json:"nextUpdate,omitempty"`
	Author       []string  `json:"author"`
	Publisher    string    `json:"publisher"`
	Genre        []string  `json:"genre"`
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
	loc, _ := time.LoadLocation("Asia/Tokyo")
	var undefinedDate time.Time = time.Date(1990, 1, 1, 0, 0, 0, 0, loc)
	jst, err := time.ParseInLocation(layout, date, loc)
	if err != nil {
		return undefinedDate
	}
	return jst
}

func eachComicWalkerContent(link string) (*Comic, error) {
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
	fmt.Println(nu)
	nextUpdate := timeInJST("2006/01/02 15:04:05", strings.Split(nu, "】")[1]+" 00:00:00")

	c := Comic{
		Site:         "ComicWalker",
		Title:        title,
		Introduction: introduction,
		Link:         parent + link,
		Latest:       latest,
		LastUpdate:   lastUpdate,
		NextUpdate:   nextUpdate,
		Author:       authors,
		Publisher:    publisher,
		Genre:        genres,
	}

	return &c, nil
}

// ComicWalkerScrape : special scraper for ComicWalker
func ComicWalkerScrape(endpoint string) []*Comic {
	doc, err := getDoc(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	comics := []*Comic{}
	links := []string{}

	// get content's link
	doc.Find("#mainContent > div > dl > dd > ul > li").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Find("a").Attr("href")
		links = append(links, link)
	})

	// get each content
	// [ToDo] use goroutine
	for _, link := range links {
		c, err := eachComicWalkerContent(link)
		if err != nil {
			continue
		}
		comics = append(comics, c)
	}

	for _, c := range comics {
		fmt.Println(*c)
	}

	return comics

	// find new title in the contents

	// if new, get comic information
}
