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
	updategrpc "yadro.com/course/update/adapters/grpc"
	"yadro.com/course/update/core"
	"yadro.com/course/update/core/mocks"
	pb "yadro.com/course/proto/update"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	assert.NotNil(t, s)
}

func TestServer_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	resp, err := s.Ping(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Status_Idle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Status(ctx).Return(core.StatusIdle)
	
	resp, err := s.Status(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, pb.Status_STATUS_IDLE, resp.Status)
}

func TestServer_Status_Running(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Status(ctx).Return(core.StatusRunning)
	
	resp, err := s.Status(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, pb.Status_STATUS_RUNNING, resp.Status)
}

func TestServer_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Update(ctx).Return(nil)
	
	resp, err := s.Update(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Update_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Update(ctx).Return(core.ErrAlreadyExists)
	
	resp, err := s.Update(ctx, &emptypb.Empty{})
	assert.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestServer_Update_OtherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Update(ctx).Return(errors.New("update failed"))
	
	resp, err := s.Update(ctx, &emptypb.Empty{})
	
	assert.Nil(t, resp)
	require.Error(t, err)
	// Plain errors get wrapped by grpc, check the actual code
	st, ok := status.FromError(err)
	assert.True(t, ok)
	// Non-status errors become Unknown by default in grpc-go
	assert.Contains(t, []codes.Code{codes.Unknown, codes.Internal}, st.Code())
}

func TestServer_Stats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	serviceStats := core.ServiceStats{
		DBStats: core.DBStats{WordsTotal: 100, WordsUnique: 50, ComicsFetched: 25},
		ComicsTotal: 200,
	}
	mockService.EXPECT().Stats(ctx).Return(serviceStats, nil)
	
	resp, err := s.Stats(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(100), resp.WordsTotal)
	assert.Equal(t, int64(50), resp.WordsUnique)
	assert.Equal(t, int64(25), resp.ComicsFetched)
	assert.Equal(t, int64(200), resp.ComicsTotal)
}

func TestServer_Stats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Stats(ctx).Return(core.ServiceStats{}, errors.New("stats error"))
	
	resp, err := s.Stats(ctx, &emptypb.Empty{})
	assert.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestServer_Drop_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Drop(ctx).Return(nil)
	
	resp, err := s.Drop(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Drop_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockUpdater(ctrl)
	s := updategrpc.NewServer(mockService)
	
	ctx := context.Background()
	mockService.EXPECT().Drop(ctx).Return(errors.New("drop failed"))
	
	resp, err := s.Drop(ctx, &emptypb.Empty{})
	assert.Nil(t, resp)
	require.Error(t, err)
}
