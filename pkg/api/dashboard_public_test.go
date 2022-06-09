package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
)

func TestAPIGetPublicDashboard(t *testing.T) {
	t.Run("It should 404 if featureflag is not enabled", func(t *testing.T) {
		sc := setupHTTPServerWithMockDb(t, false, false, featuremgmt.WithFeatures())
		dashSvc := dashboards.NewFakeDashboardService(t)
		dashSvc.On("GetPublicDashboard", mock.Anything, mock.AnythingOfType("string")).
			Return(&models.Dashboard{}, nil).Maybe()
		sc.hs.dashboardService = dashSvc

		setInitCtxSignedInViewer(sc.initCtx)
		response := callAPI(
			sc.server,
			http.MethodGet,
			"/api/public/dashboards",
			nil,
			t,
		)
		assert.Equal(t, http.StatusNotFound, response.Code)
		response = callAPI(
			sc.server,
			http.MethodGet,
			"/api/public/dashboards/asdf",
			nil,
			t,
		)
		assert.Equal(t, http.StatusNotFound, response.Code)
	})

	DashboardUid := "dashboard-abcd1234"
	pubdashUid := "pubdash-abcd1234"

	testCases := []struct {
		Name                  string
		uid                   string
		ExpectedHttpResponse  int
		publicDashboardResult *models.Dashboard
		publicDashboardErr    error
	}{
		{
			Name:                 "It gets a public dashboard",
			uid:                  pubdashUid,
			ExpectedHttpResponse: http.StatusOK,
			publicDashboardResult: &models.Dashboard{
				Data: simplejson.NewFromAny(map[string]interface{}{
					"Uid": DashboardUid,
				}),
			},
			publicDashboardErr: nil,
		},
		{
			Name:                  "It should return 404 if isPublicDashboard is false",
			uid:                   pubdashUid,
			ExpectedHttpResponse:  http.StatusNotFound,
			publicDashboardResult: nil,
			publicDashboardErr:    models.ErrPublicDashboardNotFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			sc := setupHTTPServerWithMockDb(t, false, false, featuremgmt.WithFeatures(featuremgmt.FlagPublicDashboards))
			dashSvc := dashboards.NewFakeDashboardService(t)
			dashSvc.On("GetPublicDashboard", mock.Anything, mock.AnythingOfType("string")).
				Return(test.publicDashboardResult, test.publicDashboardErr)
			sc.hs.dashboardService = dashSvc

			setInitCtxSignedInViewer(sc.initCtx)
			response := callAPI(
				sc.server,
				http.MethodGet,
				fmt.Sprintf("/api/public/dashboards/%v", test.uid),
				nil,
				t,
			)

			assert.Equal(t, test.ExpectedHttpResponse, response.Code)

			if test.publicDashboardErr == nil {
				var dashResp dtos.DashboardFullWithMeta
				err := json.Unmarshal(response.Body.Bytes(), &dashResp)
				require.NoError(t, err)

				assert.Equal(t, DashboardUid, dashResp.Dashboard.Get("Uid").MustString())
				assert.Equal(t, true, dashResp.Meta.IsPublic)
				assert.Equal(t, false, dashResp.Meta.CanEdit)
				assert.Equal(t, false, dashResp.Meta.CanDelete)
				assert.Equal(t, false, dashResp.Meta.CanSave)
			} else {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.Unmarshal(response.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, test.publicDashboardErr.Error(), errResp.Error)
			}
		})
	}
}

func TestAPIGetPublicDashboardConfig(t *testing.T) {
	pdc := &models.PublicDashboard{IsEnabled: true}

	testCases := []struct {
		Name                  string
		DashboardUid          string
		ExpectedHttpResponse  int
		PublicDashboardResult *models.PublicDashboard
		PublicDashboardError  error
	}{
		{
			Name:                  "retrieves public dashboard config when dashboard is found",
			DashboardUid:          "1",
			ExpectedHttpResponse:  http.StatusOK,
			PublicDashboardResult: pdc,
			PublicDashboardError:  nil,
		},
		{
			Name:                  "returns 404 when dashboard not found",
			DashboardUid:          "77777",
			ExpectedHttpResponse:  http.StatusNotFound,
			PublicDashboardResult: nil,
			PublicDashboardError:  models.ErrDashboardNotFound,
		},
		{
			Name:                  "returns 500 when internal server error",
			DashboardUid:          "1",
			ExpectedHttpResponse:  http.StatusInternalServerError,
			PublicDashboardResult: nil,
			PublicDashboardError:  errors.New("database broken"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			sc := setupHTTPServerWithMockDb(t, false, false, featuremgmt.WithFeatures(featuremgmt.FlagPublicDashboards))
			dashSvc := dashboards.NewFakeDashboardService(t)
			dashSvc.On("GetPublicDashboardConfig", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
				Return(test.PublicDashboardResult, test.PublicDashboardError)
			sc.hs.dashboardService = dashSvc

			setInitCtxSignedInViewer(sc.initCtx)
			response := callAPI(
				sc.server,
				http.MethodGet,
				"/api/dashboards/uid/1/public-config",
				nil,
				t,
			)

			assert.Equal(t, test.ExpectedHttpResponse, response.Code)

			if response.Code == http.StatusOK {
				var pdcResp models.PublicDashboard
				err := json.Unmarshal(response.Body.Bytes(), &pdcResp)
				require.NoError(t, err)
				assert.Equal(t, test.PublicDashboardResult, &pdcResp)
			}
		})
	}
}

func TestApiSavePublicDashboardConfig(t *testing.T) {
	testCases := []struct {
		Name                  string
		DashboardUid          string
		publicDashboardConfig *models.PublicDashboard
		ExpectedHttpResponse  int
		saveDashboardError    error
	}{
		{
			Name:                  "returns 200 when update persists",
			DashboardUid:          "1",
			publicDashboardConfig: &models.PublicDashboard{IsEnabled: true},
			ExpectedHttpResponse:  http.StatusOK,
			saveDashboardError:    nil,
		},
		{
			Name:                  "returns 500 when not persisted",
			ExpectedHttpResponse:  http.StatusInternalServerError,
			publicDashboardConfig: &models.PublicDashboard{},
			saveDashboardError:    errors.New("backend failed to save"),
		},
		{
			Name:                  "returns 404 when dashboard not found",
			ExpectedHttpResponse:  http.StatusNotFound,
			publicDashboardConfig: &models.PublicDashboard{},
			saveDashboardError:    models.ErrDashboardNotFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			sc := setupHTTPServerWithMockDb(t, false, false, featuremgmt.WithFeatures(featuremgmt.FlagPublicDashboards))

			dashSvc := dashboards.NewFakeDashboardService(t)
			dashSvc.On("SavePublicDashboardConfig", mock.Anything, mock.AnythingOfType("*dashboards.SavePublicDashboardConfigDTO")).
				Return(&models.PublicDashboard{IsEnabled: true}, test.saveDashboardError)
			sc.hs.dashboardService = dashSvc

			setInitCtxSignedInViewer(sc.initCtx)
			response := callAPI(
				sc.server,
				http.MethodPost,
				"/api/dashboards/uid/1/public-config",
				strings.NewReader(`{ "isPublic": true }`),
				t,
			)

			assert.Equal(t, test.ExpectedHttpResponse, response.Code)

			// check the result if it's a 200
			if response.Code == http.StatusOK {
				val, err := json.Marshal(test.publicDashboardConfig)
				require.NoError(t, err)
				assert.Equal(t, string(val), response.Body.String())
			}
		})
	}
}
