package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/tenkoh/japanese-webcomic/scraper"
)

// ComiWalkerEndpoint : endpoint for comic walker
const ComiWalkerEndpoint = "https://comic-walker.com/"

func main() {
	// CwComics := scraper.ComicWalkerScraper(ComiWalkerEndpoint)
	// f, _ := os.Create("tmp.json")
	// defer f.Close()

	// if err := json.NewEncoder(f).Encode(CwComics); err != nil {
	// 	log.Fatal(err)
	// }

	// for debug
	// CwComics, _ := scraper.FromJSON("tmp.json")

	// // example
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
	table := db.Table("ComicUpdateRecord")
	// var wg sync.WaitGroup

	// for _, c := range CwComics {
	// 	wg.Add(1)
	// 	go func(comic *scraper.Comic) {
	// 		defer wg.Done()
	// 		err := table.Put(comic).Run()
	// 		if err != nil {
	// 			fmt.Println("Could not put to DynamoDB")
	// 			log.Fatal(err)
	// 		}
	// 	}(c)
	// }
	// wg.Wait()

	var filtered []scraper.Comic
	err := table.Scan().Filter("'Publisher' = ?", "FLOS COMIC").All(&filtered)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, c := range filtered {
		fmt.Printf("%+v\n", c)
	}

}
