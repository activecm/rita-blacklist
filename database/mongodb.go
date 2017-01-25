package database

import (
	"log"
	"strconv"

	"github.com/ocmdev/rita-blacklist/datatypes"
	"github.com/ocmdev/rita-blacklist/hostlist"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	MongoDb struct {
		Session *mgo.Session
	}
)

// Initialize this database connection
func (m *MongoDb) Init(host string, port int) error {

	// Dial the database connection
	session, err := mgo.Dial(host + ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	m.Session = session
	return nil
}

// Insert new hosts into the database. Hosts are passed in via a chanel,
// so this can be threaded with the function retrieving hosts from
// the blacklist source
func (m *MongoDb) InsertHosts(c chan datatypes.BlacklistHost, hostdb string) {
	sess := m.Session.Copy()
	defer sess.Close()

	// Iterate over hosts
	for host := range c {

		// Insert host into database.
		err := sess.DB(hostdb).C(HostTableName).Insert(host)
		if err != nil {
			log.Println("Error inserting blacklisted host: ", err)
		}
	}

	// Setup indices to make sure queries are fast.
	sess.DB(hostdb).C(HostTableName).EnsureIndexKey("$hashed:Host")
	sess.DB(hostdb).C(HostTableName).EnsureIndexKey("$hashed:HostList")
}

// Remove a blacklist source from the database.
func (m *MongoDb) RemoveHostList(hl hostlist.HostList, dbname string) {
	sess := m.Session.Copy()
	defer sess.Close()

	// Remove all hosts that originated from this source.
	sess.DB(dbname).C(HostTableName).RemoveAll(bson.M{"HostList": hl.Name()})

	// Remove source metadata
	sess.DB(dbname).C(MetaTableName).RemoveAll(bson.M{"name": hl.Name()})

}

// Query database for hosts. Return blacklist information
func (m *MongoDb) FindHosts(hosts []string, hostdb string) []datatypes.QueryResult {
	sess := m.Session.Copy()
	defer sess.Close()

	var ret []datatypes.QueryResult

	// Iterate over hosts
	for _, val := range hosts {

		// Query database for host
		iter := sess.DB(hostdb).C(HostTableName).Find(bson.M{"Host": val}).Iter()

		// Obtain query results.
		var result datatypes.BlacklistHost
		var results []datatypes.BlacklistHost
		for iter.Next(&result) {
			results = append(results, result)
		}

		// Append results to result list.
		ret = append(ret, datatypes.QueryResult{Host: val, Results: results})
	}

	return ret
}

// Get metadata for a specific blacklist source
func (m *MongoDb) GetMetaData(hl hostlist.HostList, dbname string) (hostlist.MetaData, error) {
	sess := m.Session.Copy()
	defer sess.Close()

	var ret hostlist.MetaData
	query := sess.DB(dbname).C(MetaTableName).Find(bson.M{"name": hl.Name()})
	err := query.One(&ret)

	return ret, err
}

// Add a new source list to the metadata collection.
func (m *MongoDb) RegisterHostList(hl hostlist.HostList, dbname string) error {
	sess := m.Session.Copy()
	defer sess.Close()
	var ret error

	// Check if source list already exists
	result, _ := m.GetMetaData(hl, dbname)

	if result.Name == "" {
		// Insert new source list into metadata collection.
		ret = sess.DB(dbname).C(MetaTableName).Insert(hl.MetaData())
	}

	return ret
}

// Return a new mongo database instance.
func NewMongoDb() Database {
	mdb := MongoDb{}
	return &mdb
}
