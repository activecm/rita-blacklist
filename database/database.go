package database

import (
	"github.com/ocmdev/blacklist/datatypes"
	"github.com/ocmdev/blacklist/hostlist"
)

const HostTableName = "blacklistHosts"
const MetaTableName = "metadata"

type (
	Database interface {
		Init(host string, port int) error
		InsertHosts(chan datatypes.BlacklistHost, string)
		RemoveHostList(hostlist.HostList, string)
		FindHosts([]string, string) []datatypes.QueryResult
		GetMetaData(hostlist.HostList, string) (hostlist.MetaData, error)
		RegisterHostList(hostlist.HostList, string) error
	}
)
