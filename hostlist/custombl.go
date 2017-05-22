package hostlist

import (
  "bufio"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
  "unicode"
  "strconv"
  "net/url"
  "io/ioutil"

	"github.com/ocmdev/rita-blacklist/config"
	"github.com/ocmdev/rita-blacklist/datatypes"
)

type (
  customList struct {
    loc        string
    daysValid  float64
    name       string
  }

  customBlInfo struct {
   host     string
 }
)


// Download the blacklist file
func (m *customList) downloadFile(line chan string) error {
  // retrieve blacklist from url
  resp, err := http.Get(m.loc)
  defer resp.Body.Close()

  if err != nil {
    return err
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }

  blstr := string(body)

  // split blacklist into lines
  blsplit := strings.Split(string(blstr), "\n")
  for _, l := range blsplit {
    line <- l
  }

  return nil
}


// Reads the contents of the custom blacklist file
func (m *customList) readFile(line chan string) error {
  // Open file
  f, err := os.Open(m.loc)
  if err != nil {
    log.Println("Error reading:", err)
    return err
  }

  // Create scanner for file and split into lines
  scanner := bufio.NewScanner(f)
  scanner.Split(bufio.ScanLines)

  // Iterate over lines in file
  for scanner.Scan() {
    text := scanner.Text()

    // Send line into the output channel
    line <- text
  }

  // Close this file
  f.Close()

  return nil
}


// Parse a line from the current dataset
func (m *customList) parseLine(line string) (datatypes.BlacklistHost, error) {
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

  if len(wsRemoved) > 0 && string(wsRemoved[0]) == "#" {
    return ret, errors.New("Comment Line")
  }

  split := strings.Split(wsRemoved, ",")

  if len(split) < 2 {
    return ret, errors.New("Missing Field")
  }

  ip := split[1]
  domain := split[0]

  if domain == "" && ip == "" {
    return ret, errors.New("Missing Field")
  } else if domain == "" {
    domain = "-"
  } else if ip == "" {
    ip = "-"
  }

  ret.Host = ip
  ret.HostList = m.Name()

  ret.Info = customBlInfo {
    host:     domain,
  }

  return ret, nil
}


// Update the list of blacklisted sources
func (m *customList) UpdateList(c chan datatypes.BlacklistHost) error {
  // check if this custom list location is a url or a file
  isUrl, err := m.isValidUrl()

  if err != nil {
    return err
  }

  // create a channel for reading the blacklist
  line := make(chan string)

  if isUrl {
    // read data in a new thread
    go func(line chan string) {
      m.downloadFile(line)
      close(line)
    }(line)
  } else {
    // read data from the file in a new thread
    go func(line chan string) {
      m.readFile(line)
      close(line)
    }(line)
  }

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

  if total == 0 {
    log.Printf("Could not read blacklist %s", m.name)
  } else {
	   log.Printf("Blacklist: %s parsed %d of %d lines in file.", m.name, parseCount, total)
  }

	return nil
}

// Check if the metadata passed in is still considered valid
func (m *customList) ValidList(mdata MetaData) bool {
  // Make sure date last updated is available
  if mdata.LastUpdate < 1 {
    return false
  }

  // Check the time duration since the last time this file was updated
  lastUpdate := time.Unix(mdata.LastUpdate, 0)
  since := time.Since(lastUpdate)
  ret := false
  if since.Hours() < (m.daysValid * 24) {
    ret = true
  }

  return ret
}


// Return the name of this source
func (m *customList) Name() string {
  return m.name
}


// Return meta data about this source (to be stored in database)
func (m *customList) MetaData() MetaData {
  var ret MetaData
  ret.Name = m.name
  ret.Src = m.loc
  ret.LastUpdate = time.Now().Unix()
  return ret
}


// Initialization
func init() {
  // only init if user wants to use a custom blacklist
  conf, ok := config.GetConfig()
  if ok && conf.UseCustomBlacklist {
    // add host list for each list in config file
    for _, src := range conf.CustomBlacklistCfg {
      newlist, ok := newCustomList(&src)
      if ok {
        AddHostList(newlist)
      }
    }
  }
}

// Return a new instance of a custom source
func newCustomList(src *config.CustomBlacklistCfg) (HostList, bool) {
  ret := customList{}
  // parse name from config struct
  ret.name = src.Name

  // parse source location from config struct
  ret.loc = src.Location
  _, err := ret.isValidUrl()
  if err != nil {
    if ret.name != "" {
      log.Printf("Custom blacklist %s has an invalid location", ret.name)
    }
    return &ret, false
  }

  // if there isn't a name just use location
  if ret.name == "" {
    ret.name = ret.loc
  }

  // parse validity time from config struct
  valstr := src.ValTime
  ret.daysValid, err = strconv.ParseFloat(valstr, 64)
  // if it was left blank we don't need an error message
  if err != nil {
    // don't print error if val time was left blank
    if valstr != "" {
      log.Printf("Custom blacklist %s has an invalid validity time. Using default validity time instead.", ret.name)
    }
    ret.daysValid = 36500.0
  }

  return &ret, true
}

// Determines whether the location is a url
func (m *customList) isValidUrl() (bool, error) {
  u, err := url.ParseRequestURI(m.loc)
  if err != nil {
    return false, errors.New("Invalid URL or file path")
  }

  // check the url scheme for validity
  if u.Scheme == "http" || u.Scheme == "https" {
    return true, nil
  }

  // check if it's a file location
  if u.Scheme == "" && u.Path != "" {
    return false, nil
  }

  return true, errors.New("Invalid URL Scheme")
}
