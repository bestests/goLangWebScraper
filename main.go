package main

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

func getPage(page int) []extractedJob {

	var jobs []extractedJob

	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)

	fmt.Println("Requesting : ", pageURL)

	res, err := http.Get(pageURL)

	chkErr(err)
	chkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	chkErr(err)

	doc.Find(".jobsearch-SerpJobCard").Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)

		jobs = append(jobs, job)
	})

	return jobs
}

func extractJob(card *goquery.Selection) extractedJob {
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

	result := extractedJob{
		id:       cleanString(id),
		title:    cleanString(title),
		company:  cleanString(company),
		location: cleanString(location),
		salary:   cleanString(salary),
		summary:  cleanString(summary),
	}

	return result
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")

	chkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"ID", "TITLE", "COMPANY", "LOCATION", "SALARY", "SUMMARY"}

	wErr := w.Write(headers)

	chkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.company, job.location, job.salary, job.summary}

		jwErr := w.Write(jobSlice)

		chkErr(jwErr)
	}

	fmt.Println("DONE, extracted", len(jobs))
}

func main() {

	var jobs []extractedJob

	totalPages := getPages()

	for page := 0; page < totalPages; page++ {
		extractJob := getPage(page)
		jobs = append(jobs, extractJob...)
	}

	writeJobs(jobs)
}
