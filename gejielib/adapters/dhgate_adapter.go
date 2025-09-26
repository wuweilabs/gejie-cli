package adapters

import (
	"log"

	"github.com/playwright-community/playwright-go"
	br "github.com/zshanhui/gejiezhipin/gejielib/browser"
	"github.com/zshanhui/gejiezhipin/gejielib/meli"
)

var exampleSearchUrl = "https://www.dhgate.com/wholesale/search.do?act=search&dspm=pcen.sp.searclick.1.VtSiWHStSG6tOdKrhnni%26resource_id%3D&sus=&searchkey=keyboards&catalog=#pusearch1812"

func RunSearch(searchUrl *string, opts br.CmdOptions) []meli.MeliProduct {

	// lets implement the manual way first, then think about drying up

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

	return []meli.MeliProduct{}
}

func ScrapeProductPage(br playwright.Browser, url string) *meli.MeliProduct {

	return &meli.MeliProduct{}
}
