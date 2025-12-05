// go Colly example
package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/cli"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
	br "github.com/zshanhui/gejiezhipin/gejielib/browser"
	"github.com/zshanhui/gejiezhipin/gejielib/zhipin"
	utils "github.com/zshanhui/gejiezhipin/utils"
)

const devMode int8 = 0

func main() {
	if devMode == 1 {
		slog.Info("dev mode enabled")
		devTest()
	} else {
		cli.Execute()
	}
}

func devTest() {
	// used to test during development
	if len(os.Args) > 1 && os.Args[1] == "--zhipin" {
		zhipin.RunGejie("https://www.zhipin.com/job_detail/b6840d4438ff55c41n1609S-FFVT.html", true)
		return
	}

	opts := br.CmdOptions{
		MaxItems:     10,
		OnlyImages:   false,
		CreateCsv:    false,
		HeadlessMode: false,
	}

	if len(os.Args) > 1 && os.Args[1] == "--meli-search" {
		searchUrlPe := "https://listado.mercadolibre.com.pe/teclado-mecanico"
		// searchURlMx := "https://listado.mercadolibre.com.mx/teclado-inalambrico"
		maxItems := 2 // default value
		// Check for --max-items flag
		for i, arg := range os.Args {
			if arg == "--max-items" && i+1 < len(os.Args) {
				if parsed, err := strconv.Atoi(os.Args[i+1]); err == nil {
					maxItems = parsed
					slog.Info("retrieved maxItems", "value", maxItems)
				} else {
					slog.Error("invalid maxItems value - using default value 2", "value", os.Args[i+1])
				}
				break
			}
		}
		products := gejie.RunMeliSearch(&searchUrlPe, opts)
		for _, product := range products {
			utils.PrintProduct(&product)
		}

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-product-links" {
		searchUrlPe := "https://listado.mercadolibre.com.pe/teclado-mecanico"
		bm, _ := br.NewBrowserManager(br.DefaultBrowserOptions(), opts)
		page, _ := bm.NewPage()
		page.Goto(searchUrlPe, playwright.PageGotoOptions{
			Timeout: playwright.Float(8000),
		})

		slog.Info("test scraping product links...")
		productLinks := gejie.ScrapeProductLinksWithPaginationHelper(page, opts.MaxItems)
		for _, productLink := range productLinks {
			slog.Info("product link", "url", productLink)
		}

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-product" {
		product := gejie.ScrapeProductPageDirect(gejie.ProductUrlExample, opts)
		utils.PrintProduct(product)

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-product-images" {
		images := gejie.ScrapeProductImages(nil, gejie.ProductUrlExample)
		slog.Info("extracted images", "images", images)

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-store" {
		panic("not implemented")

	} else {
		slog.Error("command not recognized", "arg", os.Args[1])
		os.Exit(1)
	}
}
