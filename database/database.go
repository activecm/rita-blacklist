package database

import (
	"github.com/ocmdev/rita-blacklist/datatypes"
	"github.com/ocmdev/rita-blacklist/hostlist"
)

const HostTableName = "blacklistHosts"
const MetaTableName = "metadata"

type (
	Database interface {

		// Initialize database connection
		Init(host string, port int) error

		// Insert new hosts into the database. Hosts are passed in via a chanel,
		// so this can be threaded with the function retrieving hosts from
		// the blacklist source
		InsertHosts(chan datatypes.BlacklistHost, string)

		// Remove a blacklist source from the database.
		RemoveHostList(hostlist.HostList, string)

		// Check a list of hosts against the database.
		FindHosts([]string, string) []datatypes.QueryResult

		// Retrieve meta data for a specific blacklist source.
		GetMetaData(hostlist.HostList, string) (hostlist.MetaData, error)

		// Register a new blacklist source with the database.
		RegisterHostList(hostlist.HostList, string) error
	}
)
