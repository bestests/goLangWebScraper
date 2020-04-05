package main

import (
	"os"
	"strings"
	"time"

	"github.com/goLangWebScraper/scraper"
	"github.com/labstack/echo"
)

func hadleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	search := strings.ToLower(scraper.CleanString(c.FormValue("search")))

	fileNm := time.Now().Format("20060102150405") + "_indeed_" + search + ".csv"

	scraper.Scrape(search, fileNm)

	// 완료 후 삭제
	defer os.Remove(fileNm)

	return c.Attachment(fileNm, fileNm)
}

func main() {

	e := echo.New()

	e.GET("/", hadleHome)

	e.POST("/scrape", handleScrape)

	e.Logger.Fatal(e.Start(":1323"))
}
