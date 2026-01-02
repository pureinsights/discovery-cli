package cli

import (
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

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
