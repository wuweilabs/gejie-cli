package cli

import (
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
	"github.com/zshanhui/gejiezhipin/utils"
)

var meliCmd = &cobra.Command{
	Use:   "meli",
	Short: "scrape meli product listings, store pages, and analytics",
	Long:  "scrape meli product listings, store pages, and analytics",
	Run: func(cmd *cobra.Command, args []string) {
		maxItems, _ := cmd.Flags().GetInt("max-items")
		onlyImages, _ := cmd.Flags().GetBool("only-images")
		url, _ := cmd.Flags().GetString("url")
		createCsv, _ := cmd.Flags().GetBool("create-csv")
		headlessMode, _ := cmd.Flags().GetBool("headless")
		workers, _ := cmd.Flags().GetInt("workers")

		slog.Info("meli command options",
			"maxItems", maxItems,
			"onlyImages", onlyImages,
			"url", url,
			"createCsv", createCsv,
			"headless", headlessMode,
			"workers", workers,
		)
		routeMeliUrl(url, gejie.CmdOptions{
			MaxItems:     maxItems,
			OnlyImages:   onlyImages,
			CreateCsv:    createCsv,
			HeadlessMode: headlessMode,
			Workers:      workers,
		})
	},
}

var productUrlPrefixes = []string{"https://www.mercadolibre", "https://articulo.mercadolibre", "mercadolibre"}
var listUrlPrefixes = []string{"https://listado.mercadolibre", "listado.mercadolibre"}

func routeMeliUrl(url string, opts gejie.CmdOptions) {
	if url == "" {
		slog.Error("url is empty, please provide a valid meli url")
		return
	}

	topLevelTimer := utils.NewTimer("Top Level")
	defer topLevelTimer.LogElapsed()

	isProductUrl := false
	isListUrl := false
	for _, productUrl := range productUrlPrefixes {
		if strings.HasPrefix(url, productUrl) {
			isProductUrl = true
		}
	}
	for _, listUrl := range listUrlPrefixes {
		if strings.HasPrefix(url, listUrl) {
			isListUrl = true
		}
	}

	// ./gejiec meli --url https://articulo.mercadolibre.com.mx/MLM-1411526559-silla-gamer-reclinable-giratoria-ergonomica-super-comoda-_JM
	if isProductUrl {
		if opts.OnlyImages {
			slog.Info("scraping only product images", "url", url)
			images := gejie.ScrapeProductImages(nil, url)
			// just print for now
			for _, image := range images {
				slog.Info(image)
			}
			return
		}

		slog.Info("scraping product url", "url", url)
		product := gejie.ScrapeProductPageDirect(url, opts)
		utils.PrintProduct(product)

		// ./gejiec meli --url https://listado.mercadolibre.com.mx/carburador-stihl --max-items 5 --create-csv
	} else if isListUrl {
		slog.Info("scraping list url", "url", url)
		gejie.RunMeliSearch(&url, opts)

	} else {
		slog.Error("url is not a valid meli", "url", url)
	}
}

func init() {
	meliCmd.Flags().Int("max-items", 10, "max items to scrape, only for product list urls")
	meliCmd.Flags().String("url", "", "mercadolibre url to scrape - page type will be auto detected")
	meliCmd.Flags().Bool("only-images", false, "only scrape the images from given product url, no other data will be scraped")
	meliCmd.Flags().Bool("create-csv", false, "create a csv file of the scraped products")
	meliCmd.Flags().Bool("headless", false, "run the browser in headless mode")
	meliCmd.Flags().Int("workers", 2, "number of concurrent workers to scrape product pages, defaults to 2")
	rootCmd.AddCommand(meliCmd)
}
