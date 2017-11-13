package goshopify

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const variantsBasePath = "admin/variants"

// VariantService is an interface for interacting with the variant endpoints
// of the Shopify API.
// See https://help.shopify.com/api/reference/product_variant
type VariantService interface {
	List(context.Context, int, interface{}) ([]Variant, error)
	Count(context.Context, int, interface{}) (int, error)
	Get(context.Context, int, interface{}) (*Variant, error)
	Create(context.Context, int, Variant) (*Variant, error)
	Update(context.Context, Variant) (*Variant, error)
	Delete(context.Context, int, int) error
}

// VariantServiceOp handles communication with the variant related methods of
// the Shopify API.
type VariantServiceOp struct {
	client *Client
}

// Variant represents a Shopify variant
type Variant struct {
	ID                   int              `json:"id"`
	ProductID            int              `json:"product_id"`
	Title                string           `json:"title"`
	Sku                  string           `json:"sku"`
	Position             int              `json:"position"`
	Grams                int              `json:"grams"`
	InventoryPolicy      string           `json:"inventory_policy"`
	Price                *decimal.Decimal `json:"price"`
	CompareAtPrice       *decimal.Decimal `json:"compare_at_price"`
	FulfillmentService   string           `json:"fulfillment_service"`
	InventoryManagement  string           `json:"inventory_management"`
	Option1              string           `json:"option1"`
	Option2              string           `json:"option2"`
	Option3              string           `json:"option3"`
	CreatedAt            *time.Time       `json:"created_at"`
	UpdatedAt            *time.Time       `json:"updated_at"`
	Taxable              bool             `json:"taxable"`
	Barcode              string           `json:"barcode"`
	ImageID              int              `json:"image_id"`
	InventoryQuantity    int              `json:"inventory_quantity"`
	Weight               *decimal.Decimal `json:"weight"`
	WeightUnit           string           `json:"weight_unit"`
	OldInventoryQuantity int              `json:"old_inventory_quantity"`
	RequireShipping      bool             `json:"requires_shipping"`
}

// VariantResource represents the result from the variants/X.json endpoint
type VariantResource struct {
	Variant *Variant `json:"variant"`
}

// VariantsResource represents the result from the products/X/variants.json endpoint
type VariantsResource struct {
	Variants []Variant `json:"variants"`
}

// List variants
func (s *VariantServiceOp) List(ctx context.Context, productID int, options interface{}) ([]Variant, error) {
	path := fmt.Sprintf("%s/%d/variants.json", productsBasePath, productID)
	resource := new(VariantsResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Variants, err
}

// Count variants
func (s *VariantServiceOp) Count(ctx context.Context, productID int, options interface{}) (int, error) {
	path := fmt.Sprintf("%s/%d/variants/count.json", productsBasePath, productID)
	return s.client.Count(ctx, path, options)
}

// Get individual variant
func (s *VariantServiceOp) Get(ctx context.Context, variantID int, options interface{}) (*Variant, error) {
	path := fmt.Sprintf("%s/%d.json", variantsBasePath, variantID)
	resource := new(VariantResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Variant, err
}

// Create a new variant
func (s *VariantServiceOp) Create(ctx context.Context, productID int, variant Variant) (*Variant, error) {
	path := fmt.Sprintf("%s/%d/variants.json", productsBasePath, productID)
	wrappedData := VariantResource{Variant: &variant}
	resource := new(VariantResource)
	err := s.client.Post(ctx, path, wrappedData, resource)
	return resource.Variant, err
}

// Update existing variant
func (s *VariantServiceOp) Update(ctx context.Context, variant Variant) (*Variant, error) {
	path := fmt.Sprintf("%s/%d.json", variantsBasePath, variant.ID)
	wrappedData := VariantResource{Variant: &variant}
	resource := new(VariantResource)
	err := s.client.Put(ctx, path, wrappedData, resource)
	return resource.Variant, err
}

// Delete an existing product
func (s *VariantServiceOp) Delete(ctx context.Context, productID int, variantID int) error {
	return s.client.Delete(ctx, fmt.Sprintf("%s/%d/variants/%d.json", productsBasePath, productID, variantID))
}
