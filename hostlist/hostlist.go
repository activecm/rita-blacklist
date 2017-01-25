package hostlist

import "github.com/ocmdev/blacklist/datatypes"

type (
	HostList interface {
		UpdateList(c chan datatypes.BlacklistHost) error
		ValidList(MetaData) bool
		Name() string
		MetaData() MetaData
	}

	MetaData struct {
		Name       string
		Url        string
		LastUpdate int64
		FileHash   string
	}
)

var (
	availableHostLists []HostList
)

func GetAvailableHostLists() []HostList {
	return availableHostLists
}

func AddHostList(newHostList HostList) {
	availableHostLists = append(availableHostLists, newHostList)
}
