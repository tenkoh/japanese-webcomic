package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/tenkoh/japanese-webcomic/database"
	"github.com/tenkoh/japanese-webcomic/scraper"
)

const dynamoRegion = "ap-northeast-1"
const dynamoTableName = "ComicUpdateNotifier"

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

	// connect to DynamoDB
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(dynamoRegion)})
	table := db.Table(dynamoTableName)

	// write scraping result to DynamoDB
	if err := database.WriteItems(table, writer); err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
