package core_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/update/core"
	"yadro.com/course/update/core/mocks"
)

type testFixture struct {
	svc    *core.Service
	ctrl   *gomock.Controller
	db     *mocks.MockDB
	xkcd   *mocks.MockXKCD
	words  *mocks.MockWords
	broker *mocks.MockBroker
}

func setupTest(t *testing.T, concurrency int) *testFixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db := mocks.NewMockDB(ctrl)
	xkcd := mocks.NewMockXKCD(ctrl)
	words := mocks.NewMockWords(ctrl)
	broker := mocks.NewMockBroker(ctrl)

	svc, err := core.NewService(log, db, xkcd, words, broker, concurrency)
	require.NoError(t, err)

	return &testFixture{
		svc:    svc,
		ctrl:   ctrl,
		db:     db,
		xkcd:   xkcd,
		words:  words,
		broker: broker,
	}
}

func TestNewService_InvalidConcurrency(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	svc, err := core.NewService(log, nil, nil, nil, nil, 0)

	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewService_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db := mocks.NewMockDB(ctrl)
	xkcd := mocks.NewMockXKCD(ctrl)
	words := mocks.NewMockWords(ctrl)
	broker := mocks.NewMockBroker(ctrl)

	svc, err := core.NewService(log, db, xkcd, words, broker, 5)

	require.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_Status(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	status := f.svc.Status(ctx)

	assert.Equal(t, core.StatusIdle, status)
}

func TestService_Drop_Success(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.db.EXPECT().Drop(ctx).Return(nil)
	f.broker.EXPECT().NotifyDBUpdate().Return(nil)

	err := f.svc.Drop(ctx)

	assert.NoError(t, err)
}

func TestService_Drop_DBError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.db.EXPECT().Drop(ctx).Return(errors.New("drop failed"))

	err := f.svc.Drop(ctx)

	assert.Error(t, err)
}

func TestService_Drop_BrokerNotifyError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.db.EXPECT().Drop(ctx).Return(nil)
	f.broker.EXPECT().NotifyDBUpdate().Return(errors.New("notify failed"))

	// Broker error is logged but not returned
	err := f.svc.Drop(ctx)
	assert.NoError(t, err)
}

func TestService_Stats_Success(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()
	dbStats := core.DBStats{
		WordsTotal:    100,
		WordsUnique:   50,
		ComicsFetched: 25,
	}

	f.db.EXPECT().Stats(ctx).Return(dbStats, nil)
	f.xkcd.EXPECT().LastID(ctx).Return(30, nil)

	stats, err := f.svc.Stats(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, stats.WordsTotal)
	assert.Equal(t, 50, stats.WordsUnique)
	assert.Equal(t, 25, stats.ComicsFetched)
	assert.Equal(t, 29, stats.ComicsTotal)
}

func TestService_Stats_DBError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.db.EXPECT().Stats(ctx).Return(core.DBStats{}, errors.New("stats error"))

	stats, err := f.svc.Stats(ctx)

	assert.Error(t, err)
	assert.Equal(t, core.ServiceStats{}, stats)
}

func TestService_Stats_XKCDError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()
	dbStats := core.DBStats{WordsTotal: 10}

	f.db.EXPECT().Stats(ctx).Return(dbStats, nil)
	f.xkcd.EXPECT().LastID(ctx).Return(0, errors.New("xkcd error"))

	stats, err := f.svc.Stats(ctx)

	assert.Error(t, err)
	assert.Equal(t, core.ServiceStats{}, stats)
}

func TestService_Update_NoNewComics(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(10, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)

	err := f.svc.Update(ctx)

	assert.NoError(t, err)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_Success(t *testing.T) {
	f := setupTest(t, 2) // concurrency = 2
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(5, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1, 2, 3}, nil)

	info4 := core.XKCDInfo{ID: 4, URL: "url4", Title: "T4", Description: "D4"}
	info5 := core.XKCDInfo{ID: 5, URL: "url5", Title: "T5", Description: "D5"}

	f.xkcd.EXPECT().Get(gomock.Any(), 4).Return(info4, nil)
	f.xkcd.EXPECT().Get(gomock.Any(), 5).Return(info5, nil)

	f.words.EXPECT().Norm(gomock.Any(), "T4 D4").Return([]string{"a", "b"}, nil)
	f.words.EXPECT().Norm(gomock.Any(), "T5 D5").Return([]string{"c"}, nil)

	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 4, URL: "url4", Words: []string{"a", "b"}}).Return(nil)
	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 5, URL: "url5", Words: []string{"c"}}).Return(nil)

	f.broker.EXPECT().NotifyDBUpdate().Return(nil)

	err := f.svc.Update(ctx)

	assert.NoError(t, err)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_AlreadyRunning(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()
	blockCh := make(chan struct{})

	f.xkcd.EXPECT().LastID(ctx).DoAndReturn(func(ctx context.Context) (int, error) {
		<-blockCh
		return 0, errors.New("unblocked")
	})

	var firstErr error
	done := make(chan struct{})
	go func() {
		firstErr = f.svc.Update(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// Second update should immediately fail with ErrAlreadyExists
	err := f.svc.Update(ctx)
	assert.Equal(t, core.ErrAlreadyExists, err)

	close(blockCh)
	<-done
	assert.Error(t, firstErr)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_LastIDError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(0, errors.New("network error"))

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_DBIDsError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(10, nil)
	f.db.EXPECT().IDs(ctx).Return(nil, errors.New("db error"))

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_PartialGetErrors(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(3, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil) // fetch 2,3

	info3 := core.XKCDInfo{ID: 3, URL: "url3", Title: "T3", Description: "D3"}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{}, errors.New("get error"))
	f.xkcd.EXPECT().Get(gomock.Any(), 3).Return(info3, nil)

	f.words.EXPECT().Norm(gomock.Any(), "T3 D3").Return([]string{"word"}, nil)

	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 3, URL: "url3", Words: []string{"word"}}).Return(nil)

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update completed with 1 errors")
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}

func TestService_Update_NormPhraseTooLarge_TitleSuccess(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(2, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil) // fetch 2

	info := core.XKCDInfo{ID: 2, URL: "url", Title: "Important Title", Description: "huge description..."}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(info, nil)

	f.words.EXPECT().Norm(gomock.Any(), "Important Title huge description...").Return(nil, core.ErrPhraseTooLarge)
	f.words.EXPECT().Norm(gomock.Any(), "Important Title").Return([]string{"important", "title"}, nil)

	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 2, URL: "url", Words: []string{"important", "title"}}).Return(nil)
	f.broker.EXPECT().NotifyDBUpdate().Return(nil)

	err := f.svc.Update(ctx)
	assert.NoError(t, err)
}

func TestService_Update_NormPhraseTooLarge_TitleFails(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(2, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil)

	info := core.XKCDInfo{ID: 2, URL: "url", Title: "Title", Description: "desc"}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(info, nil)

	f.words.EXPECT().Norm(gomock.Any(), "Title desc").Return(nil, core.ErrPhraseTooLarge)
	f.words.EXPECT().Norm(gomock.Any(), "Title").Return(nil, errors.New("title norm error"))

	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 2, URL: "url", Words: []string{""}}).Return(nil)

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update completed with 1 errors")
}

func TestService_Update_NormReturnsNil_StopWords(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(2, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil)

	info := core.XKCDInfo{ID: 2, URL: "url", Title: "the", Description: "a an"}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(info, nil)

	f.words.EXPECT().Norm(gomock.Any(), "the a an").Return(nil, nil)

	f.db.EXPECT().Add(gomock.Any(), core.Comics{ID: 2, URL: "url", Words: []string{""}}).Return(nil)
	f.broker.EXPECT().NotifyDBUpdate().Return(nil)

	err := f.svc.Update(ctx)
	assert.NoError(t, err)
}

func TestService_Update_DBAddError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(2, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil)

	info := core.XKCDInfo{ID: 2, URL: "url", Title: "T", Description: "D"}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(info, nil)
	f.words.EXPECT().Norm(gomock.Any(), "T D").Return([]string{"x"}, nil)
	f.db.EXPECT().Add(gomock.Any(), gomock.Any()).Return(errors.New("insert error"))

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update completed with 1 errors")
}

func TestService_Update_BrokerNotifyError(t *testing.T) {
	f := setupTest(t, 1)
	ctx := context.Background()

	f.xkcd.EXPECT().LastID(ctx).Return(2, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1}, nil)

	info := core.XKCDInfo{ID: 2, URL: "url", Title: "T", Description: "D"}

	f.xkcd.EXPECT().Get(gomock.Any(), 2).Return(info, nil)
	f.words.EXPECT().Norm(gomock.Any(), "T D").Return([]string{"word"}, nil)
	f.db.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)

	f.broker.EXPECT().NotifyDBUpdate().Return(errors.New("notify error"))

	err := f.svc.Update(ctx)
	assert.NoError(t, err)
}

func TestService_Update_ContextCancelled(t *testing.T) {
	f := setupTest(t, 2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f.xkcd.EXPECT().LastID(ctx).Return(5, nil)
	f.db.EXPECT().IDs(ctx).Return([]int{1, 2, 3}, nil)

	err := f.svc.Update(ctx)
	assert.Error(t, err)
	assert.Equal(t, core.StatusIdle, f.svc.Status(ctx))
}
