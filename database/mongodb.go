package database

import (
	"log"
	"strconv"

	"github.com/ocmdev/blacklist/datatypes"
	"github.com/ocmdev/blacklist/hostlist"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	MongoDb struct {
		Session *mgo.Session
	}
)

func (m *MongoDb) Init(host string, port int) error {
	session, err := mgo.Dial(host + ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	m.Session = session
	return nil
}

func (m *MongoDb) InsertHosts(c chan datatypes.BlacklistHost, hostdb string) {
	sess := m.Session.Copy()
	defer sess.Close()

	for host := range c {
		err := sess.DB(hostdb).C(HostTableName).Insert(host)
		if err != nil {
			log.Println("Error inserting blacklisted host: ", err)
		}
	}

	sess.DB(hostdb).C(HostTableName).EnsureIndexKey("$hashed:Host")
	sess.DB(hostdb).C(HostTableName).EnsureIndexKey("$hashed:HostList")

}

func (m *MongoDb) RemoveHostList(hl hostlist.HostList, dbname string) {
	sess := m.Session.Copy()
	defer sess.Close()

	sess.DB(dbname).C(HostTableName).RemoveAll(bson.M{"HostList": hl.Name()})
	sess.DB(dbname).C(MetaTableName).RemoveAll(bson.M{"name": hl.Name()})

}

func (m *MongoDb) FindHosts(hosts []string, hostdb string) []datatypes.QueryResult {
	sess := m.Session.Copy()
	defer sess.Close()

	var ret []datatypes.QueryResult

	for _, val := range hosts {
		iter := sess.DB(hostdb).C(HostTableName).Find(bson.M{"Host": val}).Iter()

		var result datatypes.BlacklistHost
		var results []datatypes.BlacklistHost
		for iter.Next(&result) {
			results = append(results, result)
		}

		ret = append(ret, datatypes.QueryResult{Host: val, Results: results})
	}

	return ret
}

func (m *MongoDb) GetMetaData(hl hostlist.HostList, dbname string) (hostlist.MetaData, error) {
	sess := m.Session.Copy()
	defer sess.Close()

	var ret hostlist.MetaData

	query := sess.DB(dbname).C(MetaTableName).Find(bson.M{"name": hl.Name()})

	err := query.One(&ret)

	return ret, err
}

func (m *MongoDb) RegisterHostList(hl hostlist.HostList, dbname string) error {
	sess := m.Session.Copy()
	defer sess.Close()
	var ret error

	result, _ := m.GetMetaData(hl, dbname)

	if result.Name == "" {
		ret = sess.DB(dbname).C(MetaTableName).Insert(hl.MetaData())
	}

	return ret
}

func NewMongoDb() Database {
	mdb := MongoDb{}

	return &mdb
}
