package discovery

import (
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
// This data can later be written to a ZIP file.
func (backup backupRestore) Export() ([]byte, error) {
	return backup.execute(http.MethodGet, "/export")
}

// Import reads the given file containing the entities to be imported, and then calls the endpoint to do so.
// It sets the conflict resolution strategy to the one sent as a parameter and returns the status of the imported entities.
func (restore backupRestore) Import(onConflict OnConflict, file string) (gjson.Result, error) {
	return execute(restore.client, http.MethodPost, "/import", WithFile(file), WithQueryParameters(map[string][]string{"onConflict": {string(onConflict)}}))
}
