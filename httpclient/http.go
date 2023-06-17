package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/hnlq715/gobreak"
	"go.uber.org/zap"
)

type Client struct {
	ctx              context.Context
	name             string
	baseURL          string
	params           map[string]string
	additionalParams map[string][]string
	headerParams     map[string]string
	method           string
	app              string
	HTTPClient       HTTPClient
	body             io.Reader
	logger           *zap.Logger
	fallback         func(context.Context, error) error
}

func New(ctx context.Context, timeout time.Duration, method, operationName, baseURL, app string, logger *zap.Logger) *Client {
	return &Client{
		ctx:        ctx,
		method:     method,
		name:       operationName,
		baseURL:    baseURL,
		app:        app,
		HTTPClient: Get(timeout),
		logger:     logger,
	}
}

func Get(timeout time.Duration) HTTPClient {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        60,
			MaxIdleConnsPerHost: 60,
		},
	}
}

func (c *Client) WithHTTPMethod(method string) *Client {
	c.method = method
	return c
}

func (c *Client) WithHTTClient(httpClient HTTPClient) *Client {
	c.HTTPClient = httpClient
	return c
}

func (c *Client) WithParams(params map[string]string) *Client {
	c.params = params
	return c
}

// WithAdditionalParams - adds more query parameters of the same kind
func (c *Client) WithAdditionalParams(params map[string][]string) *Client {
	c.additionalParams = params
	return c
}

func (ccient *Client) WithBody(body io.Reader) *Client {
	ccient.body = body
	return ccient
}

func (c *Client) WithHeaderParams(headerParams map[string]string) *Client {
	c.headerParams = headerParams
	return c
}

func (c *Client) WithFallback(fallback func(context.Context, error) error) *Client {
	c.fallback = fallback
	return c
}

func (c *Client) Call() (*http.Response, string, error) {
	c.logger.With(zap.String("http_method", c.method), zap.String("base_url", c.baseURL))
	c.logger.Info("[START] setting up client call")

	req, err := http.NewRequest(c.method, c.baseURL, c.body)
	req.Close = true
	if err != nil {
		c.logger.Error("[ERROR] while creating the request")
		return nil, "", err
	}

	query := req.URL.Query()
	for key, value := range c.params {
		query.Add(key, value)
	}

	if c.additionalParams != nil {
		for key, additonalValueIn := range c.additionalParams {
			for _, value := range additonalValueIn {
				query.Add(key, value)
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	for key, value := range c.headerParams {
		req.Header.Add(key, value)
	}

	c.logger.With(
		zap.String("URL", req.URL.String()),
		zap.Any("headers", c.headerParams),
		zap.Any("params", c.params),
		zap.Any("body", c.body),
	)

	response, err := c.call(req)

	if err != nil {
		c.logger.Error("[ERROR] executing request", zap.Error(err))
		return nil, "", err
	}

	var content string

	if response != nil {
		reader, err := responseBodyReader(response)
		defer closeResponseBodyReader(reader, c.logger)

		if err != nil {
			c.logger.Error("[ERROR] reading response body", zap.Error(err))
			return nil, "", err
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			c.logger.Error("[ERROR] reading response body", zap.Error(err))
			return nil, "", err
		} else {
			content = string(body)
		}
	}

	return response, content, nil
}

func (c *Client) call(req *http.Request) (*http.Response, error) {
	response := make(chan *http.Response, 1)

	errs := gobreak.Do(c.ctx, c.name, func(context.Context) error {
		resp, err := c.HTTPClient.Do(req)
		if resp != nil {
			response <- resp
		}

		return err
	}, c.fallback)

	if errs == nil {
		return <-response, nil
	}

	return nil, errs
}

// responseBodyReader - resolves different content types of response bodies
func responseBodyReader(res *http.Response) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error

	switch res.Header.Get("Content-Encoding") {
	default:
		reader = res.Body
	}

	return reader, err
}

func closeResponseBodyReader(reader io.ReadCloser, log *zap.Logger) {
	err := reader.Close()
	if err != nil {
		log.Error("[ERROR] closing response body reader", zap.Error(err))
	}
}
