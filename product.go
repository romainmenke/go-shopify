package goshopify

import (
	"context"
	"fmt"
	"time"
)

const productsBasePath = "admin/products"

// ProductService is an interface for interfacing with the product endpoints
// of the Shopify API.
// See: https://help.shopify.com/api/reference/product
type ProductService interface {
	List(context.Context, interface{}) ([]Product, error)
	Count(context.Context, interface{}) (int, error)
	Get(context.Context, int, interface{}) (*Product, error)
	Create(context.Context, Product) (*Product, error)
	Update(context.Context, Product) (*Product, error)
	Delete(context.Context, int) error
}

// ProductServiceOp handles communication with the product related methods of
// the Shopify API.
type ProductServiceOp struct {
	client *Client
}

// Product represents a Shopify product
type Product struct {
	ID             int             `json:"id"`
	Title          string          `json:"title"`
	BodyHTML       string          `json:"body_html"`
	Vendor         string          `json:"vendor"`
	ProductType    string          `json:"product_type"`
	Handle         string          `json:"handle"`
	CreatedAt      *time.Time      `json:"created_at"`
	UpdatedAt      *time.Time      `json:"updated_at"`
	PublishedAt    *time.Time      `json:"published_at"`
	PublishedScope string          `json:"published_scope"`
	Tags           string          `json:"tags"`
	Options        []ProductOption `json:"options"`
	Variants       []Variant       `json:"variants"`
	Image          Image           `json:"image"`
	Images         []Image         `json:"images"`
}

// The options provided by Shopify
type ProductOption struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}

// Represents the result from the products/X.json endpoint
type ProductResource struct {
	Product *Product `json:"product"`
}

// Represents the result from the products.json endpoint
type ProductsResource struct {
	Products []Product `json:"products"`
}

// List products
func (s *ProductServiceOp) List(ctx context.Context, options interface{}) ([]Product, error) {
	path := fmt.Sprintf("%s.json", productsBasePath)
	resource := new(ProductsResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Products, err
}

// Count products
func (s *ProductServiceOp) Count(ctx context.Context, options interface{}) (int, error) {
	path := fmt.Sprintf("%s/count.json", productsBasePath)
	return s.client.Count(ctx, path, options)
}

// Get individual product
func (s *ProductServiceOp) Get(ctx context.Context, productID int, options interface{}) (*Product, error) {
	path := fmt.Sprintf("%s/%d.json", productsBasePath, productID)
	resource := new(ProductResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Product, err
}

// Create a new product
func (s *ProductServiceOp) Create(ctx context.Context, product Product) (*Product, error) {
	path := fmt.Sprintf("%s.json", productsBasePath)
	wrappedData := ProductResource{Product: &product}
	resource := new(ProductResource)
	err := s.client.Post(ctx, path, wrappedData, resource)
	return resource.Product, err
}

// Update an existing product
func (s *ProductServiceOp) Update(ctx context.Context, product Product) (*Product, error) {
	path := fmt.Sprintf("%s/%d.json", productsBasePath, product.ID)
	wrappedData := ProductResource{Product: &product}
	resource := new(ProductResource)
	err := s.client.Put(ctx, path, wrappedData, resource)
	return resource.Product, err
}

// Delete an existing product
func (s *ProductServiceOp) Delete(ctx context.Context, productID int) error {
	return s.client.Delete(ctx, fmt.Sprintf("%s/%d.json", productsBasePath, productID))
}
