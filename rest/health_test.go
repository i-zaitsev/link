package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Health(rec, newRequest(t, http.MethodGet, "/", nil))

	assert := assertions{t, rec}
	assert.statusOk()
	assert.body("OK")
}
