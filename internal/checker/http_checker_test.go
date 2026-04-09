package checker_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thwqsz/uptime-monitor/internal/checker"
)

func TestHTTPChecker_Check_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := checker.NewHTTPChecker(server.Client())

	resp, err := ch.Check(context.Background(), server.URL, 2*time.Second)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Error)
	require.GreaterOrEqual(t, resp.Duration, time.Duration(0))
}

func TestHTTPChecker_Check_HTTP500(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ch := checker.NewHTTPChecker(server.Client())

	resp, err := ch.Check(context.Background(), server.URL, 2*time.Second)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Error)
	require.GreaterOrEqual(t, resp.Duration, time.Duration(0))
}

func TestHTTPChecker_Check_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := checker.NewHTTPChecker(server.Client())

	resp, err := ch.Check(context.Background(), server.URL, 50*time.Millisecond)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Error(t, resp.Error)
	require.Equal(t, 0, resp.StatusCode)
	require.GreaterOrEqual(t, resp.Duration, time.Duration(0))
	require.True(t, errors.Is(resp.Error, context.DeadlineExceeded))
}

func TestHTTPChecker_Check_InvalidURL(t *testing.T) {
	t.Parallel()

	ch := checker.NewHTTPChecker(&http.Client{})

	resp, err := ch.Check(context.Background(), "://bad-url", 2*time.Second)

	require.Error(t, err)
	require.Nil(t, resp)
}
