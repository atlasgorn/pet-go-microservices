package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
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

type Config struct {
	Port string `yaml:"port" env:"PORT" env-default:"8080"`
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
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
