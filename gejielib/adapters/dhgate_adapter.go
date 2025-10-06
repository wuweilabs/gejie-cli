package adapters

import (
	"log"
	"log/slog"

	"github.com/playwright-community/playwright-go"
	br "github.com/zshanhui/gejiezhipin/gejielib/browser"
	"github.com/zshanhui/gejiezhipin/gejielib/meli"
)

var exampleSearchUrl = "https://www.dhgate.com/wholesale/search.do?act=search&dspm=pcen.sp.searclick.1.VtSiWHStSG6tOdKrhnni%26resource_id%3D&sus=&searchkey=keyboards&catalog=#pusearch1812"

func blockIndexMediaFiles(ctx playwright.BrowserContext) {
	// Speed up navigation: block heavy resources not needed for scraping links
	_ = ctx.Route("**/*", func(route playwright.Route) {
		rt := route.Request().ResourceType()
		switch rt {
		case "image", "media", "font":
			_ = route.Abort()
		default:
			_ = route.Continue()
		}
	})
}

func RunDhgateSearch(searchUrl *string, opts br.CmdOptions) []meli.MeliProduct {

	// lets implement the manual way first, then think about DRY'ing it up later with generalized pattern

	if searchUrl != nil {
		searchUrl = &exampleSearchUrl
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browserOpts := br.DefaultBrowserOptions()
	browserOpts.Headless = false
	if opts.HeadlessMode {
		browserOpts.Headless = true
	}

	nb, err := br.CreateBrowser(pw, browserOpts, opts)
	if err != nil {
		log.Fatalf("could not create new browser agent: %v", err)
	}
	defer nb.Close()

	ctx, err := nb.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}
	defer ctx.Close()

	blockIndexMediaFiles(ctx)

	searchIndexPage, err := ctx.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	defer searchIndexPage.Close()

	// Navigate and wait only for DOMContentLoaded to avoid long waits for lazy resources
	_, err = searchIndexPage.Goto(*searchUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		log.Fatalf("failed to navigate: %v", err)
	}

	// finally we can do stuff
	slog.Info("\nDhgate search index page loaded, proceed to scrape links...")

	// TODO: think about how to handle ads and endorsed product links

	// !the pagination for dhgate keeps going even though the product ends after a couple hundred for most categories, limit to 250 for now to prevent infinite loops
	if opts.MaxItems > 250 {
		opts.MaxItems = 250
	}
	linkOpts := ScrapeProductLinkPageOpts{
		MaxItems:               int(opts.MaxItems),
		ProductLinkSelector:    "div.product-list-warp div.product-item",
		PaginationNextSelector: "span.next-btn",
	}
	productLinks := ScrapeProductLinksWithPagination(searchIndexPage, linkOpts)

	slog.Info("total product links scraped", "count", len(productLinks), "firstLink", productLinks[0], "lastLink", productLinks[len(productLinks)-1])

	return []meli.MeliProduct{}
}

func ScrapeDhgateProductPage(br playwright.Browser, url string) *meli.MeliProduct {

	return &meli.MeliProduct{}
}
