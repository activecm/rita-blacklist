package hostlist

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
  "unicode"

	"github.com/ocmdev/rita-blacklist/datatypes"
)

// URL to retrieve this list
const MdlUrl = "http://www.malwaredomainlist.com/mdlcsv.php"

// Location to download list
const MdlDownloadLoc = "/tmp/mdl_full.csv"

type (
	Mdl struct {
	}
)

// Download the blacklist file
func (m *Mdl) downloadFile(fname string) error {

	// Create the file to copy data into
	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	// Retrieve the file from the URL
	resp, err := http.Get(MdlUrl)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	// Copy http response into file
	_, err = io.Copy(out, resp.Body)

	return err
}

// Read the contents of the downloaded csv file.
func (m *Mdl) readCsvFile(fname string, line chan string) error {
	// Open file.
	rc, err := os.Open(fname)
	if err != nil {
		return err
	}

	// Create scanner for file and split into lines.
	scanner := bufio.NewScanner(rc)
	scanner.Split(bufio.ScanLines)

	// Iterate over lines in file
	for scanner.Scan() {
		text := scanner.Text()

		// Send line into the output chanel
		line <- text
	}

	// Close this file.
	rc.Close()

	return nil
}

// Parse a line from the malware domain list dataset
func (m *Mdl) parseLine(line string) (datatypes.BlacklistHost, error) {
	var ret datatypes.BlacklistHost

	if len(line) < 1 {
		return ret, errors.New("Empty Line")
	}

  wsRemoved := ""
  for _, ch := range line {
    if !unicode.IsSpace(ch) {
      wsRemoved = wsRemoved + string(ch)
    }
  }

  // The first character will be a quotation mark so check second
  if len(wsRemoved) > 1 && wsRemoved[1] == '#' {
    return ret, errors.New("Comment Line")
  }

  // Can't just split on comma in case one of the fields contains a comma
  // Splitting on "," instead will insure we're splitting between fields
  // Also saves us the trouble of trimming quote marks on every field
	split := strings.Split(wsRemoved, "\",\"")

	if len(split) < 9 {
		return ret, errors.New("Missing Field")
	}

	ret.Host = split[2]
	ret.HostList = m.Name()

  host := split[1]
  if host == "-" {
    // If there isn't a host name just give it the ip
    host = ret.Host
  }

	ret.Info = blInfo{
		date:        strings.TrimPrefix(split[0], "\""),
		host:        host,
		country:     strings.TrimSuffix(split[8], "\""),
		blacklistId: -1,
	}

	return ret, nil
}

// Update the list of blacklisted sources.
func (m *Mdl) UpdateList(c chan datatypes.BlacklistHost) error {

	// Download new blacklist file.
	err := m.downloadFile(MdlDownloadLoc)
	if err != nil {
		return err
	}

	// Create a channel for reading from the file.
	line := make(chan string)

	// Read data from the csv file in a new thread.
	go func(line chan string) {
		m.readCsvFile(MdlDownloadLoc, line)
		close(line)
	}(line)

	// Obtain and parse lines from the blacklist file.
	parseCount := 0
	total := 0
	for l := range line {
		host, err := m.parseLine(l)
		if err == nil {
			c <- host
		} else {
			parseCount += 1
		}
		total += 1
	}

	log.Printf("Blacklist: %s parsed %d of %d lines in file.", m.Name(), parseCount, total)

	return nil
}

// Check if the metadata passed in is still considered valid.
func (m *Mdl) ValidList(mdata MetaData) bool {

	// Make sure the date last updated is available.
	if mdata.LastUpdate < 1 {
		return false
	}

	// Check the time duration since the last time this file was updated
	// For this list older than 1 year is not valid
	lastUpdate := time.Unix(mdata.LastUpdate, 0)
	since := time.Since(lastUpdate)
	ret := false
	if since.Hours() < (365 * 24) {
		ret = true
	}

	return ret
}

// Return the name of this source.
func (m *Mdl) Name() string {
	return "Malware Domain List"
}

// Return meta data about this source (to be stored in database)
func (m *Mdl) MetaData() MetaData {
	var ret MetaData
	ret.Name = m.Name()
	ret.Src = MdlUrl
	ret.LastUpdate = time.Now().Unix()
	return ret
}

// Initialization
func init() {
	AddHostList(NewMdl())
}

// Return a new instance of this source
func NewMdl() HostList {
	return &Mdl{}
}
