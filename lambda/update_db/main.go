package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tenkoh/japanese-webcomic/database"
	"github.com/tenkoh/japanese-webcomic/scraper"
)

// HandleLambdaEvent : main function
func HandleLambdaEvent() error {
	// scraping comics
	CwComics := scraper.ComicWalkerScraper()

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

	// write scraping result to DynamoDB
	if err := database.WriteItems(writer); err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
