package main

import (
	"fmt"

	"github.com/tenkoh/japanese-webcomic/scraper"
)

// ComiWalkerEndpoint : endpoint for comic walker
const ComiWalkerEndpoint = "https://comic-walker.com/contents/list/?p=1"

func main() {
	CwComics := scraper.ComicWalkerScrape(ComiWalkerEndpoint)
	fmt.Println(CwComics)
}
