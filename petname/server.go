package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/ilyakaznacheev/cleanenv"
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
	for range req.Names {
		name := petname.Generate(int(req.Words), req.Separator)
		if err := stream.Send(&petnamepb.PetnameResponse{Name: name}); err != nil {
			return status.Errorf(codes.Internal, "failed to send name: %v", err)
		}
	}

	return nil
}

type Config struct {
	Port string `yaml:"port" env:"PETNAME_GRPC_PORT" env-default:"8080"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("Error reading config file", "error", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
