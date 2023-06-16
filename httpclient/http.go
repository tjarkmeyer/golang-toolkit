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

func New(ctx context.Context, timeout time.Duration, method string, operationName string, baseURL string, app string, logger *zap.Logger) *Client {
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
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        60,
			MaxIdleConnsPerHost: 60,
		},
	}

	return client
}

func (client *Client) WithHTTPMethod(method string) *Client {
	client.method = method
	return client
}

func (client *Client) WithHTTClient(httpClient HTTPClient) *Client {
	client.HTTPClient = httpClient
	return client
}

func (client *Client) WithParams(params map[string]string) *Client {
	client.params = params
	return client
}

// WithAdditionalParams adds more query parameters of the same kind
func (client *Client) WithAdditionalParams(params map[string][]string) *Client {
	client.additionalParams = params
	return client
}

func (client *Client) WithBody(body io.Reader) *Client {
	client.body = body
	return client
}

func (client *Client) WithHeaderParams(headerParams map[string]string) *Client {
	client.headerParams = headerParams
	return client
}

func (client *Client) WithFallback(fallback func(context.Context, error) error) *Client {
	client.fallback = fallback
	return client
}

func (client *Client) Call() (*http.Response, string, error) {
	client.logger = client.logger.With(zap.String("http_method", client.method), zap.String("base_url", client.baseURL))

	client.logger.Info("[START] setting up client call")

	req, err := http.NewRequest(client.method, client.baseURL, client.body)
	req.Close = true
	if err != nil {
		client.logger.Error("[ERROR] creation the request")
		return nil, "", err
	}

	query := req.URL.Query()
	for key, value := range client.params {
		query.Add(key, value)
	}

	if client.additionalParams != nil {
		for key, additonalValueIn := range client.additionalParams {
			for _, value := range additonalValueIn {
				query.Add(key, value)
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	for key, value := range client.headerParams {
		req.Header.Add(key, value)
	}

	client.logger = client.logger.With(
		zap.String("URL", req.URL.String()),
		zap.Any("headers", client.headerParams),
		zap.Any("params", client.params),
		zap.Any("body", client.body),
	)

	response, err := client.call(req, client.fallback)

	if err != nil {
		client.logger.Error("[ERROR] executing request", zap.Error(err))
		return nil, "", err
	}

	var content string

	if response != nil {
		reader, err := responseBodyReader(response)
		defer closeResponseBodyReader(reader, client.logger)

		if err != nil {
			client.logger.Error("[ERROR] reading response body", zap.Error(err))
			return nil, "", err
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			client.logger.Error("[ERROR] reading response body", zap.Error(err))
			return nil, "", err
		} else {
			content = string(body)
		}
	}

	return response, content, nil
}

func (client *Client) call(req *http.Request, fallback func(context.Context, error) error) (*http.Response, error) {
	response := make(chan *http.Response, 1)

	errs := gobreak.Do(client.ctx, client.name, func(context.Context) error {
		resp, err := client.HTTPClient.Do(req)
		if resp != nil {
			response <- resp
		}

		return err
	}, fallback)

	if errs == nil {
		return <-response, nil
	}

	return nil, errs
}

// responseBodyReader resolves different content types of response bodies
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
