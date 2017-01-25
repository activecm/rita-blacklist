package hostlist

import (
	"archive/zip"
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ocmdev/blacklist/datatypes"
)

type (
	MyIpMs struct {
	}
)

func (m *MyIpMs) downloadFile(fname string) error {
	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get("https://myip.ms/files/blacklist/general/full_blacklist_database.zip")
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)

	return err
}

func (m *MyIpMs) readZipFile(fname string, line chan string) error {
	r, err := zip.OpenReader(fname)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(rc)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			text := scanner.Text()
			line <- text
		}

		rc.Close()
	}

	return nil
}

func (m *MyIpMs) UpdateList(c chan datatypes.BlacklistHost) error {
	err := m.downloadFile("/tmp/myipms_full.zip")
	if err != nil {
		return err
	}

	line := make(chan string)

	// var wg sync.WaitGroup
	// wg.Add(1)
	go func(line chan string) {
		m.readZipFile("/tmp/myipms_full.zip", line)
		// wg.Done()
		close(line)
	}(line)

	for l := range line {
		if len(l) > 0 && l[0] == '#' {
			continue
		}
		split := strings.Split(l, "\t")

		var host datatypes.BlacklistHost
		host.Host = split[0]
		host.HostList = m.Name()
		c <- host
	}

	return nil
}

func (m *MyIpMs) ValidList(mdata MetaData) bool {
	if mdata.LastUpdate < 1 {
		return false
	}

	lastUpdate := time.Unix(mdata.LastUpdate, 0)
	since := time.Since(lastUpdate)
	ret := false
	if since.Hours() < (5 * 24) {
		ret = true
	}

	return ret
}

func (m *MyIpMs) Name() string {
	return "myip-ms"
}

func (m *MyIpMs) MetaData() MetaData {
	var ret MetaData
	ret.Name = m.Name()
	ret.Url = "https://myip.ms/files/blacklist/general/full_blacklist_database.zip"
	ret.LastUpdate = time.Now().Unix()
	return ret
}

func init() {
	AddHostList(NewMyIpMs())
}

func NewMyIpMs() HostList {
	return &MyIpMs{}
}
