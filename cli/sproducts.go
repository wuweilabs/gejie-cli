package cli

import (
	"log/slog"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cobra"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
	"github.com/zshanhui/gejiezhipin/gejielib/adapters"
	br "github.com/zshanhui/gejiezhipin/gejielib/browser"
	"github.com/zshanhui/gejiezhipin/gejielib/meli"
	"github.com/zshanhui/gejiezhipin/utils"
)

var scrapeProductsCmd = &cobra.Command{
	Use:   "scrape-products",
	Short: "scrape eCommerce product listings, store pages, and analytics",
	Long:  "scrape eCommerce product listings, store pages, and analytics",
	Run: func(cmd *cobra.Command, args []string) {
		site, _ := cmd.Flags().GetString("site")
		maxItems, _ := cmd.Flags().GetInt("max-items")
		onlyImages, _ := cmd.Flags().GetBool("only-images")
		url, _ := cmd.Flags().GetString("url")
		createCsv, _ := cmd.Flags().GetBool("create-csv")
		headlessMode, _ := cmd.Flags().GetBool("headless")
		workers, _ := cmd.Flags().GetInt("workers")
		browserStr, _ := cmd.Flags().GetString("browser")

		browserType := br.GetBrowserType(browserStr)

		commerceSite := br.ParseCommerceSite(site)

		slog.Info("scrape-products command options",
			"maxItems", maxItems,
			"onlyImages", onlyImages,
			"url", url,
			"createCsv", createCsv,
			"headless", headlessMode,
			"workers", workers,
		)
		routeSiteProductUrl(url, br.CmdOptions{
			Site:         commerceSite,
			MaxItems:     maxItems,
			OnlyImages:   onlyImages,
			CreateCsv:    createCsv,
			HeadlessMode: headlessMode,
			Workers:      workers,
			Browser:      browserType,
		})
	},
}

func routeSiteProductUrl(url string, opts br.CmdOptions) {
	if url == "" {
		slog.Error("url is empty, please provide a valid eCommerce site url")
		return
	}

	if opts.Site == br.NotSupported {
		slog.Error("url is not a valid site url we support, try: meli, dhgate", "url", url)
	}

	topLevelTimer := utils.NewTimer("Top Level")
	defer topLevelTimer.LogElapsed()

	switch opts.Site {
	case br.MercadoLibre:
		RouteMeli(url, opts)
	case br.DHGate:
		slog.Info("handle dhgate products search")
	default:
		slog.Warn("unsupported site", "url", url)
	}
}

type RouteOptions struct {
	productUrlPrefixes      []string
	listUrlPrefixes         []string
	scrapeProductImages     func(page playwright.Page, url string) []string
	scrapeProductPageDirect func(url string, opts br.CmdOptions) *meli.MeliProduct
	runSearch               func(searchUrl *string, opts br.CmdOptions) []meli.MeliProduct
}

func RouteMeli(url string, opts br.CmdOptions) {
	productUrlPrefixes := []string{"https://www.mercadolibre", "https://articulo.mercadolibre", "mercadolibre"}
	listUrlPrefixes := []string{"https://listado.mercadolibre", "listado.mercadolibre"}

	routeUrls(url, opts, RouteOptions{
		productUrlPrefixes:      productUrlPrefixes,
		listUrlPrefixes:         listUrlPrefixes,
		scrapeProductImages:     gejie.ScrapeProductImages,
		scrapeProductPageDirect: gejie.ScrapeProductPageDirect,
		runSearch:               gejie.RunMeliSearch,
	})
}

func RouteDhgate(url string, opts br.CmdOptions) {
	productUrlPrefixes := []string{"https://www.dhgate.com/product",
		"dhgate.com/product"}
	listUrlPrefixes := []string{"https://www.dhgate.com/wholesale/search.do",
		"dhgate.com/wholesale/search.do"}

	routeUrls(url, opts, RouteOptions{
		productUrlPrefixes:      productUrlPrefixes,
		listUrlPrefixes:         listUrlPrefixes,
		runSearch:               adapters.RunDhgateSearch,
		scrapeProductPageDirect: nil,
		scrapeProductImages:     nil,
	})
}

func routeUrls(url string, opts br.CmdOptions, rOpts RouteOptions) {
	isProductUrl := false
	isIndexUrl := false
	for _, productUrl := range rOpts.productUrlPrefixes {
		if strings.HasPrefix(url, productUrl) {
			isProductUrl = true
		}
	}
	for _, listUrl := range rOpts.listUrlPrefixes {
		if strings.HasPrefix(url, listUrl) {
			isIndexUrl = true
		}
	}

	// ./gejiec scrape-products --url https://articulo.mercadolibre.com.mx/MLM-1411526559-silla-gamer-reclinable-giratoria-ergonomica-super-comoda-_JM
	if isProductUrl {
		if opts.OnlyImages && rOpts.scrapeProductImages != nil {
			slog.Info("scraping only product images", "url", url)
			images := rOpts.scrapeProductImages(nil, url)
			// just print for now
			for _, image := range images {
				slog.Info(image)
			}
			return
		}

		slog.Info("scraping product url", "url", url)
		if rOpts.scrapeProductPageDirect != nil {
			product := rOpts.scrapeProductPageDirect(url, opts)
			utils.PrintProduct(product)
		}
		// ./gejiec scrape-products --url https://listado.mercadolibre.com.mx/carburador-stihl --max-items 5 --create-csv
	} else if isIndexUrl {
		if rOpts.runSearch != nil {
			slog.Info("scraping list url", "url", url)
			rOpts.runSearch(&url, opts)
		} else {
			slog.Warn("RouteOptions.runSearch not implement for this site",
				"site", opts.Site.String())
		}
	} else {
		slog.Error("not a valid url pattern for mercadolibre", "url", url)
	}
}

func init() {
	scrapeProductsCmd.Flags().Int("max-items", 10, "max items to scrape, only for product list urls")
	scrapeProductsCmd.Flags().String("url", "", "mercadolibre url to scrape - page type will be auto detected")
	scrapeProductsCmd.Flags().Bool("only-images", false, "only scrape the images from given product url, no other data will be scraped")
	scrapeProductsCmd.Flags().Bool("create-csv", false, "create a csv file of the scraped products")
	scrapeProductsCmd.Flags().Bool("headless", false, "run the browser in headless mode")
	scrapeProductsCmd.Flags().Int("workers", 2, "number of concurrent workers to scrape product pages, defaults to 2")
	scrapeProductsCmd.Flags().String("browser", "chrome", "options are chromium, firefox")
	rootCmd.AddCommand(scrapeProductsCmd)
}
