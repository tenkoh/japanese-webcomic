package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/tenkoh/japanese-webcomic/database"
	"github.com/tenkoh/japanese-webcomic/scraper"
)

const dynamoRegion = "ap-northeast-1"
const dynamoTableName = "ComicUpdateNotifier"

func main() {
	// scraping comics
	doScrape := false
	var CwComics []*scraper.Comic
	if doScrape {
		CwComics = scraper.ComicWalkerScraper()

		// for debug
		f, _ := os.Create("../../tmp.json")
		defer f.Close()

		if err := json.NewEncoder(f).Encode(CwComics); err != nil {
			log.Fatal(err)
		}
	} else {
		CwComics, _ = scraper.FromJSON("../../tmp.json")
	}

	// translate scraping data into suitable format of DynamoDB
	// And put it into interface for dynamo writer
	var dcomics []*scraper.DynamoComic
	for _, c := range CwComics {
		dcomics = append(dcomics, c.DynamoMarshal()...)
	}
	var writer = make([]interface{}, len(dcomics))
	for i, v := range dcomics {
		writer[i] = v
	}

	// connect to DynamoDB
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(dynamoRegion)})
	table := db.Table(dynamoTableName)

	// write scraping result to DynamoDB
	database.WriteItems(table, writer)

	// var filtered []trans.DynamoComic
	// // err := table.Get("CID", "1fea47e3674751242efe7d7e5ae2cab8").Range("DataType", dynamo.BeginsWith, "content#list").All(&filtered)
	// // if err != nil {
	// // 	log.Fatal(err)
	// // 	return
	// // }
	// // for _, c := range filtered {
	// // 	fmt.Printf("%+v\n", c)
	// // }

	// // if search with "contains" is required, use the expression below.
	// // For example: search title = .*アンゴルモア.* could be written as below.
	// err := table.Get("DataType", "content#fav#mainlist").Index("DataType-DataValue-index").Filter("contains($,?)", "Title", "アンゴルモア").Order(dynamo.Descending).All(&filtered)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// fmt.Println(filtered)
}
