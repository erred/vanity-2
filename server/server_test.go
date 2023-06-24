package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr/testr"
	"go.seankhliao.com/svcrunner"
)

func TestServeHTTP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo-bar", nil)
	rec := httptest.NewRecorder()

	ctx := context.Background()
	svr := New(&http.Server{})
	svr.Init(ctx, svcrunner.Tools{
		Log: testr.New(t),
	})

	svr.ServeHTTP(rec, req)
	res := rec.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", res.Status, http.StatusText(http.StatusOK))
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("read body err = %v", err)
	}
	if !bytes.Contains(b, []byte(`go.seankhliao.com/foo-bar git https://github.com/seankhliao/foo-bar`)) {
		t.Errorf("missing go-import content: %s", b)
	}
}
