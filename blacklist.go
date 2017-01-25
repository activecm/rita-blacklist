package blacklist

import (
	"fmt"
	"log"
	"sync"

	"github.com/ocmdev/blacklist/database"
	"github.com/ocmdev/blacklist/datatypes"
	"github.com/ocmdev/blacklist/hostlist"
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

func (b *BlackList) Init(host string, port int, dbname string) {
	err := b.Db.Init(host, port)
	if err != nil {
		log.Println("Error opening blacklist database")
		return
	}

	for _, val := range b.Sources {
		sourceInfo, _ := b.Db.GetMetaData(val, dbname)
		if !val.ValidList(sourceInfo) {
			b.Db.RemoveHostList(val, dbname)
			b.Db.RegisterHostList(val, dbname)

			c := make(chan datatypes.BlacklistHost)

			var wg1 sync.WaitGroup
			wg1.Add(1)

			go func() {
				val.UpdateList(c)
				wg1.Done()
			}()

			var wg2 sync.WaitGroup
			wg2.Add(1)
			go func() {
				b.Db.InsertHosts(c, dbname)
				wg2.Done()
			}()

			wg1.Wait()
			close(c)
			wg2.Wait()

		}
	}
}

func (b *BlackList) CheckHosts(hosts []string, dbname string) {
	fmt.Println(b.Db.FindHosts(hosts, dbname))
}
