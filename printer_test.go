package wildcat

import (
	"strings"
	"testing"
)

func createResultSetForTest() *ResultSet {
	argf := NewArgf([]string{"testdata/wc/humpty_dumpty.txt", "testdata/wc/ja/sakura_sakura.txt"}, &ReadOptions{})
	targets, _ := argf.CollectTargets()
	rs, _ := targets.CountAll(func() Counter { return NewCounter(All) })
	return rs
}

func TestXmlPrinter(t *testing.T) {
	writer := new(strings.Builder)
	rs := createResultSetForTest()
	rs.Print(NewPrinter(writer, "xml", &defaultSizer{}))
	result := writer.String()
	if !strings.Contains(result, `<result><file-name>testdata/wc/humpty_dumpty.txt</file-name><lines>4</lines><words>26</words><characters>142</characters><bytes>142</bytes></result>`) {
		t.Errorf("printed xml did not contains the result of humpty_dumpty.txt, got %s", result)
	}
	if !strings.Contains(result, `<result><file-name>testdata/wc/ja/sakura_sakura.txt</file-name><lines>15</lines><words>26</words><characters>118</characters><bytes>298</bytes></result>`) {
		t.Errorf("printed xml did not contains the result of sakura_sakura.txt, got %s", result)
	}
	if !strings.Contains(result, `<result><file-name>total</file-name><lines>19</lines><words>52</words><characters>260</characters><bytes>440</bytes></result>`) {
		t.Errorf("printed xml did not contains the result of total, got %s", result)
	}
}

func TestDefaultPrinter(t *testing.T) {
	writer := new(strings.Builder)
	rs := createResultSetForTest()
	rs.Print(NewPrinter(writer, "unknown", &defaultSizer{}))
	result := writer.String()
	if !strings.Contains(result, `4         26        142        142 testdata/wc/humpty_dumpty.txt`) {
		t.Errorf("the result by DefaultPrinter did not contains humpty_dumpty.txt, got %s", result)
	}
	if !strings.Contains(result, `15         26        118        298 testdata/wc/ja/sakura_sakura.txt`) {
		t.Errorf("the result by DefaultPrinter did not contains sakura_sakura.txt, got %s", result)
	}
	if !strings.Contains(result, `19         52        260        440 total`) {
		t.Errorf("the result by DefaultPrinter did not contains total, got %s", result)
	}
}

func TestJsonPrinter(t *testing.T) {
	writer := new(strings.Builder)
	rs := createResultSetForTest()
	rs.Print(NewPrinter(writer, "json", &defaultSizer{}))
	result := writer.String()
	if !strings.Contains(result, `{"filename":"testdata/wc/humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"}`) {
		t.Errorf("the result by JsonPrinter did not contains humpty_dumpty.txt, got %s", result)
	}
	if !strings.Contains(result, `{"filename":"testdata/wc/ja/sakura_sakura.txt","lines":"15","words":"26","characters":"118","bytes":"298"}`) {
		t.Errorf("the result by JsonPrinter did not contains sakura_sakura.txt, got %s", result)
	}
	if !strings.Contains(result, `{"filename":"total","lines":"19","words":"52","characters":"260","bytes":"440"}`) {
		t.Errorf("the result by JsonPrinter did not contains total, got %s", result)
	}
}

func TestCsvPrinter(t *testing.T) {
	writer := new(strings.Builder)
	rs := createResultSetForTest()
	rs.Print(NewPrinter(writer, "csv", &defaultSizer{}))
	result := writer.String()
	if !strings.Contains(result, `testdata/wc/humpty_dumpty.txt,"4","26","142","142"`) {
		t.Errorf("the result by CsvPrinter did not contains humpty_dumpty.txt, got %s", result)
	}
	if !strings.Contains(result, `testdata/wc/ja/sakura_sakura.txt,"15","26","118","298"`) {
		t.Errorf("the result by CsvPrinter did not contains sakura_sakura.txt, got %s", result)
	}
	if !strings.Contains(result, `total,"19","52","260","440"`) {
		t.Errorf("the result by CsvPrinter did not contains total, got %s", result)
	}
}
