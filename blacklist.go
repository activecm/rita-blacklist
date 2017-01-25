package blacklist

import (
	"log"
	"sync"

	"github.com/ocmdev/rita-blacklist/database"
	"github.com/ocmdev/rita-blacklist/datatypes"
	"github.com/ocmdev/rita-blacklist/hostlist"
)

type (
	BlackList struct {
		Db      database.Database
		Sources []hostlist.HostList
	}
)

func NewBlackList() *BlackList {
	return &BlackList{
		Db:      database.NewMongoDb(),
		Sources: hostlist.GetAvailableHostLists(),
	}
}

// Initialize the blacklist database(s)
func (b *BlackList) Init(host string, port int, dbname string) {

	// Open the database connection
	err := b.Db.Init(host, port)
	if err != nil {
		log.Println("Error opening blacklist database")
		return
	}

	// Iterate over all of the blacklist sources
	for _, val := range b.Sources {

		// Get meta data about this source
		sourceInfo, _ := b.Db.GetMetaData(val, dbname)

		// Ensure the source is still valid
		if !val.ValidList(sourceInfo) {

			log.Println("Updating blacklist source: ", val.Name(), ". This may take a few minutes...")

			// Remove all host entries from this source
			b.Db.RemoveHostList(val, dbname)
			b.Db.RegisterHostList(val, dbname)

			// Chanel for populating database with new hosts.
			c := make(chan datatypes.BlacklistHost)

			// Ask this source for an updated list of hosts.
			var wg1 sync.WaitGroup
			wg1.Add(1)
			go func() {
				val.UpdateList(c)
				wg1.Done()
			}()

			// Insert hosts from the current source into the database.
			var wg2 sync.WaitGroup
			wg2.Add(1)
			go func() {
				b.Db.InsertHosts(c, dbname)
				wg2.Done()
			}()

			// Wait for everything to finish.
			wg1.Wait()
			close(c)
			wg2.Wait()
		}
	}
}

// Check for hosts in the blacklist database.
func (b *BlackList) CheckHosts(hosts []string, dbname string) []datatypes.QueryResult {
	return b.Db.FindHosts(hosts, dbname)
}
