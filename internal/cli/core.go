package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
)

// CoreFileController defines the methods to interact with files.
type CoreFileController interface {
	Upload(key, file string) (gjson.Result, error)
	Retrieve(key string) ([]byte, error)
	List() ([]gjson.Result, error)
	Delete(key string) (gjson.Result, error)
}

// GetFile download an individual file. If directory structure does not exists, it gets created.
func GetFile(client CoreFileController, key string, output string) (gjson.Result, error) {
	file, err := client.Retrieve(key)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not get file with key %q", key)
	}

	fullPath := filepath.Join(output, key)
	os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return gjson.Result{}, NewErrorWithCause(ErrorExitCode, err, "Could not create the necessary directories to write the file %q", fullPath)

	}

	err = os.WriteFile(fullPath, file, 0o644)
	if err != nil {
		return gjson.Result{}, NormalizeWriteFileError(fullPath, err)
	}

	return gjson.Parse(`{"acknowledged": true}`), nil
}

// GetFiles uses the GetFile function to download all the specified files.
func (d discovery) GetFiles(client CoreFileController, keys []string, output string, printer Printer) error {
	var response gjson.Result
	var err error
	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	for _, key := range keys {
		response, err = GetFile(client, key, output)
		if err != nil {
			return err
		}
	}
	return printer(*d.IOStreams(), response)
}

// GetFileList obtains the list of all the available files
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

// recursiveStore is a recursive function that saves everything inside a directory. If the recursive flag
// is set to true, it will also save recursively the file inside the subdirectories found in the directory specified.
func recursiveStore(client CoreFileController, currentkey, startingKey string, recursive bool) (gjson.Result, error) {
	response := gjson.Parse(`{"acknowledged": false}`)

	files, err := os.ReadDir(currentkey)
	if err != nil {
		return response, err
	}

	newKey, err := filepath.Rel(startingKey, currentkey)
	if err != nil {
		return response, err
	}

	for _, file := range files {
		if file.IsDir() {
			if recursive {
				response, err = recursiveStore(client, filepath.Join(currentkey, file.Name()), startingKey, recursive)
				if err != nil {
					return response, err
				}
			}
			continue
		}

		response, err = client.Upload(filepath.ToSlash(filepath.Join(newKey, file.Name())), filepath.Join(currentkey, file.Name()))
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

// StoreFiles is in charge of determine if the key/path specified by the user is from a directory or
// from a file. If it is a directory we assume they want to upload every file inside the directory.
// If it is a file we assume they want to store that individual file
func (d discovery) StoreFiles(client CoreFileController, key string, recursive bool, printer Printer) error {
	response := gjson.Result{}
	fileInfo, err := os.Stat(key)

	if err != nil {
		return NewErrorWithCause(ErrorExitCode, err, "The path %q does not exist", filepath.ToSlash(key))
	}

	if printer == nil {
		printer = JsonObjectPrinter(true)
	}

	if fileInfo.IsDir() {
		response, err = recursiveStore(client, key, key, recursive)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not store directory %q", filepath.ToSlash(key))
		}
	} else {
		response, err = client.Upload(filepath.Base(key), key)
		if err != nil {
			return NewErrorWithCause(ErrorExitCode, err, "Could not store the file with path %q", filepath.ToSlash(key))
		}
	}

	return printer(*d.IOStreams(), response)
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
