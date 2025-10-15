package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newRequest(tb testing.TB, method string, target string, body io.Reader) *http.Request {
	tb.Helper()
	r, err := http.NewRequest(method, target, body)
	if err != nil {
		tb.Fatalf("newRequest() err = %v, want nil", err)
	}
	return r
}

type assertions struct {
	t   *testing.T
	rec *httptest.ResponseRecorder
}

func (a assertions) status(want int) {
	if code := a.rec.Code; code != want {
		a.t.Errorf("got status code = %d, want %d", code, want)
	}
}

func (a assertions) statusOk() {
	a.status(http.StatusOK)
}

func (a assertions) body(want string) {
	if got := a.rec.Body.String(); !strings.Contains(got, want) {
		a.t.Errorf("\ngot body = %s\nwant contains %s", got, want)
	}
}
