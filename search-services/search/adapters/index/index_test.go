package index_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/adapters/index"
	"yadro.com/course/search/core"
	"yadro.com/course/search/core/mocks"
)

func TestNew(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := &mocks.MockDB{}
	
	idx, err := index.New(log, mockDB, 30*time.Second)
	
	require.NoError(t, err)
	assert.NotNil(t, idx)
}

func TestIndexer_ISearchComics_EmptyIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	ctx := context.Background()
	results, err := idx.ISearchComics(ctx, 10, []string{"test"})
	
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestIndexer_ISearchComics_MatchFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	// Pre-populate index via Build
	mockDB.EXPECT().GetAllComics(gomock.Any()).Return([]core.Comics{
		{ID: 1, URL: "https://example.com/1", Words: []string{"test", "word"}},
		{ID: 2, URL: "https://example.com/2", Words: []string{"test", "another"}},
	}, nil)
	
	ctx := context.Background()
	err := idx.Build(ctx)
	require.NoError(t, err)
	
	// Search for "test"
	results, err := idx.ISearchComics(ctx, 10, []string{"test"})
	
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0].ID)
	assert.Equal(t, 2, results[1].ID)
}

func TestIndexer_ISearchComics_Limit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	mockDB.EXPECT().GetAllComics(gomock.Any()).Return([]core.Comics{
		{ID: 1, URL: "url1", Words: []string{"common"}},
		{ID: 2, URL: "url2", Words: []string{"common"}},
		{ID: 3, URL: "url3", Words: []string{"common"}},
	}, nil)
	
	ctx := context.Background()
	_ = idx.Build(ctx)
	
	results, err := idx.ISearchComics(ctx, 2, []string{"common"})
	
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestIndexer_ISearchComics_NoMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	mockDB.EXPECT().GetAllComics(gomock.Any()).Return([]core.Comics{
		{ID: 1, URL: "url1", Words: []string{"alpha"}},
	}, nil)
	
	ctx := context.Background()
	_ = idx.Build(ctx)
	
	results, err := idx.ISearchComics(ctx, 10, []string{"beta"})
	
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestIndexer_Build_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	mockDB.EXPECT().GetAllComics(gomock.Any()).Return([]core.Comics{
		{ID: 1, URL: "https://xkcd.com/1", Words: []string{"hello", "world"}},
		{ID: 2, URL: "https://xkcd.com/2", Words: []string{"hello", "golang"}},
	}, nil)
	
	ctx := context.Background()
	err := idx.Build(ctx)
	
	assert.NoError(t, err)
	
	// Verify index was populated by searching
	results, err := idx.ISearchComics(ctx, 10, []string{"hello"})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestIndexer_Build_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockDB := mocks.NewMockDB(ctrl)
	
	idx, _ := index.New(log, mockDB, 30*time.Second)
	
	mockDB.EXPECT().GetAllComics(gomock.Any()).Return(nil, assert.AnError)
	
	ctx := context.Background()
	err := idx.Build(ctx)
	
	assert.Error(t, err)
}
