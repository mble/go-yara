package yara

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func makeScanner(t *testing.T, rule string) *Scanner {
	c, err := NewCompiler()
	if c == nil || err != nil {
		t.Fatal("NewCompiler():", err)
	}
	if err = c.AddString(rule, ""); err != nil {
		t.Fatal("AddString():", err)
	}
	r, err := c.GetRules()
	if err != nil {
		t.Fatal("GetRules:", err)
	}
	s, err := NewScanner(r)
	if err != nil {
		t.Fatal("NewScanner:", err)
	}
	return s
}

func TestScannerSimpleMatch(t *testing.T) {
	s := makeScanner(t,
		"rule test : tag1 { meta: author = \"Matt Blewitt\" strings: $a = \"abc\" fullword condition: $a }")
	m, err := s.ScanMem([]byte(" abc "), 0, 0)
	if err != nil {
		t.Errorf("ScanMem: %s", err)
	}
	t.Logf("Matches: %+v", m)
}

func TestScannerSimpleFileMatch(t *testing.T) {
	s := makeScanner(t,
		"rule test : tag1 { meta: author = \"Matt Blewitt\" strings: $a = \"abc\" fullword condition: $a }")
	tf, _ := ioutil.TempFile("", "TestScannerSimpleFileMatch")
	defer os.Remove(tf.Name())
	tf.Write([]byte(" abc "))
	tf.Close()
	m, err := s.ScanFile(tf.Name(), 0, 0)
	if err != nil {
		t.Errorf("ScanFile(%s): %s", tf.Name(), err)
	}
	t.Logf("Matches: %+v", m)
}

func TestScannerSimpleFileDescriptorMatch(t *testing.T) {
	s := makeScanner(t,
		"rule test : tag1 { meta: author = \"Matt Blewitt\" strings: $a = \"abc\" fullword condition: $a }")
	tf, _ := ioutil.TempFile("", "TestScannerSimpleFileDescriptorMatch")
	defer os.Remove(tf.Name())
	tf.Write([]byte(" abc "))
	tf.Seek(0, os.SEEK_SET)
	m, err := s.ScanFileDescriptor(tf.Fd(), 0, 0)
	if err != nil {
		t.Errorf("ScanFileDescriptor(%v): %s", tf.Fd(), err)
	}
	t.Logf("Matches: %+v", m)

}

// TestScannerIndependence tests that two scanners can
// execute with different external variables and the same ruleset
func TestScannerIndependence(t *testing.T) {
	rulesStr := `
		rule test {
			condition: bool_var and int_var == 1 and str_var == "foo"
		}
	`

	c, err := NewCompiler()
	if c == nil || err != nil {
		t.Fatal("NewCompiler():", err)
	}

	c.DefineVariable("bool_var", false)
	c.DefineVariable("int_var", 0)
	c.DefineVariable("str_var", "")

	if err = c.AddString(rulesStr, ""); err != nil {
		t.Fatal("AddString():", err)
	}

	r, err := c.GetRules()
	if err != nil {
		t.Fatal("GetRules:", err)
	}

	s1, err := NewScanner(r)
	if err != nil {
		t.Fatal("NewScanner:", err)
	}

	s2, err := NewScanner(r)
	if err != nil {
		t.Fatal("NewScanner:", err)
	}

	s1.DefineVariable("bool_var", true)
	s1.DefineVariable("int_var", 1)
	s1.DefineVariable("str_var", "foo")

	s2.DefineVariable("bool_var", false)
	s2.DefineVariable("int_var", 2)
	s2.DefineVariable("str_var", "bar")

	m1, err := s1.ScanMem([]byte(""), 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	m2, err := s2.ScanMem([]byte(""), 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !(len(m1) > 0) {
		t.Errorf("wanted >0 matches, got %d", len(m1))
	}

	if len(m2) != 0 {
		t.Errorf("wanted 0 matches, got %d", len(m2))
	}

	t.Logf("Matches 1: %+v", m1)
	t.Logf("Matches 2: %+v", m2)
}

func TestScannerImportDataCallback(t *testing.T) {
	cb := newTestCallback(t)
	s := makeScanner(t, `
		import "tests"
		import "pe"
		rule t1 { condition: true }
		rule t2 { condition: false }
		rule t3 {
			condition: tests.module_data == "callback-data-for-tests-module"
		}`)
	if err := s.ScanMemWithCallback([]byte(""), 0, 0, cb); err != nil {
		t.Error(err)
	}
	for _, module := range []string{"tests", "pe"} {
		if _, ok := cb.modules[module]; !ok {
			t.Errorf("ImportModule was not called for %s", module)
		}
	}
	for _, rule := range []string{"t1", "t3"} {
		if _, ok := cb.matched["t1"]; !ok {
			t.Errorf("RuleMatching was not called for %s", rule)
		}
	}
	if _, ok := cb.notMatched["t2"]; !ok {
		t.Errorf("RuleNotMatching was not called for %s", "t2")
	}
	if !cb.finished {
		t.Errorf("ScanFinished was not called")
	}
	runtime.GC()
}
