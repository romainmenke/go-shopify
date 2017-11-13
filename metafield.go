package goshopify

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const metafieldsBasePath = "admin/metafields"

type MetafieldService interface {
	List(context.Context, interface{}) ([]*Metafield, error)
	ListForObject(context.Context, string, interface{}) ([]*Metafield, error)
	Get(context.Context, int, interface{}) (*Metafield, error)
	Create(context.Context, *Metafield) (*Metafield, error)
	Update(context.Context, *Metafield) (*Metafield, error)
	Delete(context.Context, int) error
}

type MetafieldServiceOp struct {
	client *Client
}

// Metafield represents a Shopify metafield
type Metafield struct {
	ID            int         `json:"id,omitempty"`
	Namespace     string      `json:"namespace,omitempty"`
	Key           string      `json:"key,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ValueType     string      `json:"value_type,omitempty"`
	Description   string      `json:"description,omitempty"`
	OwnerID       int         `json:"owner_id,omitempty"`
	CreatedAt     string      `json:"created_at,omitempty"`
	UpdatedAt     string      `json:"updated_at,omitempty"`
	OwnerResource string      `json:"owner_resource,omitempty"`
}

type MetafieldOption struct {
	Limit        int       `url:"limit"`
	SinceID      int       `url:"since_id"`
	CreatedAtMin time.Time `url:"created_at_min"`
	CreatedAtMax time.Time `url:"created_at_max"`
	UpdatedAtMin time.Time `url:"updated_at_min"`
	UpdatedAtMax time.Time `url:"updated_at_max"`
	Namespace    string    `url:"namespace"`
	Key          string    `url:"key"`
	ValueType    string    `url:"value_type"`
	Fields       string    `url:"fields"`
}

type MetafieldResource struct {
	Metafield *Metafield `json:"metafield"`
}

type MetafieldsResource struct {
	Metafields []*Metafield `json:"metafields"`
}

func (s *MetafieldServiceOp) List(ctx context.Context, options interface{}) ([]*Metafield, error) {
	path := fmt.Sprintf("%s.json", metafieldsBasePath)
	resource := new(MetafieldsResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Metafields, err
}

func (s *MetafieldServiceOp) ListForObject(ctx context.Context, resourcePath string, options interface{}) ([]*Metafield, error) {
	resourcePath = strings.TrimSuffix(resourcePath, "/")
	resourcePath = strings.TrimPrefix(resourcePath, "/")
	path := fmt.Sprintf("/admin/%s/metafields.json", resourcePath)
	resource := new(MetafieldsResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Metafields, err
}

func (s *MetafieldServiceOp) Get(ctx context.Context, metafieldID int, options interface{}) (*Metafield, error) {
	path := fmt.Sprintf("%s/%d.json", metafieldsBasePath, metafieldID)
	resource := new(MetafieldResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Metafield, err
}

func (s *MetafieldServiceOp) Create(ctx context.Context, metafield *Metafield) (*Metafield, error) {
	path := fmt.Sprintf("%s.json", metafieldsBasePath)
	wrappedData := MetafieldResource{Metafield: metafield}
	resource := new(MetafieldResource)
	err := s.client.Post(ctx, path, wrappedData, resource)
	return resource.Metafield, err
}

func (s *MetafieldServiceOp) Update(ctx context.Context, metafield *Metafield) (*Metafield, error) {
	path := fmt.Sprintf("%s/%d.json", metafieldsBasePath, metafield.ID)
	wrappedData := MetafieldResource{Metafield: metafield}
	resource := new(MetafieldResource)
	err := s.client.Put(ctx, path, wrappedData, resource)
	return resource.Metafield, err
}

func (s *MetafieldServiceOp) Delete(ctx context.Context, metafieldID int) error {
	return s.client.Delete(ctx, fmt.Sprintf("%s/%d.json", metafieldsBasePath, metafieldID))
}
