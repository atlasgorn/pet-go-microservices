package core_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/core"
	"yadro.com/course/search/core/mocks"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db := mocks.NewMockDB(ctrl)
	words := mocks.NewMockWords(ctrl)
	indexer := mocks.NewMockIndexer(ctrl)
	
	svc, err := core.NewService(log, db, words, indexer)
	
	require.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_Search_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	mockWords := mocks.NewMockWords(ctrl)
	mockIndexer := mocks.NewMockIndexer(ctrl)
	
	svc, _ := core.NewService(log, mockDB, mockWords, mockIndexer)
	
	ctx := context.Background()
	phrase := "test query"
	normWords := []string{"test", "queri"}
	expectedComics := []core.PbComic{
		{ID: 1, URL: "https://example.com/1"},
	}
	
	mockWords.EXPECT().Norm(ctx, phrase).Return(normWords, nil)
	mockDB.EXPECT().SearchComics(ctx, 10, normWords).Return(expectedComics, nil)
	
	result, err := svc.Search(ctx, 10, phrase)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedComics, result)
}

func TestService_Search_WordsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	mockWords := mocks.NewMockWords(ctrl)
	mockIndexer := mocks.NewMockIndexer(ctrl)
	
	svc, _ := core.NewService(log, mockDB, mockWords, mockIndexer)
	
	ctx := context.Background()
	phrase := "test"
	
	mockWords.EXPECT().Norm(ctx, phrase).Return(nil, errors.New("normalization failed"))
	
	result, err := svc.Search(ctx, 10, phrase)
	
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestService_Search_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	mockWords := mocks.NewMockWords(ctrl)
	mockIndexer := mocks.NewMockIndexer(ctrl)
	
	svc, _ := core.NewService(log, mockDB, mockWords, mockIndexer)
	
	ctx := context.Background()
	phrase := "test"
	normWords := []string{"test"}
	
	mockWords.EXPECT().Norm(ctx, phrase).Return(normWords, nil)
	mockDB.EXPECT().SearchComics(ctx, 10, normWords).Return(nil, errors.New("db error"))
	
	result, err := svc.Search(ctx, 10, phrase)
	
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestService_ISearch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	mockWords := mocks.NewMockWords(ctrl)
	mockIndexer := mocks.NewMockIndexer(ctrl)
	
	svc, _ := core.NewService(log, mockDB, mockWords, mockIndexer)
	
	ctx := context.Background()
	phrase := "indexed query"
	normWords := []string{"index", "queri"}
	expectedComics := []core.PbComic{
		{ID: 42, URL: "https://xkcd.com/42"},
	}
	
	mockWords.EXPECT().Norm(ctx, phrase).Return(normWords, nil)
	mockIndexer.EXPECT().ISearchComics(ctx, 5, normWords).Return(expectedComics, nil)
	
	result, err := svc.ISearch(ctx, 5, phrase)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedComics, result)
}

func TestService_ISearch_IndexerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	mockWords := mocks.NewMockWords(ctrl)
	mockIndexer := mocks.NewMockIndexer(ctrl)
	
	svc, _ := core.NewService(log, mockDB, mockWords, mockIndexer)
	
	ctx := context.Background()
	phrase := "fail"
	normWords := []string{"fail"}
	
	mockWords.EXPECT().Norm(ctx, phrase).Return(normWords, nil)
	mockIndexer.EXPECT().ISearchComics(ctx, 10, normWords).Return(nil, errors.New("index error"))
	
	result, err := svc.ISearch(ctx, 10, phrase)
	
	assert.Error(t, err)
	assert.Nil(t, result)
}
