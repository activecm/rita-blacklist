package hostlist

import "github.com/ocmdev/rita-blacklist/datatypes"

type (

	// Interface for processing blacklist sources
	HostList interface {
		UpdateList(c chan datatypes.BlacklistHost) error
		ValidList(MetaData) bool
		Name() string
		MetaData() MetaData
	}

	// Information about a blacklist source
	MetaData struct {
		Name       string
		Src        string
		LastUpdate int64
		FileHash   string
	}
)

var (
	availableHostLists []HostList
)

// Return a list of available blacklist sources
func GetAvailableHostLists() []HostList {
	return availableHostLists
}

// Add source to list of blacklist sources
func AddHostList(newHostList HostList) {
	availableHostLists = append(availableHostLists, newHostList)
}
