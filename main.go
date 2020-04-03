package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var baseURL = "https://kr.indeed.com/jobs?q=python&limit=50"

func chkErr(err error) {

	if err != nil {
		log.Fatalln(err)
	}
}

func chkCode(res *http.Response) {

	if res.StatusCode != 200 {
		log.Fatalln("Response FAILED CODE - ", res.StatusCode)
	}
}

func getPages() int {

	pages := 0

	res, err := http.Get(baseURL)

	chkErr(err)
	chkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	chkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func getPage(page int) {
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)

	fmt.Println("PAGE URL : ", pageURL)
}

func main() {
	totalPages := getPages()

	for page := 0; page < totalPages; page++ {
		getPage(page)
	}
}
