package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ratoneando/config"
	"ratoneando/product"
	"ratoneando/scrapers"
)

func NormalizedScraper(c *gin.Context) {
	referer := c.Request.Referer()
	query := c.Query("q")

	// Get the client IP
	clientIp := c.Request.Header.Get("X-Envoy-External-Address")
	if clientIp == "" {
		clientIp = c.ClientIP()
	}

	// Check if the request is coming from a valid source
	if config.ENV == "production" && (referer == "" || !strings.Contains(referer, config.WEB_URL)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden."})
		return
	}

	// Quick check to see if the query is valid
	if strings.ToLower(query) != query {
		fmt.Println("Uppercase query", query, c.Request.Header, clientIp)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
		return
	}

	// Middlewares
	// TODO

	// Run the scrapers
	var wg sync.WaitGroup
	var mu sync.Mutex

	scrappers := []func(string) ([]product.Schema, error){
		scrapers.CotoScraper,
		scrapers.JumboScraper,
	}

	type result struct {
		Products []product.Schema
		Error    error
	}

	results := make([]result, len(scrappers))

	for i, scrapper := range scrappers {
		wg.Add(1)
		go func(i int, scrapper func(string) ([]product.Schema, error)) {
			defer wg.Done()
			products, err := scrapper(query)
			mu.Lock()
			results[i] = result{Products: products, Error: err}
			mu.Unlock()
		}(i, scrapper)
	}

	wg.Wait()

	var failedScrappers []string
	var normalizedProducts []product.Schema

	for _, result := range results {
		if result.Error != nil {
			failedScrappers = append(failedScrappers, result.Error.Error())
		} else {
			normalizedProducts = append(normalizedProducts, result.Products...)
		}
	}

	// Return the products
	c.JSON(http.StatusOK, gin.H{
		"products":       normalizedProducts,
		"failedScrapers": failedScrappers,
		"timestamp":      time.Now(),
	})
}