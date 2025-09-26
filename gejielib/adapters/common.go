package adapters

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/utils"
)

// common scrape funcs for all eCommerce sites

func ScrapeSinglePageProductLinks(page playwright.Page, productLinkSelector string) ([]string, error) {
	productLinksTimer := utils.NewTimer("Scrape Per Page Product Links")
	defer productLinksTimer.LogElapsed()

	productLinks, err := page.Locator(productLinkSelector).All()
	if err != nil {
		return []string{}, fmt.Errorf("could not extract product links: %w", err)
	}

	productLinkUrls := []string{}
	baseURL, _ := url.Parse(page.URL())
	for _, productLink := range productLinks {
		linkUrl, err := productLink.GetAttribute("href")

		// skip click1 links since they are not product links
		if strings.HasPrefix(linkUrl, "https://click1") {
			continue
		}
		if err != nil {
			log.Fatalf("could not parse product link url")
			continue
		}
		if linkUrl == "" {
			continue
		}

		ref, err := url.Parse(linkUrl)
		if err != nil {
			// skip malformed URLs
			continue
		}
		abs := baseURL.ResolveReference(ref)
		// strip query params and fragments
		abs.RawQuery = ""
		abs.Fragment = ""

		productLinkUrls = append(productLinkUrls, abs.String())
	}

	return productLinkUrls, nil
}

func ScrapeProductLinksWithPagination(page playwright.Page, opts ScrapeProductLinkPageOpts) []string {
	allProductLinks := []string{}
	currentPage := 1

	for len(allProductLinks) < opts.MaxItems {
		slog.Info("scraping page", "page", currentPage)

		curPageProductLinks, err := ScrapeSinglePageProductLinks(page, opts.ProductLinkSelector)
		if err != nil {
			slog.Error("error scraping product links", "error", err)
			return []string{}
		}
		slog.Info("found product links on page", "count", len(curPageProductLinks), "page", currentPage)
		if len(curPageProductLinks) == 0 {
			slog.Error("no product links found on page, stopping pagination")
			return []string{}
		}

		remainingItems := opts.MaxItems - len(allProductLinks)
		if len(curPageProductLinks) <= remainingItems {
			allProductLinks = append(allProductLinks, curPageProductLinks...)
		} else {
			allProductLinks = append(allProductLinks, curPageProductLinks[:remainingItems]...)
		}

		if len(allProductLinks) >= opts.MaxItems {
			slog.Info("reached max items, stopping pagination", "maxItems", opts.MaxItems)
			break
		}

		// next page
		nextButton := page.Locator(string(opts.PaginationNextSelector))
		nextExists, err := nextButton.Count()
		if err != nil {
			log.Printf("error checking next page button: %v", err)
			break
		}
		if nextExists == 0 {
			slog.Info("no more pages available, stopping pagination")
			break
		}

		// click next button
		err = nextButton.Click()
		if err != nil {
			log.Printf("error clicking next page button: %s", err)
			break
		}

		err = page.Locator(string(opts.ProductLinkSelector)).Last().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateAttached,
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			log.Printf("error waiting for next page to load: %v", err)
			break
		}

		// Sleep for 1-3 seconds between pages
		coupleSecs := 1 + rand.Intn(3)
		slog.Info("sleeping between pages", "seconds", coupleSecs)
		time.Sleep(time.Duration(coupleSecs) * time.Second)
		currentPage++
	}

	slog.Info("total products links scraped", "pages", currentPage, "count", len(allProductLinks))
	return allProductLinks
}

type ScrapeProductLinkPageOpts struct {
	MaxItems               int
	ProductLinkSelector    string
	PaginationNextSelector string
}
