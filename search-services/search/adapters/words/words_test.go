package words_test

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "yadro.com/course/proto/words"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/core"
)

type mockWordsServer struct {
	pb.UnimplementedWordsServer
	normFunc func(context.Context, *pb.WordsRequest) (*pb.WordsReply, error)
}

func (m *mockWordsServer) Norm(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
	if m.normFunc != nil {
		return m.normFunc(ctx, req)
	}
	return &pb.WordsReply{Words: []string{"test"}}, nil
}

func (m *mockWordsServer) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func startTestServer(t *testing.T, normFunc func(context.Context, *pb.WordsRequest) (*pb.WordsReply, error)) string {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	pb.RegisterWordsServer(s, &mockWordsServer{normFunc: normFunc})

	go func() { _ = s.Serve(lis) }()
	t.Cleanup(func() { s.Stop(); _ = lis.Close() })

	return lis.Addr().String()
}

func TestNewClient(t *testing.T) {
	addr := startTestServer(t, nil)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)

	require.NoError(t, err)
	assert.NotNil(t, client)
	if client != nil {
		_ = client.Close()
	}
}

func TestClient_Norm_Success(t *testing.T) {
	addr := startTestServer(t, func(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
		return &pb.WordsReply{Words: []string{"hello", "world"}}, nil
	})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	ctx := context.Background()

	result, err := client.Norm(ctx, "hello world")

	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, result)
}

func TestClient_Norm_ResourceExhausted(t *testing.T) {
	addr := startTestServer(t, func(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
		return nil, status.Error(codes.ResourceExhausted, "too large")
	})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	ctx := context.Background()

	result, err := client.Norm(ctx, "too large phrase")

	assert.ErrorIs(t, err, core.ErrPhraseTooLarge)
	assert.Nil(t, result)
}

func TestClient_Norm_OtherError(t *testing.T) {
	addr := startTestServer(t, func(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
		return nil, errors.New("internal error")
	})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	ctx := context.Background()

	result, err := client.Norm(ctx, "test")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestClient_Ping(t *testing.T) {
	addr := startTestServer(t, nil)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	ctx := context.Background()

	err = client.Ping(ctx)

	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	addr := startTestServer(t, nil)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)

	err = client.Close()

	assert.NoError(t, err)
}
