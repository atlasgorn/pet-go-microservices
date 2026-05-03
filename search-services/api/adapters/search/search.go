package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
	conn   *grpc.ClientConn
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
		conn:   conn,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Search(ctx context.Context, limit int, phrase string) ([]core.PbComic, error) {
	resp, err := c.client.Search(ctx, &searchpb.SearchRequest{Limit: int64(limit), Phrase: phrase})
	if err != nil {
		return nil, err
	}
	results := make([]core.PbComic, 0, len(resp.Comics))
	for _, comic := range resp.Comics {
		results = append(results, core.PbComic{
			ID:  int(comic.Id),
			URL: comic.Url,
		})
	}

	return results, nil
}

func (c Client) ISearch(ctx context.Context, limit int, phrase string) ([]core.PbComic, error) {
	resp, err := c.client.ISearch(ctx, &searchpb.SearchRequest{Limit: int64(limit), Phrase: phrase})
	if err != nil {
		return nil, err
	}
	results := make([]core.PbComic, 0, len(resp.Comics))
	for _, comic := range resp.Comics {
		results = append(results, core.PbComic{
			ID:  int(comic.Id),
			URL: comic.Url,
		})
	}

	return results, nil
}

func (c Client) Close() error {
	return c.conn.Close()
}
