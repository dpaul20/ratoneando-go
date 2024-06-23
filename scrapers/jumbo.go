package scrapers

import (
	"encoding/json"

	"ratoneando/cores/api"
	"ratoneando/product"
)

type ResponseProduct struct {
	ProductId   string   `json:"productId"`
	ProductName string   `json:"productName"`
	Link        string   `json:"link"`
	ProductData []string `json:"ProductData"`
	Items       []struct {
		Images []struct {
			ImageUrl string `json:"imageUrl"`
		} `json:"images"`
		Sellers []struct {
			CommertialOffer struct {
				Price                float64 `json:"Price"`
				ListPrice            float64 `json:"ListPrice"`
				PriceWithoutDiscount float64 `json:"PriceWithoutDiscount"`
				AvailableQuantity    int     `json:"AvailableQuantity"`
				IsAvailable          bool    `json:"IsAvailable"`
			} `json:"commertialOffer"`
		} `json:"sellers"`
	} `json:"items"`
}

type ProductData struct {
	MeasurementUnitUn string  `json:"MeasurementUnit"`
	UnitMultiplierUn  float64 `json:"UnitMultiplier"`
}

type RawProduct struct {
	ResponseProduct
	ProductData
}

type ResponseStructure []ResponseProduct

func JumboScraper(query string) ([]product.Schema, error) {
	return api.Core(api.CoreProps[ResponseStructure, RawProduct]{
		Query:         query,
		BaseUrl:       "https://www.jumbo.com.ar",
		SearchPattern: func(q string) string { return "/api/catalog_system/pub/products/search/?ft=" + q },
		Source:        "jumbo",
		Normalizer: func(response ResponseStructure) []RawProduct {
			var normalizedProducts []RawProduct

			for _, rawProduct := range response {
				var productData ProductData
				json.Unmarshal([]byte(rawProduct.ProductData[0]), &productData)

				normalizedProducts = append(normalizedProducts, RawProduct{
					ResponseProduct: rawProduct,
					ProductData:     productData,
				})
			}

			return normalizedProducts
		},
		Extractor: func(rawProduct RawProduct) product.ExtendedSchema {
			return product.ExtendedSchema{
				ID:          rawProduct.ProductId,
				Source:      "jumbo",
				Name:        rawProduct.ProductName,
				Link:        rawProduct.Link,
				Image:       rawProduct.Items[0].Images[0].ImageUrl,
				Unavailable: !rawProduct.Items[0].Sellers[0].CommertialOffer.IsAvailable,
				Price:       rawProduct.Items[0].Sellers[0].CommertialOffer.Price,
				ListPrice:   rawProduct.Items[0].Sellers[0].CommertialOffer.ListPrice,
				Unit:        rawProduct.MeasurementUnitUn,
				UnitFactor:  rawProduct.UnitMultiplierUn,
			}
		},
	})
}