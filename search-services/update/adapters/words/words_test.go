package words_test

import (
	"context"
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
	"yadro.com/course/update/adapters/words"
	"yadro.com/course/update/core"
)

type mockWordsServer struct {
	pb.UnimplementedWordsServer
}

func (m *mockWordsServer) Norm(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
	return &pb.WordsReply{Words: []string{"test"}}, nil
}

func (m *mockWordsServer) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type mockWordsServerResourceExhausted struct{ pb.UnimplementedWordsServer }

func (m *mockWordsServerResourceExhausted) Norm(ctx context.Context, req *pb.WordsRequest) (*pb.WordsReply, error) {
	return nil, status.Error(codes.ResourceExhausted, "too large")
}

func startTestServer(t *testing.T, srv pb.WordsServer) string {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer()
	pb.RegisterWordsServer(s, srv)

	go func() { _ = s.Serve(lis) }()
	t.Cleanup(func() { s.Stop(); _ = lis.Close() })

	return lis.Addr().String()
}

func TestNewClient(t *testing.T) {
	addr := startTestServer(t, &mockWordsServer{})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	assert.NotNil(t, client)
	if client != nil {
		_ = client.Close()
	}
}

func TestClient_Norm_Success(t *testing.T) {
	addr := startTestServer(t, &mockWordsServer{})
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

	assert.NoError(t, err)
	assert.Equal(t, []string{"test"}, result)
}

func TestClient_Norm_ResourceExhausted(t *testing.T) {
	addr := startTestServer(t, &mockWordsServerResourceExhausted{})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	_, err = client.Norm(context.Background(), "large")
	assert.ErrorIs(t, err, core.ErrPhraseTooLarge)
}

func TestClient_Ping(t *testing.T) {
	addr := startTestServer(t, &mockWordsServer{})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	err = client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	addr := startTestServer(t, &mockWordsServer{})
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	client, err := words.NewClient(addr, log)
	require.NoError(t, err)
	err = client.Close()
	assert.NoError(t, err)
}
