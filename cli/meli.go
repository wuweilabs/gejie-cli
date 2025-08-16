package cli

import (
	"strings"

	"github.com/spf13/cobra"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
)

var meliCmd = &cobra.Command{
	Use:   "meli",
	Short: "scrape meli product listings, store pages, and analytics",
	Long:  "scrape meli product listings, store pages, and analytics",
	Run: func(cmd *cobra.Command, args []string) {
		maxItems, _ := cmd.Flags().GetInt("max-items")
		url, _ := cmd.Flags().GetString("url")
		routeMeliUrl(url, maxItems)
	},
}

var productUrlPrefixes = []string{"https://www.mercadolibre.com", "https://articulo.mercadolibre.com", "mercadolibre.com"}
var listUrlPrefixes = []string{"https://listado.mercadolibre.com", "listado.mercadolibre.com"}

func routeMeliUrl(url string, maxItems int) {
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
	if isProductUrl {
		gejie.ScrapeProductPage()
	}
	if isListUrl {
		gejie.RunMeliSearch(&url, int8(maxItems))
	}
}

func init() {
	meliCmd.Flags().IntP("max-items", "m", 10, "max items to scrape")
	rootCmd.AddCommand(meliCmd)
}
