/*
API: List for top page
contents list ordered by update date

api: /list/{page}

each page shows contents on 3 days.
*/

package main

import (
	"fmt"
	"log"

	"github.com/tenkoh/japanese-webcomic/localtime"
	"github.com/tenkoh/japanese-webcomic/response"
)

const datePerPage = 1

func dataRange(page int) (begin, end string) {
	today := localtime.Today()
	layout := "2006-01-02"
	end = today.AddDate(0, 0, 1-1*datePerPage*(page-1)).Format(layout)
	begin = today.AddDate(0, 0, 1-1*datePerPage*page).Format(layout)
	return
}

func main() {
	pagination := 2
	begin, end := dataRange(pagination)
	resp, err := response.ContentsList(begin, end, 10)
	if err != nil {
		log.Fatalln(err)
	}

	for _, r := range resp {
		fmt.Printf("%+v\n", r)
	}
}
