package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/tenkoh/japanese-webcomic/scraper"
)

// ComiWalkerEndpoint : endpoint for comic walker
const ComiWalkerEndpoint = "https://comic-walker.com/"

func main() {
	CwComics := scraper.ComicWalkerScraper(ComiWalkerEndpoint)
	f, _ := os.Create("tmp.json")
	defer f.Close()

	if err := json.NewEncoder(f).Encode(CwComics); err != nil {
		log.Fatal(err)
	}
}
