// Headless web scraper using Playwright for Go
package gejie

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/playwright-community/playwright-go"
)

const zhipinBaseUrl = "https://www.zhipin.com"

// const exampleMercadoLibreXiaoMi15 = "https://listado.mercadolibre.com.pe/xiaomi-15"

type JobPosting struct {
	PageTitle    string
	JobPostTitle string
	SalaryRange  string
	Content      string
	Tags         []string
	Url          string
	ScrapedAt    time.Time
}

func RunZhipin(firstUrl string, collectLinks bool) {

	// directly create url frontier for now
	urlFrontier := NewURLFrontier()

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser (visible by default for development)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create new browser context and page
	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	scrapedJobPostings := []JobPosting{}
	urlFrontier.Add(firstUrl)
	jp := ScrapePageUrl(page, firstUrl)
	scrapedJobPostings = append(scrapedJobPostings, jp)
	urlFrontier.MarkVisited(firstUrl)

	if collectLinks {
		collectMoreLinks(page, urlFrontier)
	}

	// loop until no more urls to visit
	url, hasMore := urlFrontier.GetNext()
	pagesScrapped := 1
	const scrapePageLimit = 10
	for hasMore && pagesScrapped <= scrapePageLimit {
		// Add random delay between 1-3 seconds
		delay := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
		time.Sleep(delay)

		fullUrl := fmt.Sprintf("%s%s", zhipinBaseUrl, url)
		log.Printf("Scraping URL: %s", fullUrl)
		jp = ScrapePageUrl(page, fullUrl)
		scrapedJobPostings = append(scrapedJobPostings, jp)

		urlFrontier.MarkVisited(url)
		pagesScrapped++
		url, hasMore = urlFrontier.GetNext()
		if hasMore {
			log.Printf("has more urls (%d) to visit\n\n", urlFrontier.CountRemaining())
		} else {
			log.Print("no more urls left")
		}
	}

	fmt.Print(scrapedJobPostings)
	// Add small delay to observe results
	time.Sleep(2 * time.Second)
}

func ScrapePageUrl(page playwright.Page, url string) JobPosting {

	// Example: Navigate to a page and scrape title
	if _, err := page.Goto(url); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	// Wait for page to load
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	// Get page title
	pageTitle, err := page.Title()
	if err != nil {
		log.Fatalf("could not get title: %v", err)
	}

	jobPostTitleElem := page.Locator("#main > div.job-banner > div > div > div.info-primary > div.name > h1")
	playwright.NewPlaywrightAssertions(10000).Locator(jobPostTitleElem).ToBeVisible()

	// get the job post name
	jobPostTitleText, err := jobPostTitleElem.First().TextContent()
	if err != nil {
		log.Fatalf("could not exctract job posting title")
	}

	salaryRangeText, err := page.Locator("#main > div.job-banner > div > div > div.info-primary > div.name > span").First().TextContent()
	if err != nil {
		log.Fatalf("could not extract job salary range")
	}

	jobPostContent, err := page.Locator("#main > div.job-box > div > div.job-detail > div:nth-child(1) > div.job-sec-text").First().TextContent()
	if err != nil {
		log.Fatalf("could not extract job salary range")
	}

	postTagElems, err := page.Locator("#main > div.job-box > div > div.job-detail > div:nth-child(1) > ul > li").AllTextContents()
	if err != nil {
		log.Fatalf("could not extract job posting tags")
	}

	// log.Println("post tags", postTagElems)

	newJobPosting := JobPosting{
		PageTitle:    pageTitle,
		JobPostTitle: jobPostTitleText,
		Url:          url,
		SalaryRange:  salaryRangeText,
		Content:      jobPostContent,
		Tags:         postTagElems,
		ScrapedAt:    time.Now(),
	}

	// take screenshot
	// if _, err := page.Screenshot(playwright.PageScreenshotOptions{
	// 	Path: playwright.String("./screenshots/jd-screenshot.png"),
	// }); err != nil {
	// 	log.Printf("could not take screenshot: %v", err)
	// }

	// collect more job links
	log.Printf("got job content text: %s", newJobPosting)

	return newJobPosting
}

func collectMoreLinks(page playwright.Page, urlFrontier URLFrontierInterface) {
	moreJobsListSelector := "ul.look-job-list"
	moreJobsListElems, err := page.Locator(moreJobsListSelector).All()
	if err != nil {
		log.Fatalf("could not extract recommended job postings")
	}

	var moreJobUrls []string
	for _, elem := range moreJobsListElems {
		jobLinks, err := elem.Locator("li > a").All()
		if err != nil {
			log.Printf("error getting text for element %v", err)
			continue
		}
		for _, jl := range jobLinks {
			linkUrl, err := jl.GetAttribute("href")
			if err != nil {
				log.Fatalf("could not parse job listing url")
				continue
			}
			// log.Printf("\njob listing url %s", linkUrl)
			moreJobUrls = append(moreJobUrls, linkUrl)
		}
	}

	log.Println("extra job links: ", moreJobUrls)
	urlFrontier.BulkAdd(moreJobUrls)
	log.Printf("\nTotal urls in frontier: %d", urlFrontier.Count())
}

// const exampleSearchViewUrl = "https://www.zhipin.com/web/geek/jobs?query=agents&city=101010100"

// FromSearchView crawls the Zhipin job postings from the search query view
func FromSearchView(term string) {
	log.Printf("collect job posts from search term: %s", term)
}
