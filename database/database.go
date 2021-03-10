package database

import (
	"log"

	"github.com/guregu/dynamo"
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
func WriteItems(table dynamo.Table, items []interface{}) error {
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
