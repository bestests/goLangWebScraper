package main

import (
	"strings"
	"time"

	"github.com/goLangWebScraper/scraper"
	"github.com/labstack/echo"
)

func hadleHome(c echo.Context) error {
	return c.File("home.html")
}

func handelScrape(c echo.Context) error {
	search := strings.ToLower(scraper.CleanString(c.FormValue("search")))

	fileNm := time.Now().Format("20060102150405") + "_indeed_" + search + ".csv"

	scraper.Scrape(search, fileNm)

	return c.Attachment(fileNm, fileNm)
}

func main() {
	//scraper.Scrape("python")
	e := echo.New()

	e.GET("/", hadleHome)

	e.POST("/scrape", handelScrape)

	e.Logger.Fatal(e.Start(":1323"))
}
