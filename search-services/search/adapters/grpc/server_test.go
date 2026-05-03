package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "yadro.com/course/proto/search"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/core"
	"yadro.com/course/search/core/mocks"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	assert.NotNil(t, s)
}

func TestServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	ctx := context.Background()
	resp, err := s.Ping(ctx, &emptypb.Empty{})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Search_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	ctx := context.Background()
	req := &pb.SearchRequest{Limit: 10, Phrase: "test query"}
	expectedComics := []core.PbComic{
		{ID: 1, URL: "https://example.com/1"},
		{ID: 2, URL: "https://example.com/2"},
	}

	mockService.EXPECT().Search(ctx, 10, "test query").Return(expectedComics, nil)

	resp, err := s.Search(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Comics, 2)
	assert.Equal(t, int64(1), resp.Comics[0].Id)
	assert.Equal(t, "https://example.com/1", resp.Comics[0].Url)
}

func TestServer_Search_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	ctx := context.Background()
	req := &pb.SearchRequest{Limit: 5, Phrase: "fail"}

	mockService.EXPECT().Search(ctx, 5, "fail").Return(nil, errors.New("search failed"))

	resp, err := s.Search(ctx, req)

	assert.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestServer_ISearch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	ctx := context.Background()
	req := &pb.SearchRequest{Limit: 3, Phrase: "indexed"}
	expectedComics := []core.PbComic{
		{ID: 42, URL: "https://xkcd.com/42"},
	}

	mockService.EXPECT().ISearch(ctx, 3, "indexed").Return(expectedComics, nil)

	resp, err := s.ISearch(ctx, req)

	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Comics, 1)
	assert.Equal(t, int64(42), resp.Comics[0].Id)
}

func TestServer_ISearch_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSearcher(ctrl)
	s := searchgrpc.NewServer(mockService)

	ctx := context.Background()
	req := &pb.SearchRequest{Limit: 10, Phrase: "error"}

	mockService.EXPECT().ISearch(ctx, 10, "error").Return(nil, errors.New("index error"))

	resp, err := s.ISearch(ctx, req)

	assert.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
