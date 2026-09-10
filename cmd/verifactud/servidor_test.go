package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServidor(t *testing.T) {

	testCases := []struct {
		name           string
		nif            string
		header         string
		expectedStatus int
	}{
		{
			name:           "Valid tenant and token",
			nif:            "89890001K",
			header:         "Bearer valid_token",
			expectedStatus: 204,
		},
		{
			name:           "NIF desconocido",
			nif:            "00000000X",
			header:         "Bearer valid_token",
			expectedStatus: 401,
		},
		{
			name:           "Invalid token",
			nif:            "89890001K",
			header:         "Bearer invalid_token",
			expectedStatus: 401,
		},
		{
			name:           "Missing token",
			nif:            "89890001K",
			header:         "",
			expectedStatus: 401,
		},
	}
	s := &servidor{
		tenants: map[string]TenantConfig{
			"89890001K": {Token: "valid_token"},
		},
	}

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/"+tc.nif+"/alta", nil)

			req.SetPathValue("nif", tc.nif)

			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rec := httptest.NewRecorder()

			s.auth(next).ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}

}
