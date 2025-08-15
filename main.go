// go Colly example
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/zshanhui/gejiezhipin/gejie"
	"github.com/zshanhui/gejiezhipin/sakuram"
)

const zhipinJobDetailPageEx = "https://www.zhipin.com/job_detail/2d635ccda89345b01HF72Nm1EVpZ.html"
const ClassName_jobDescription = "#main > div.job-box > div > div.job-detail > div:nth-child(1) > div.job-sec-text"
const userAgent = ""

func handleMsakuraCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: --msakura <target_url> [selector]")
		return
	}

	target := os.Args[2]
	selector := "div.plyr__video-wrapper > video"
	if len(os.Args) > 3 {
		selector = os.Args[3]
	}
	if err := sakuram.DownloadVideoFromUrl(target, selector, false); err != nil {
		fmt.Printf("msakura error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--zhipin" {
		gejie.RunZhipin("https://www.zhipin.com/job_detail/b6840d4438ff55c41n1609S-FFVT.html", true)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--meli" {
		maxItems := 10 // default value
		// Check for --max-items flag
		for i, arg := range os.Args {
			if arg == "--max-items" && i+1 < len(os.Args) {
				if parsed, err := strconv.Atoi(os.Args[i+1]); err == nil {
					maxItems = parsed
				} else {
					fmt.Printf("Invalid maxItems value: %s. Using default value 10.\n", os.Args[i+1])
				}
				break
			}
		}
		gejie.RunMeliSearch(nil, int8(maxItems))
		return
	}

	// https://articulo.mercadolibre.com.mx/MLM-1304950718-carburador-de-motosierra-para-stihl-021-023-025-210-230-250-_JM

	// if len(os.Args) > 1 && os.Args[1] == "--msakura" {
	// 	handleMsakuraCommand()
	// 	return
	// }

	// Default Colly implementation
	// c := colly.NewCollector(
	// 	colly.Debugger(&debug.LogDebugger{}),
	// )

	// className := "#notify-bar > h1"
	// c.OnHTML(className, func(el *colly.HTMLElement) {
	// 	fmt.Printf("scraping job details main content from standalone page, Text='%s'\n", el.Text)
	// 	scrapeJobDetailMainContent(el)
	// })

	// c.OnResponse(func(r *colly.Response) {
	// 	// r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	// 	// r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	// 	// r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
	// 	// r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
	// 	// r.Headers.Set("Connection", "keep-alive")
	// 	// r.Headers.Set("Upgrade-Insecure-Requests", "1")
	// 	fmt.Println("OnResponse: got response...", string(r.Body))
	// })

	// c.OnError(func(r *colly.Response, err error) {
	// 	fmt.Printf("something went wrong: %s", err)
	// })

	// c.Visit("https://example.com")

	// c.OnScraped(func(r *colly.Response) {
	// 	fmt.Println("scraping finished", r.Request.URL)
	// })

	// fmt.Printf("crawling finished...")
	// for url := range visitedSites {
	// 	fmt.Printf("visited: %s\n", url)
	// }
}

// func scrapeJobDetailMainContent(el *colly.HTMLElement) {
// 	contentText := el.Text
// 	fmt.Println(contentText)
// }
