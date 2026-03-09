package main

import (
	"context"
	"flag"
	"log"
	"net"

	petname "github.com/dustinkirkland/golang-petname"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	petnamepb "yadro.com/course/proto"
)

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Generate(_ context.Context, req *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	if req.Words <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "words count must be positive, got: %d", req.Words)
	}
	name := petname.Generate(int(req.Words), req.Separator)

	return &petnamepb.PetnameResponse{Name: name}, nil
}

func (s *server) GenerateMany(req *petnamepb.PetnameStreamRequest, stream petnamepb.PetnameGenerator_GenerateManyServer) error {
	if req.Words <= 0 {
		return status.Errorf(codes.InvalidArgument, "words count must be positive, got: %d", req.Words)
	}
	if req.Names <= 0 {
		return status.Errorf(codes.InvalidArgument, "name count must be positive, got: %d", req.Names)
	}
	for i := int64(0); i < req.Names; i++ {
		name := petname.Generate(int(req.Words), req.Separator)
		if err := stream.Send(&petnamepb.PetnameResponse{Name: name}); err != nil {
			return status.Errorf(codes.Internal, "failed to send name: %v", err)
		}
	}

	return nil
}

func main() {
	var address string
	flag.StringVar(&address, "address", ":8080", "server address")
	flag.Parse()

	log.Printf("starting server at %s", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
