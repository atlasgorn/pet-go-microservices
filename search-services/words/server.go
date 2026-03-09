package main

import (
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

const maxMessageSize = 4 << 10 // 4 KiB

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	size := len(in.Phrase)
	if size > maxMessageSize {
		return nil, status.Errorf(codes.ResourceExhausted,
			"message size %d bytes exceeds maximum allowed size of %d bytes",
			size, maxMessageSize)
	}
	result := words.Normalize(in.Phrase)

	return &wordspb.WordsReply{
		Words: result,
	}, nil
}

func main() {
	var address string
	flag.StringVar(&address, "address", ":8080", "server address")
	flag.Parse()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
