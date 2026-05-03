package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"yadro.com/course/closers"
	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	resp, err := c.client.Get(c.url + "/" + strconv.Itoa(id) + "/info.0.json")
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("cannot get info for comic %d: %w", id, err)
	}
	defer closers.CloseOrLog(resp.Body, c.log)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("cannot read response body: %w", err)
	}

	data := struct {
		ID          int    `json:"num"`
		URL         string `json:"img"`
		Title       string `json:"title"`
		Description string `json:"transcript"`
		Alternative string `json:"alt"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	return core.XKCDInfo{
		ID:          data.ID,
		URL:         data.URL,
		Title:       data.Title,
		Description: data.Description + data.Alternative,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	resp, err := c.client.Get(c.url + "/info.0.json")
	if err != nil {
		return 0, fmt.Errorf("cannot get last XKCD info: %w", err)
	}
	defer closers.CloseOrLog(resp.Body, c.log)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("cannot read response body: %w", err)
	}

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, fmt.Errorf("cannot unmarshal response body: %w", err)
	}

	num, ok := data["num"]
	if !ok {
		return 0, fmt.Errorf("response body does not contain num field")
	}
	id := int(num.(float64))
	return id, nil
}
