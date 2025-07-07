package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/eslami200117/ala_unlimited/model/extract"
)

type ProductResponse struct {
	Data struct {
		Product struct {
			ID             float64            `json:"id"`
			Variants       []*VariantResponse `json:"variants"`
			DefaultVariant *VariantResponse   `json:"default_variant"` // Changed to pointer to make it nullable
		} `json:"product"`
	} `json:"data"`
}

type VariantResponse struct {
	Color struct {
		Title string `json:"title"`
	} `json:"color"`
	Price struct {
		SellingPrice float64 `json:"selling_price"`
		IsPromotion  bool    `json:"is_promotion"`
	} `json:"price"`
	Seller struct {
		ID float64 `json:"id"`
	} `json:"seller"`
}

func (c *Core) findPrice(colors []string, resp *http.Response) (result *extract.ExtProductPrice, err error) {
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	result = &extract.ExtProductPrice{
		Status:   http.StatusOK,
		Variants: make(map[string][]*extract.Variant),
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var productResp ProductResponse
	if err := json.Unmarshal(body, &productResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle the case where DefaultVariant might be nil
	var defaultColor string
	var defaultSellerID int

	if productResp.Data.Product.DefaultVariant != nil {
		result.DKP = int(productResp.Data.Product.ID)
		result.BuyBoxPrice = int(productResp.Data.Product.DefaultVariant.Price.SellingPrice) / 10
		defaultColor = productResp.Data.Product.DefaultVariant.Color.Title
		defaultSellerID = int(productResp.Data.Product.DefaultVariant.Seller.ID)
	}

	for _, color := range colors {
		variants := c.extractVariantsForColor(color, productResp.Data.Product.Variants, defaultColor, defaultSellerID, c.sellerMap)
		if len(variants) > 0 {
			result.Variants[color] = variants
		}
	}

	sortVariantsByPrice(result.Variants)

	return result, nil
}

func (c *Core) extractVariantsForColor(color string, variants []*VariantResponse, defaultColor string, defaultSellerID int, sellerMap map[int]string) []*extract.Variant {
	var result []*extract.Variant
	c.sellerMutex.RLock()
	defer c.sellerMutex.RUnlock()
	for _, v := range variants {
		if v.Color.Title == color {
			sellerID := int(v.Seller.ID)
			sellerName := sellerMap[sellerID]
			if _, exists := sellerMap[sellerID]; !exists {
				c.logger.Warn().Str("seller id", strconv.Itoa(sellerID))
				sellerName = "False"
			}
			variant := &extract.Variant{
				Seller:         sellerName,
				SellerID:       sellerID,
				Price:          int(v.Price.SellingPrice) / 10,
				VarWinner:      color == defaultColor,
				Promotion:      v.Price.IsPromotion,
				BuyBoxSellerID: defaultSellerID,
			}

			result = append(result, variant)
		}
	}

	return result
}

func sortVariantsByPrice(variantMap map[string][]*extract.Variant) {
	for color := range variantMap {
		sort.Slice(variantMap[color], func(i, j int) bool {
			return variantMap[color][i].Price < variantMap[color][j].Price
		})
	}
}
