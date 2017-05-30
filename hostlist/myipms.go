package hostlist

import (
	"archive/zip"
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ocmdev/rita-blacklist/datatypes"
)

// URL to retrieve this list
const MyIpMsUrl = "https://myip.ms/files/blacklist/general/full_blacklist_database.zip"

// Location to download list
const DownloadLocation = "/tmp/myipms_full.zip"

type (
	MyIpMs struct {
	}

	blInfo struct {
		date        string
		host        string
		country     string
		blacklistId int
	}
)

// Download the blacklist file
func (m *MyIpMs) downloadFile(fname string) error {

	// Create the file to copy data into
	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	// Retrieve the file from the URL
	resp, err := http.Get(MyIpMsUrl)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	// Copy http response into file
	_, err = io.Copy(out, resp.Body)

	return err
}

// Read the contents of the downloaded zip file.
func (m *MyIpMs) readZipFile(fname string, line chan string) error {

	// Open the archive
	r, err := zip.OpenReader(fname)
	if err != nil {
		return err
	}
	defer r.Close()

	// Iterate over files in the archive
	for _, f := range r.File {

		// Open individual file.
		rc, err := f.Open()
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
	}

	return nil
}

// Parse a line from the myip.ms dataset
func (m *MyIpMs) parseLine(line string) (datatypes.BlacklistHost, error) {
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

	if len(wsRemoved) > 0 && wsRemoved[0] == '#' {
		return ret, errors.New("Comment Line")
	}

	final := strings.Replace(wsRemoved, "#", ",", -1)
	split := strings.Split(final, ",")

	if len(split) < 5 {
		return ret, errors.New("Missing Field")
	}

	ret.Host = split[0]
	ret.HostList = m.Name()

	id, err := strconv.Atoi(split[4])
	if err != nil {
		id = -1
	}
	ret.Info = blInfo{
		date:        split[1],
		host:        split[2],
		country:     split[3],
		blacklistId: id,
	}

	return ret, nil
}

// Update the list of blacklisted sources.
func (m *MyIpMs) UpdateList(c chan datatypes.BlacklistHost) error {

	// Download new blacklist file.
	err := m.downloadFile(DownloadLocation)
	if err != nil {
		return err
	}

	// Create a chanel for reading from the file.
	line := make(chan string)

	// Read data from the zip file in a new thread.
	go func(line chan string) {
		m.readZipFile(DownloadLocation, line)
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
func (m *MyIpMs) ValidList(mdata MetaData) bool {

	// Make sure the date last updated is available.
	if mdata.LastUpdate < 1 {
		return false
	}

	// Check the time duration since the last time this file was updated
	// For this list, older than 8 days is not valid.
	lastUpdate := time.Unix(mdata.LastUpdate, 0)
	since := time.Since(lastUpdate)
	ret := false
	if since.Hours() < (8 * 24) {
		ret = true
	}

	return ret
}

// Return the name of this source.
func (m *MyIpMs) Name() string {
	return "myip-ms"
}

// Return meta data about this source (to be stored in database)
func (m *MyIpMs) MetaData() MetaData {
	var ret MetaData
	ret.Name = m.Name()
	ret.Src = MyIpMsUrl
	ret.LastUpdate = time.Now().Unix()
	return ret
}

// Initialization
func init() {
	AddHostList(NewMyIpMs())
}

// Return a new instance of this source
func NewMyIpMs() HostList {
	return &MyIpMs{}
}
