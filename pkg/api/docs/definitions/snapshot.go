package apidocs

import (
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
)

// swagger:route POST /snapshots snapshot createSnapshot
//
// When creating a snapshot using the API, you have to provide the full dashboard payload including the snapshot data. This endpoint is designed for the Grafana UI.
//
// Snapshot public mode should be enabled or authentication is required.
//
// Responses:
// 200: createSnapshotResponse
// 401: unauthorisedError
// 403: forbiddenError
// 500: internalServerError

// swagger:route GET /dashboard/snapshots snapshot getSnapshots
//
// List snapshots.
//
// Responses:
// 200: getSnapshotsResponse
// 500: internalServerError

// swagger:route GET /snapshots/{key} snapshot getSnapshotByKey
//
// Get Snapshot by Key.
//
// Responses:
// 200: snapshotResponse
// 404: notFoundError
// 500: internalServerError

// swagger:route DELETE /snapshots/{key} snapshot deleteSnapshotByKey
//
// Delete Snapshot by Key.
//
// Responses:
// 200: okResponse
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError

// swagger:route GET /snapshots-delete/{deleteKey} snapshot deleteSnapshotByDeleteKey
//
// Delete Snapshot by deleteKey.
//
// Snapshot public mode should be enabled or authentication is required.
//
// Responses:
// 200: okResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError

// swagger:route GET /snapshot/shared-options snapshot getSnapshotSharingOptions
//
// Get snapshot sharing settings.
//
// Responses:
// 200: getSnapshotSharingOptionsResponse
// 401: unauthorisedError

// swagger:parameters createSnapshot
type CreateSnapshotParam struct {
	// in:body
	Body CreateDashboardSnapshotCommand `json:"body"`
}

// swagger:parameters getSnapshots
type GetSnapshotsParams struct {
	// Search Query
	// in:query
	Query string `json:"query"`
	// Limit the number of returned results
	// in:query
	// default:1000
	Limit int64 `json:"limit"`
}

// swagger:parameters getSnapshotByKey deleteSnapshotByKey
type SnapshotByKeyParam struct {
	// in:path
	Key string `json:"key"`
}

// swagger:parameters deleteSnapshotByDeleteKey
type DeleteSnapshotByDeleteKeyParam struct {
	// in:path
	DeleteKey string `json:"deleteKey"`
}

// swagger:response createSnapshotResponse
type CreateSnapshotResponse struct {
	// in:body
	Body struct {
		// Unique key
		Key string `json:"key"`
		// Unique key used to delete the snapshot. It is different from the key so that only the creator can delete the snapshot.
		DeleteKey string `json:"deleteKey"`
		URL       string `json:"url"`
		DeleteUrl string `json:"deleteUrl"`
		// Snapshot id
		ID int64 `json:"id"`
	} `json:"body"`
}

// swagger:response getSnapshotsResponse
type GetSnapshotsResponse struct {
	// in:body
	Body []*models.DashboardSnapshotDTO `json:"body"`
}

// swagger:response snapshotResponse
type SnapshotResponse DashboardResponse

// swagger:response getSnapshotSharingOptionsResponse
type GetSnapshotSharingOptionsResponse struct {
	// in:body
	Body struct {
		ExternalSnapshotURL  string `json:"externalSnapshotURL"`
		ExternalSnapshotName string `json:"externalSnapshotName"`
		ExternalEnabled      bool   `json:"externalEnabled"`
	} `json:"body"`
}

// CreateDashboardSnapshotCommand is same as models.CreateDashboardSnapshotCommand but with swagger annotations
// swagger:model
type CreateDashboardSnapshotCommand struct {
	// The complete dashboard model.
	// required:true
	Dashboard *simplejson.Json `json:"dashboard" binding:"Required"`
	// Snapshot name
	// required:false
	Name string `json:"name"`
	// When the snapshot should expire in seconds in seconds. Default is never to expire.
	// required:false
	// default:0
	Expires int64 `json:"expires"`
	// Save the snapshot on an external server rather than locally.
	// required:false
	// default: false
	External bool `json:"external"`
	// Define the unique key. Required if `external` is `true`.
	// required:false
	Key string `json:"key"`
	// Unique key used to delete the snapshot. It is different from the `key` so that only the creator can delete the snapshot. Required if `external` is `true`.
	// required:false
	DeleteKey string `json:"deleteKey"`
}
