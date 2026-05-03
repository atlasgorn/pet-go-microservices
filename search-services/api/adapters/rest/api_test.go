package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/core"
	"yadro.com/course/api/core/mocks"
)

type mockAuthenticator struct {
	loginFunc func(string, string) (string, error)
}

func (m *mockAuthenticator) Login(user, password string) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(user, password)
	}
	return "", nil
}

func TestNewMetricsHandler(t *testing.T) {
	handler := rest.NewMetricsHandler()
	assert.NotNil(t, handler)
}

func TestNewLoginHandler_Success(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockAuth := &mockAuthenticator{
		loginFunc: func(user, password string) (string, error) {
			assert.Equal(t, "admin", user)
			assert.Equal(t, "secret", password)
			return "token123", nil
		},
	}

	handler := rest.NewLoginHandler(log, mockAuth)

	body := `{"name":"admin","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "token123", rr.Body.String())
}

func TestNewLoginHandler_WrongMethod(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockAuth := &mockAuthenticator{}

	handler := rest.NewLoginHandler(log, mockAuth)

	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestNewLoginHandler_BadJSON(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockAuth := &mockAuthenticator{}

	handler := rest.NewLoginHandler(log, mockAuth)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{invalid}`))
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestNewLoginHandler_InvalidCredentials(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockAuth := &mockAuthenticator{
		loginFunc: func(user, password string) (string, error) {
			return "", errors.New("invalid")
		},
	}

	handler := rest.NewLoginHandler(log, mockAuth)

	body := `{"name":"admin","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestNewPingHandler_AllOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockPinger := mocks.NewMockPinger(ctrl)

	mockPinger.EXPECT().Ping(gomock.Any()).Return(nil)

	pingers := map[string]core.Pinger{"svc": mockPinger}
	handler := rest.NewPingHandler(log, pingers)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "ok", resp["replies"].(map[string]any)["svc"])
}

func TestNewPingHandler_Unavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockPinger := mocks.NewMockPinger(ctrl)

	mockPinger.EXPECT().Ping(gomock.Any()).Return(errors.New("down"))

	pingers := map[string]core.Pinger{"svc": mockPinger}
	handler := rest.NewPingHandler(log, pingers)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "unavailable", resp["replies"].(map[string]any)["svc"])
}

func TestNewUpdateHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Update(gomock.Any()).Return(nil)

	handler := rest.NewUpdateHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodPost, "/api/db/update", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewUpdateHandler_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Update(gomock.Any()).Return(core.ErrAlreadyExists)

	handler := rest.NewUpdateHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodPost, "/api/db/update", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
}

func TestNewUpdateHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Update(gomock.Any()).Return(errors.New("failed"))

	handler := rest.NewUpdateHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodPost, "/api/db/update", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNewUpdateStatsHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	stats := core.UpdateStats{
		WordsTotal:    100,
		WordsUnique:   50,
		ComicsFetched: 25,
		ComicsTotal:   200,
	}
	mockUpdater.EXPECT().Stats(gomock.Any()).Return(stats, nil)

	handler := rest.NewUpdateStatsHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodGet, "/api/db/stats", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]int
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 100, resp["words_total"])
}

func TestNewUpdateStatsHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Stats(gomock.Any()).Return(core.UpdateStats{}, errors.New("error"))

	handler := rest.NewUpdateStatsHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodGet, "/api/db/stats", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNewUpdateStatusHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Status(gomock.Any()).Return(core.StatusUpdateIdle, nil)

	handler := rest.NewUpdateStatusHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodGet, "/api/db/status", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "idle", resp["status"])
}

func TestNewUpdateStatusHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Status(gomock.Any()).Return(core.StatusUpdateUnknown, errors.New("error"))

	handler := rest.NewUpdateStatusHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodGet, "/api/db/status", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNewDropHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Drop(gomock.Any()).Return(nil)

	handler := rest.NewDropHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodDelete, "/api/db", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewDropHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockUpdater := mocks.NewMockUpdater(ctrl)

	mockUpdater.EXPECT().Drop(gomock.Any()).Return(errors.New("drop failed"))

	handler := rest.NewDropHandler(log, mockUpdater)

	req := httptest.NewRequest(http.MethodDelete, "/api/db", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNewSearchHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(ctrl)

	comics := []core.PbComic{{ID: 1, URL: "https://xkcd.com/1"}}
	mockSearcher.EXPECT().Search(gomock.Any(), 10, "test").Return(comics, nil)

	handler := rest.NewSearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/search?phrase=test&limit=10", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 1, int(resp["total"].(float64)))
}

func TestNewSearchHandler_MissingPhrase(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(gomock.NewController(t))

	handler := rest.NewSearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/search?limit=10", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestNewSearchHandler_InvalidLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(ctrl)
	// Invalid limit defaults to 10, so we need to expect Search call
	comics := []core.PbComic{}
	mockSearcher.EXPECT().Search(gomock.Any(), 10, "test").Return(comics, nil)

	handler := rest.NewSearchHandler(log, mockSearcher)

	// "invalid" parses to 0, which triggers default limit
	req := httptest.NewRequest(http.MethodGet, "/api/search?phrase=test&limit=invalid", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewSearchHandler_ZeroLimit(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(gomock.NewController(t))

	handler := rest.NewSearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/search?phrase=test&limit=0", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestNewSearchHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(ctrl)

	mockSearcher.EXPECT().Search(gomock.Any(), 10, "test").Return(nil, errors.New("search error"))

	handler := rest.NewSearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/search?phrase=test&limit=10", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestNewISearchHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(ctrl)

	comics := []core.PbComic{{ID: 42, URL: "https://xkcd.com/42"}}
	mockSearcher.EXPECT().ISearch(gomock.Any(), 5, "index").Return(comics, nil)

	handler := rest.NewISearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/isearch?phrase=index&limit=5", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, 1, int(resp["total"].(float64)))
}

func TestNewISearchHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mockSearcher := mocks.NewMockSearcher(ctrl)

	mockSearcher.EXPECT().ISearch(gomock.Any(), 10, "fail").Return(nil, errors.New("error"))

	handler := rest.NewISearchHandler(log, mockSearcher)

	req := httptest.NewRequest(http.MethodGet, "/api/isearch?phrase=fail&limit=10", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
