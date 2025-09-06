package cli

import (
	"fmt"
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

		fmt.Printf("maxItems: %d, onlyImages: %t, url: %s, createCsv: %t, headless: %t",
			maxItems, onlyImages, url, createCsv, headlessMode)
		routeMeliUrl(url, gejie.CmdOptions{
			MaxItems:     maxItems,
			OnlyImages:   onlyImages,
			CreateCsv:    createCsv,
			HeadlessMode: headlessMode,
		})
	},
}

var productUrlPrefixes = []string{"https://www.mercadolibre", "https://articulo.mercadolibre", "mercadolibre"}
var listUrlPrefixes = []string{"https://listado.mercadolibre", "listado.mercadolibre"}

func routeMeliUrl(url string, opts gejie.CmdOptions) {
	if url == "" {
		fmt.Printf("url is empty, please provide a valid meli url")
		return
	}

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
			fmt.Printf("\nscraping only product images: %s", url)
			images := gejie.ScrapeProductImages(nil, url)
			// just print for now
			for _, image := range images {
				fmt.Println(image)
			}
			return
		}

		fmt.Printf("\nscraping product url: %s", url)
		product := gejie.ScrapeProductPageDirect(url)
		utils.PrintProduct(product)

		// ./gejiec meli --url https://listado.mercadolibre.com.mx/carburador-stihl --max-items 5 --create-csv
	} else if isListUrl {
		fmt.Printf("\nscraping list url: %s", url)
		gejie.RunMeliSearch(&url, opts)

	} else {
		fmt.Printf("url is not a valid meli, url: %s", url)
	}
}

func init() {
	meliCmd.Flags().Int("max-items", 10, "max items to scrape, only for product list urls")
	meliCmd.Flags().String("url", "", "mercadolibre url to scrape - page type will be auto detected")
	meliCmd.Flags().Bool("only-images", false, "only scrape the images from given product url, no other data will be scraped")
	meliCmd.Flags().Bool("create-csv", false, "create a csv file of the scraped products")
	meliCmd.Flags().Bool("headless", false, "run the browser in headless mode")
	rootCmd.AddCommand(meliCmd)
}
