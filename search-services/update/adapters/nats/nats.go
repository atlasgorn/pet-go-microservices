package nats

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Client struct {
	log *slog.Logger
	nc  *nats.Conn
}

func NewClient(log *slog.Logger, address string) (*Client, error) {
	nc, err := nats.Connect(address)
	return &Client{log: log, nc: nc}, err
}

func (c *Client) NotifyDBUpdate() error {
	err := c.nc.Publish("xkcd.db.updated", nil)
	if err != nil {
		c.log.Error("could not publish message", "error", err)
		return err
	}
	err = c.nc.Flush()
	if err != nil {
		c.log.Error("could not flush nats", "error", err)
		return err
	}
	return nil
}

func (c *Client) Close() {
	c.nc.Close()
}
