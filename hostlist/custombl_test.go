package hostlist

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/ocmdev/rita-blacklist/datatypes"
)

type urltestcase struct {
	line string
	isUrl  bool
	err		 bool
}

var customParseTests = []testcase{
	{"domain.com, 1.2.3.4, US", false},
	{"#domain.com, 1.2.3.4, US", true},
	{"domain.com, 1.2.3.4", true},
	{"domain.com, 1.2.3.4, US, some other field", false},
	{"", true},
}

var customIsURLTests = []urltestcase{
  {"/file.txt", false, false},
  {"/folder/file.txt", false, false},
  {"domain.com", false, false},
	{"https://www.domain.com/file.csv", true, false},
	{"scheme://domain.com/file.csv", true, true},
  {"http://domain.com", true, false},
}


func formatCustomResult(blhost datatypes.BlacklistHost) string {
	v := reflect.ValueOf(blhost)

	var ret string
	ret += "\n"

	for i := 0; i < v.NumField(); i++ {
		ret += fmt.Sprintf("\t%s:\t%v\n", v.Type().Field(i).Name, v.Field(i).Interface())
	}

	return ret
}

func TestCustomParse(t *testing.T) {
	m := customList{name: "name"}
	for _, test := range customParseTests {
		res, err := m.parseLine(test.line)
		t.Logf("---> \"%s\"\n\tExpected Error:%s%s", test.line, strconv.FormatBool(test.err), formatCustomResult(res))
		if (err != nil) != test.err {
			t.Fail()
		}
	}
}

func TestIsUrl(t *testing.T) {
	for _, test := range customIsURLTests {
    m := customList{loc: test.line}
		isUrl, err := m.isValidUrl()
		t.Logf("---> \"%s\"\n\tExpected isUrl Result:%s%s", test.line, strconv.FormatBool(test.isUrl),"---> \"%s\"\n\tExpected Error:%s%s", strconv.FormatBool(test.err))
		if isUrl != test.isUrl && (err != nil) != test.err {
			t.Fail()
		}
	}
}
