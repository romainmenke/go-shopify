// Package goshopify provides methods for making requests to Shopify's admin API.
package goshopify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	UserAgent = "goshopify/1.0.0"
)

// App represents basic app settings such as Api key, secret, scope, and redirect url.
// See oauth.go for OAuth related helper functions.
type App struct {
	ApiKey      string
	ApiSecret   string
	RedirectUrl string
	Scope       string
	Password    string
}

// Client manages communication with the Shopify API.
type Client struct {
	// HTTP client used to communicate with the DO API.
	Client *http.Client

	// App settings
	app App

	// Base URL for API requests.
	// This is set on a per-store basis which means that each store must have
	// its own client.
	baseURL *url.URL

	// A permanent access token
	token string

	// Services used for communicating with the API
	Product   ProductService
	Customer  CustomerService
	Order     OrderService
	Shop      ShopService
	Webhook   WebhookService
	Variant   VariantService
	Image     ImageService
	Metafield MetafieldService
}

// A general response error that follows a similar layout to Shopify's response
// errors, i.e. either a single message or a list of messages.
type ResponseError struct {
	Status  int
	Message string
	Errors  []string
}

func (e ResponseError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	sort.Strings(e.Errors)
	s := strings.Join(e.Errors, ", ")

	if s != "" {
		return s
	}

	return "Unknown Error"
}

// An error specific to a rate-limiting response. Embeds the ResponseError to
// allow consumers to handle it the same was a normal ResponseError.
type RateLimitError struct {
	ResponseError
	RetryAfter int
}

// Creates an API request. A relative URL can be provided in urlStr, which will
// be resolved to the BaseURL of the Client. Relative URLS should always be
// specified without a preceding slash. If specified, the value pointed to by
// body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body, options interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Make the full url based on the relative path
	u := c.baseURL.ResolveReference(rel)

	// Add custom options
	if options != nil {
		optionsQuery, err := query.Values(options)
		if err != nil {
			return nil, err
		}

		for k, values := range u.Query() {
			for _, v := range values {
				optionsQuery.Add(k, v)
			}
		}
		u.RawQuery = optionsQuery.Encode()
	}

	// A bit of JSON ceremony
	var js []byte

	if body != nil {
		js, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", UserAgent)
	if c.token != "" {
		req.Header.Add("X-Shopify-Access-Token", c.token)
	} else if c.app.Password != "" {
		req.SetBasicAuth(c.app.ApiKey, c.app.Password)
	}
	return req, nil
}

// NewClient returns a new Shopify API client with an already authenticated shopname and
// token.
func NewClient(app App, shopName, token string) *Client {
	httpClient := http.DefaultClient

	baseURL, _ := url.Parse(ShopBaseUrl(shopName))

	c := &Client{Client: httpClient, app: app, baseURL: baseURL, token: token}
	c.Product = &ProductServiceOp{client: c}
	c.Customer = &CustomerServiceOp{client: c}
	c.Order = &OrderServiceOp{client: c}
	c.Shop = &ShopServiceOp{client: c}
	c.Webhook = &WebhookServiceOp{client: c}
	c.Variant = &VariantServiceOp{client: c}
	c.Image = &ImageServiceOp{client: c}
	c.Metafield = &MetafieldServiceOp{client: c}

	return c
}

// Do sends an API request and populates the given interface with the parsed
// response. It does not make much sense to call Do without a prepared
// interface instance.
func (c *Client) Do(req *http.Request, v interface{}) error {
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = CheckResponseError(resp)
	if err != nil {
		return err
	}

	if v != nil {
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&v)
		if err != nil {
			return err
		}
	}

	return nil
}

func wrapSpecificError(r *http.Response, err ResponseError) error {
	if err.Status == 429 {
		f, _ := strconv.ParseFloat(r.Header.Get("retry-after"), 64)
		return RateLimitError{
			ResponseError: err,
			RetryAfter:    int(f),
		}
	}
	return err
}

func CheckResponseError(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	// Create an anonoymous struct to parse the JSON data into.
	shopifyError := struct {
		Error  string      `json:"error"`
		Errors interface{} `json:"errors"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&shopifyError)
	if err != nil {
		return err
	}

	// Create the response error from the Shopify error.
	responseError := ResponseError{
		Status:  r.StatusCode,
		Message: shopifyError.Error,
	}

	// If the errors field is not filled out, we can return here.
	if shopifyError.Errors == nil {
		return wrapSpecificError(r, responseError)
	}

	// Shopify errors usually have the form:
	// {
	//   "errors": {
	//     "title": [
	//       "something is wrong"
	//     ]
	//   }
	// }
	// This structure is flattened to a single array:
	// [ "title: something is wrong" ]
	//
	// Unfortunately, "errors" can also be a single string so we have to deal
	// with that. Lots of reflection :-(
	kind := reflect.TypeOf(shopifyError.Errors).Kind()
	if kind == reflect.String {
		// Single string, use as message
		responseError.Message = shopifyError.Errors.(string)
	} else if kind == reflect.Slice {
		// An array, parse each entry as a string and join them on the message
		// json always serializes JSON arrays into []interface{}
		for _, elem := range shopifyError.Errors.([]interface{}) {
			responseError.Errors = append(responseError.Errors, fmt.Sprint(elem))
		}
		responseError.Message = strings.Join(responseError.Errors, ", ")
	} else if kind == reflect.Map {
		// A map, parse each error for each key in the map.
		// json always serializes into map[string]interface{} for objects
		for k, v := range shopifyError.Errors.(map[string]interface{}) {
			// Check to make sure the interface is a slice
			// json always serializes JSON arrays into []interface{}
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				for _, elem := range v.([]interface{}) {
					// If the primary message of the response error is not set, use
					// any message.
					if responseError.Message == "" {
						responseError.Message = fmt.Sprint(elem)
					}
					topicAndElem := fmt.Sprintf("%v: %v", k, elem)
					responseError.Errors = append(responseError.Errors, topicAndElem)
				}
			}
		}
	}

	return wrapSpecificError(r, responseError)
}

// General list options that can be used for most collections of entities.
type ListOptions struct {
	Page         int       `url:"page,omitempty"`
	Limit        int       `url:"limit,omitempty"`
	SinceID      int       `url:"since_id,omitempty"`
	CreatedAtMin time.Time `url:"created_at_min,omitempty"`
	CreatedAtMax time.Time `url:"created_at_max,omitempty"`
	UpdatedAtMin time.Time `url:"updated_at_min,omitempty"`
	UpdatedAtMax time.Time `url:"updated_at_max,omitempty"`
	Order        string    `url:"order,omitempty"`
}

// General count options that can be used for most collection counts.
type CountOptions struct {
	CreatedAtMin time.Time `url:"created_at_min,omitempty"`
	CreatedAtMax time.Time `url:"created_at_max,omitempty"`
	UpdatedAtMin time.Time `url:"updated_at_min,omitempty"`
	UpdatedAtMax time.Time `url:"updated_at_max,omitempty"`
}

func (c *Client) Count(ctx context.Context, path string, options interface{}) (int, error) {
	resource := struct {
		Count int `json:"count"`
	}{}
	err := c.Get(ctx, path, &resource, options)
	return resource.Count, err
}

// CreateAndDo performs a web request to Shopify with the given method (GET,
// POST, PUT, DELETE) and relative path (e.g. "/admin/orders.json").
// The data, options and resource arguments are optional and only relevant in
// certain situations.
// If the data argument is non-nil, it will be used as the body of the request
// for POST and PUT requests.
// The options argument is used for specifying request options such as search
// parameters like created_at_min
// Any data returned from Shopify will be marshalled into resource argument.
func (c *Client) CreateAndDo(ctx context.Context, method, path string, data, options, resource interface{}) error {
	req, err := c.NewRequest(method, path, data, options)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	err = c.Do(req, resource)
	if err != nil {
		return err
	}

	return nil
}

// Get performs a GET request for the given path and saves the result in the
// given resource.
func (c *Client) Get(ctx context.Context, path string, resource, options interface{}) error {
	return c.CreateAndDo(ctx, "GET", path, nil, options, resource)
}

// Post performs a POST request for the given path and saves the result in the
// given resource.
func (c *Client) Post(ctx context.Context, path string, data, resource interface{}) error {
	return c.CreateAndDo(ctx, "POST", path, data, nil, resource)
}

// Put performs a PUT request for the given path and saves the result in the
// given resource.
func (c *Client) Put(ctx context.Context, path string, data, resource interface{}) error {
	return c.CreateAndDo(ctx, "PUT", path, data, nil, resource)
}

// Delete performs a DELETE request for the given path
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.CreateAndDo(ctx, "DELETE", path, nil, nil, nil)
}
