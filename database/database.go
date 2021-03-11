package database

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

const dynamoRegion = "ap-northeast-1"
const dynamoTableName = "ComicUpdateNotifier"

var (
	db    = dynamo.New(session.New(), &aws.Config{Region: aws.String(dynamoRegion)})
	table = db.Table(dynamoTableName)
)

// WriteError : Error in putting itemes to DynamoDB
type WriteError struct{}

func (e *WriteError) Error() string {
	return "Fail to write items"
}

// WriteComics : async writer of comics to DynamoDB
// func WriteComics(cs []*scraper.Comic, table dynamo.Table) {
// 	var wg sync.WaitGroup
// 	for _, c := range cs {
// 		dcomics := trans.DynamoMarshal(c)
// 		for _, dc := range dcomics {
// 			wg.Add(1)
// 			go func(d *trans.DynamoComic) {
// 				defer wg.Done()
// 				err := table.Put(d).Run()
// 				if err != nil {
// 					log.Fatalf("Fail to put %s: %s\n", d.CID, d.DataType)
// 				}
// 			}(dc)
// 		}
// 	}
// 	wg.Wait()
// }

// WriteItems : Batch writer to DynamoDB
func WriteItems(items []interface{}) error {
	var wcc dynamo.ConsumedCapacity
	wrote, err := table.Batch().Write().Put(items...).ConsumedCapacity(&wcc).Run()
	if wrote != len(items) {
		log.Fatalf("unexpected wrote: operation count: %d not eq. to the expected: %d", wrote, len(items))
		return &WriteError{}
	}
	if err != nil {
		log.Fatalln("unexpected error:", err)
		return &WriteError{}
	}
	if wcc.Total == 0 {
		log.Fatalln("bad consumed capacity", wcc)
		return &WriteError{}
	}
	return nil
}

// QueryDateWithLimit : Descending order list with range
// ex) If you want to get 2021-03-10, begin = 2020-03-10, end = 2020-03-11 (+1Day)
func QueryDateWithLimit(begin, end string, limit int64) *dynamo.Query {
	return table.Get("DataType", "content#fav#mainlist").Range("DataValue", dynamo.Between, begin, end).Index("DataType-DataValue-index").Order(dynamo.Descending).Limit(limit)
}
