package goshopify

import (
	"context"
	"fmt"
	"time"
)

const webhooksBasePath = "admin/webhooks"

// WebhookService is an interface for interfacing with the webhook endpoints of
// the Shopify API.
// See: https://help.shopify.com/api/reference/webhook
type WebhookService interface {
	List(context.Context, interface{}) ([]Webhook, error)
	Count(context.Context, interface{}) (int, error)
	Get(context.Context, int, interface{}) (*Webhook, error)
	Create(context.Context, Webhook) (*Webhook, error)
	Update(context.Context, Webhook) (*Webhook, error)
	Delete(context.Context, int) error
}

// WebhookServiceOp handles communication with the webhook-related methods of
// the Shopify API.
type WebhookServiceOp struct {
	client *Client
}

// Webhook represents a Shopify webhook
type Webhook struct {
	ID                  int        `json:"id"`
	Address             string     `json:"address"`
	Topic               string     `json:"topic"`
	Format              string     `json:"format"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
	Fields              []string   `json:"fields"`
	MetafieldNamespaces []string   `json:"metafield_namespaces"`
}

// WebhookOptions can be used for filtering webhooks on a List request.
type WebhookOptions struct {
	Address string `url:"address,omitempty"`
	Topic   string `url:"topic,omitempty"`
}

// WebhookResource represents the result from the admin/webhooks.json endpoint
type WebhookResource struct {
	Webhook *Webhook `json:"webhook"`
}

// WebhooksResource is the root object for a webhook get request.
type WebhooksResource struct {
	Webhooks []Webhook `json:"webhooks"`
}

// List webhooks
func (s *WebhookServiceOp) List(ctx context.Context, options interface{}) ([]Webhook, error) {
	path := fmt.Sprintf("%s.json", webhooksBasePath)
	resource := new(WebhooksResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Webhooks, err
}

// Count webhooks
func (s *WebhookServiceOp) Count(ctx context.Context, options interface{}) (int, error) {
	path := fmt.Sprintf("%s/count.json", webhooksBasePath)
	return s.client.Count(ctx, path, options)
}

// Get individual webhook
func (s *WebhookServiceOp) Get(ctx context.Context, webhookdID int, options interface{}) (*Webhook, error) {
	path := fmt.Sprintf("%s/%d.json", webhooksBasePath, webhookdID)
	resource := new(WebhookResource)
	err := s.client.Get(ctx, path, resource, options)
	return resource.Webhook, err
}

// Create a new webhook
func (s *WebhookServiceOp) Create(ctx context.Context, webhook Webhook) (*Webhook, error) {
	path := fmt.Sprintf("%s.json", webhooksBasePath)
	wrappedData := WebhookResource{Webhook: &webhook}
	resource := new(WebhookResource)
	err := s.client.Post(ctx, path, wrappedData, resource)
	return resource.Webhook, err
}

// Update an existing webhook.
func (s *WebhookServiceOp) Update(ctx context.Context, webhook Webhook) (*Webhook, error) {
	path := fmt.Sprintf("%s/%d.json", webhooksBasePath, webhook.ID)
	wrappedData := WebhookResource{Webhook: &webhook}
	resource := new(WebhookResource)
	err := s.client.Put(ctx, path, wrappedData, resource)
	return resource.Webhook, err
}

// Delete an existing webhooks
func (s *WebhookServiceOp) Delete(ctx context.Context, ID int) error {
	return s.client.Delete(ctx, fmt.Sprintf("%s/%d.json", webhooksBasePath, ID))
}
