package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *Server) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.service.Search(context.Background(), int(req.Limit), req.Phrase)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var reply searchpb.SearchReply
	for _, comic := range comics {
		reply.Comics = append(reply.Comics, &searchpb.Comic{Id: int64(comic.ID), Url: comic.URL})
	}
	return &reply, nil
}

func (s *Server) ISearch(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.service.ISearch(context.Background(), int(req.Limit), req.Phrase)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var reply searchpb.SearchReply
	for _, comic := range comics {
		reply.Comics = append(reply.Comics, &searchpb.Comic{Id: int64(comic.ID), Url: comic.URL})
	}
	return &reply, nil
}
