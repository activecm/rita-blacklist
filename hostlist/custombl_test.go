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

// getting false with an err means an invalid file or URL path,
// getting true with an err means an invalid scheme on the URL (we only accept http and https)
// getting true without an err means a valid URL
// getting false without an err means a valid file path
var customIsURLTests = []urltestcase{
  {"/file.txt", false, false},
  {"/folder/file.txt", false, false},
	{"/folder/folder/folder/file.txt", false, false},
  {"domain.com", false, true},
	{"https://www.domain.com/file.csv", true, false},
	{"scheme://domain.com/file.csv", true, true},
  {"http://domain.com", true, false},
	{"", false, true},
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

func FormatUrlResult(isUrl bool, e bool) string {
	var ret string
	ret += "\n\n"
	ret += fmt.Sprintf("\tIs URL:\t\t%s\n\tGot Error:\t%s", strconv.FormatBool(isUrl), strconv.FormatBool(e))
	return ret
}

func TestIsUrl(t *testing.T) {
	for _, test := range customIsURLTests {
    m := customList{loc: test.line}
		isUrl, err := m.isValidUrl()
		gotErr := err != nil
		t.Logf("---> \"%s\"\n\tExpected URL:\t%s\n\tExpected Error:\t%s%s", test.line, strconv.FormatBool(test.isUrl), strconv.FormatBool(test.err), FormatUrlResult(isUrl, gotErr))
		if isUrl != test.isUrl || gotErr != test.err {
			t.Fail()
		}
	}
}
