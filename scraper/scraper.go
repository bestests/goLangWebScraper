package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

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

func getPages(url string) int {

	pages := 0

	res, err := http.Get(url)

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

func getPage(url string, page int, c chan<- []extractedJob) {

	var jobs []extractedJob

	pageURL := url + "&start=" + strconv.Itoa(page*50)

	fmt.Println("Requesting : ", pageURL)

	res, err := http.Get(pageURL)

	chkErr(err)
	chkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	chkErr(err)

	cards := doc.Find(".jobsearch-SerpJobCard")

	jobc := make(chan extractedJob)

	cards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, jobc)
	})

	for i := 0; i < cards.Length(); i++ {

		job := <-jobc

		jobs = append(jobs, job)
	}

	c <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := card.Find(".title>a").Text()
	company, comErr := card.Find(".sjcl>div>.company>a").Html()

	if comErr != nil {
		company = card.Find(".sjcl>div>.company").Text()
	} else {
		company = card.Find(".sjcl>div>.company>a").Text()
	}

	location := card.Find(".sjcl>.location").Text()

	salary, salErr := card.Find(".salarySnippet>.salary>.salaryText").Html()

	if salErr != nil {
		salary = ""
	}

	summary := card.Find(".summary").Text()

	c <- extractedJob{
		id:       CleanString(id),
		title:    CleanString(title),
		company:  CleanString(company),
		location: CleanString(location),
		salary:   CleanString(salary),
		summary:  CleanString(summary),
	}
}

// CleanString - clean string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func writeJobs(jobs []extractedJob, fileNm string) {
	file, err := os.Create(fileNm)

	chkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"LINK", "TITLE", "COMPANY", "LOCATION", "SALARY", "SUMMARY"}

	wErr := w.Write(headers)

	chkErr(wErr)

	writeC := make(chan []string)

	for _, job := range jobs {
		go jobSlice(job, writeC)
	}

	for i := 0; i < len(jobs); i++ {
		jwErr := w.Write(<-writeC)

		chkErr(jwErr)
	}

	fmt.Println("DONE, extracted", len(jobs))
}

func jobSlice(job extractedJob, writeC chan<- []string) {
	writeC <- []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.company, job.location, job.salary, job.summary}
}

// Scrape - Scrape indeed site
func Scrape(search string, fileNm string) {

	var baseURL = "https://kr.indeed.com/jobs?q=" + search + "&limit=50"

	var jobs []extractedJob

	totalPages := getPages(baseURL)

	c := make(chan []extractedJob)

	for page := 0; page < totalPages; page++ {
		go getPage(baseURL, page, c)
	}

	for page := 0; page < totalPages; page++ {

		extractJob := <-c

		jobs = append(jobs, extractJob...)
	}

	writeJobs(jobs, fileNm)
}
