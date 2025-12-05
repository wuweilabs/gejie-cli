package meli

import (
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/zshanhui/gejiezhipin/utils"
)

type Price struct {
	AmountCentsMin int
	AmountCentsMax int
	CurrencyCode   utils.CurrencyCode
}

type GejieProductListing struct {
	Title              string
	Price              Price
	Url                string
	ReviewCount        *uint32
	Rating             *float32
	ImageUrls          []string
	SoldMoreThan       *uint32
	EstimatedSoldCount *uint32
	DescriptionContent string
	StoreInfo          StoreInfo
}

type StoreInfo struct {
	Name                 string
	Url                  string
	LogoImageSrc         string
	LogoImageSrcOriginal string
}

func ScrapeProductPageDummy(browser playwright.Browser, url string) *GejieProductListing {
	time.Sleep(time.Duration(3) * time.Second)

	reviewCount := uint32(123)
	reviewScore := float32(4.5)
	soldMoreThan := uint32(50)
	estimatedSold := uint32(75)
	return &GejieProductListing{
		Title: "Sample Product Title",
		Price: Price{
			AmountCentsMin: 1999,
			AmountCentsMax: 1999,
			CurrencyCode:   utils.CurrencyCode("USD")},
		Url:                url,
		ReviewCount:        &reviewCount,
		Rating:             &reviewScore,
		ImageUrls:          []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
		SoldMoreThan:       &soldMoreThan,
		EstimatedSoldCount: &estimatedSold,
		DescriptionContent: "This is a sample product description for demonstration purposes.",
		StoreInfo: StoreInfo{
			Name:                 "Sample Store",
			Url:                  "https://example.com/store",
			LogoImageSrc:         "https://example.com/logo.jpg",
			LogoImageSrcOriginal: "https://example.com/logo-original.jpg",
		},
	}
}
