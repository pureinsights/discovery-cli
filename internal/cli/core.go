package cli

import (
	//"fmt"
	"os"
	"path/filepath"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// CoreFileController defines the methods to interact with files.
type CoreFileController interface {
	Upload(key, file string) (gjson.Result, error)
	Retrieve(key string) ([]byte, error)
	List() ([]gjson.Result, error)
	Delete(key string) (gjson.Result, error)
}

// GetFile
func GetFile(client CoreFileController, key string, output string) (gjson.Result, error) {
	file, err := client.Retrieve(key)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not get file with key %q", key)
	}
	
	fullPath := filepath.Join(output,key)
	os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create the necessary directories to write the file %q", fullPath)
		
	}

	err = os.WriteFile(fullPath,file,0o644)
	if err != nil {
		return gjson.Result{}, NormalizeWriteFileError(fullPath, err)
	}

	return gjson.Parse(`{"acknowledged": true}`), nil
}

// GetFiles
func (d discovery) GetFiles(client CoreFileController, keys []string, output string, printer Printer) error {
	var response gjson.Result
	var err error 
	if printer == nil {
		printer = JsonObjectPrinter(true)
	}
	
	for _, key := range keys {
		response, err = GetFile(client,key,output)
		if err != nil {
			return err
		}
	}
	return printer(*d.IOStreams(), response)
}

// GetFileList
func (d discovery) GetFileList(client CoreFileController, printer Printer) error {
	files, err := client.List()
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get file list")
	}
	
	if printer == nil {
		printer = JsonArrayPrinter(false)
	}

	return printer(*d.IOStreams(), files...)
}

// ServerPinger defines the interface to ping servers.
type ServerPinger interface {
	Searcher
	Ping(id uuid.UUID) (gjson.Result, error)
}

// PingServer pings the server with the given name or ID.
func (d discovery) PingServer(client ServerPinger, server string, printer Printer) error {
	serverId, err := GetEntityId(d, client, server)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not get server ID.")
	}

	pingResult, err := client.Ping(serverId)
	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "Could not ping server with id %q", serverId.String())
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	return printer(*d.IOStreams(), pingResult)
}
