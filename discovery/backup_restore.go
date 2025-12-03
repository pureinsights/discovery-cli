package discovery

import (
	"mime"
	"net/http"

	"github.com/tidwall/gjson"
)

// BckupRestore is the struct that exports and imports entities.
type backupRestore struct {
	client
}

// OnConflict is a type that is used to declare constants that represent the three conflict resolution options.
type OnConflict string

// The constants represent the options to ignore the new duplicated entities, fail if there are duplicated entities, and update the duplicated entities with the new values.
const (
	OnConflictIgnore OnConflict = "IGNORE"
	OnConflictFail   OnConflict = "FAIL"
	OnConflictUpdate OnConflict = "UPDATE"
)

// Export obtains the bytes with the information of the exported entities.
// This data can later be written to a ZIP file with the filename received in the Content Disposition header.
func (backup backupRestore) Export() ([]byte, string, error) {
	c := backup.client
	request := c.client.R()

	if c.ApiKey != "" {
		request.SetHeader("X-API-Key", c.ApiKey)
	}

	response, err := request.Execute(http.MethodGet, c.client.BaseURL+"/export")
	if err != nil {
		return nil, "", err
	}

	if response.IsError() {
		return nil, "", Error{
			Status: response.StatusCode(),
			Body:   gjson.ParseBytes(response.Body()),
		}
	}

	contentDisposition := response.Header().Get("Content-Disposition")
	filename := "discovery.zip"

	if contentDisposition != "" {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			if value := params["filename"]; value != "" {
				filename = value
			}
		} else {
			return nil, filename, err
		}
	}

	return response.Body(), filename, nil
}

// Import reads the given file containing the entities to be imported, and then calls the endpoint to do so.
// It sets the conflict resolution strategy to the one sent as a parameter and returns the status of the imported entities.
func (restore backupRestore) Import(onConflict OnConflict, file string) (gjson.Result, error) {
	return execute(restore.client, http.MethodPost, "/import", WithFile(file), WithQueryParameters(map[string][]string{"onConflict": {string(onConflict)}}))
}
