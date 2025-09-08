// go Colly example
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/cli"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
	"github.com/zshanhui/gejiezhipin/gejielib/zhipin"
	utils "github.com/zshanhui/gejiezhipin/utils"
)

const devMode int8 = 0

func main() {
	if devMode == 1 {
		fmt.Print("dev mode enabled\n")
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

	opts := gejie.CmdOptions{
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
					fmt.Printf("retrieved maxItems: %d\n", maxItems)
				} else {
					fmt.Printf("invalid maxItems value: %s - using default value 2.\n", os.Args[i+1])
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
		bm, _ := gejie.NewBrowserManager(gejie.DefaultBrowserOptions())
		page, _ := bm.NewPage()
		page.Goto(searchUrlPe, playwright.PageGotoOptions{
			Timeout: playwright.Float(8000),
		})

		fmt.Print("test scraping product links...\n")
		urlFrontier := gejie.ScrapeProductLinksWithPagination(page, opts.MaxItems)
		fmt.Printf("Total links scraped: %d\n", urlFrontier.Count())
		
		for {
			productLink, hasNext := urlFrontier.GetNext()
			if !hasNext {
				break
			}
			fmt.Println(productLink)
		}

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-product" {
		product := gejie.ScrapeProductPageDirect(gejie.ProductUrlExample)
		utils.PrintProduct(product)

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-product-images" {
		images := gejie.ScrapeProductImages(nil, gejie.ProductUrlExample)
		fmt.Print("extracted images: ", images)

	} else if len(os.Args) > 1 && os.Args[1] == "--meli-store" {
		panic("not implemented")

	} else {
		fmt.Printf("command not recognized: %s", os.Args[1])
		os.Exit(1)
	}
}
