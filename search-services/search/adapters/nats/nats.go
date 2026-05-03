package nats

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

type Server struct {
	log     *slog.Logger
	nc      *nats.Conn
	indexer core.Indexer
}

func NewServer(log *slog.Logger, address string, idx core.Indexer) (*Server, error) {
	nc, err := nats.Connect(address)
	return &Server{log: log, nc: nc, indexer: idx}, err
}

func (c *Server) Run(ctx context.Context) error {
	_, err := c.nc.Subscribe("xkcd.db.updated", func(msg *nats.Msg) {
		if err := c.indexer.Build(ctx); err != nil {
			c.log.ErrorContext(ctx, "error building index", "error", err)
		} else {
			c.log.InfoContext(ctx, "index build successfully")
		}
	})

	return err
}

func (c *Server) Close() {
	c.nc.Close()
}
