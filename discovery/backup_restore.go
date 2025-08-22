package discovery

import (
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

type backupRestore struct {
	client
}

func (backup backupRestore) Export() ([]byte, error) {

	export, err := backup.execute(http.MethodGet, "/export")
	if err != nil {
		return []byte{}, err
	}

	return export, nil
}

func (restore backupRestore) Import(name string, data io.Reader) (gjson.Result, error) {

	summary, err := execute(restore.client, http.MethodGet, "/summary")
	if err != nil {
		return gjson.Result{}, err
	}

	return summary, nil
}
