package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerListPassesSupportedSortToService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantSort  string
		wantOrder string
	}{
		{
			name:      "usage_sort",
			query:     "/api/v1/admin/accounts?sort_by=usage&sort_order=asc",
			wantSort:  "usage",
			wantOrder: "asc",
		},
		{
			name:      "name_sort",
			query:     "/api/v1/admin/accounts?sort_by=name&sort_order=desc",
			wantSort:  "name",
			wantOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			adminSvc := newStubAdminService()
			handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			router.GET("/api/v1/admin/accounts", handler.List)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, tt.wantSort, adminSvc.lastListAccounts.sortBy)
			require.Equal(t, tt.wantOrder, adminSvc.lastListAccounts.sortOrder)
		})
	}
}
