package client

import (
	"context"
	"github.com/jomei/notionapi"
	"golang.org/x/time/rate"
	"time"
)

type Client struct {
	cli     *notionapi.Client
	limiter *rate.Limiter
}

func (c *Client) C(ctx context.Context) (*notionapi.Client, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return c.cli, nil
}

func New(key string) *Client {
	client := notionapi.NewClient(notionapi.Token(key))

	return &Client{
		cli:     client,
		limiter: rate.NewLimiter(rate.Every(time.Second), 3),
	}
}
