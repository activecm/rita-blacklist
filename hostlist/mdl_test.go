package hostlist

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/ocmdev/rita-blacklist/datatypes"
)

/*type testcase struct {
	line string
	err  bool
}*/

var mdlTests = []testcase{
	{"\"2001/01/02_00:00\",\"-\",\"1.2.3.4\",\"-\",\"Trojan\",\"-\",\"33314\",\"0\",\"US\",", false},
	{"\"#2001/01/02_00:00\",\"-\",\"1.2.3.4\",\"-\",\"Trojan\",\"-\",\"33314\",\"0\",\"US\"", true},
	{"\"2001/01/02_00:00\",\"-\",\"1.2.3.4\",\"-\",\"Trojan\",\"-\",", true},
	{"\"2001/01/02_00:00\",\"url.com\",\"1.2.3.4\",\"-\",\"Trojan\",\"-\",\"33314\",\"0\",\"US\",", false},
	{"", true},
}

func formatMdlResult(blhost datatypes.BlacklistHost) string {
	v := reflect.ValueOf(blhost)

	var ret string
	ret += "\n"

	for i := 0; i < v.NumField(); i++ {
		ret += fmt.Sprintf("\t%s:\t%v\n", v.Type().Field(i).Name, v.Field(i).Interface())
	}

	return ret
}

func TestMdlParse(t *testing.T) {
	m := Mdl{}
	for _, test := range mdlTests {
		res, err := m.parseLine(test.line)
		t.Logf("---> \"%s\"\n\tExpected Error:%s%s", test.line, strconv.FormatBool(test.err), formatMdlResult(res))
		if (err != nil) != test.err {
			t.Fail()
		}
	}
}
