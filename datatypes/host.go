package datatypes

type (
	BlacklistHost struct {
		Host     string      `bson:"Host"`
		HostList string      `bson:"HostList"`
		Info     interface{} `bson:"Info"`
	}

	QueryResult struct {
		Host    string
		Results []BlacklistHost
	}
)
