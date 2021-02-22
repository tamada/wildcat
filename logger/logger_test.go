package logger

import (
	"log"
	"strings"
	"testing"
)

func TestLogLevel(t *testing.T) {
	testdata := []struct {
		giveLevel string
		wontLevel Level
	}{
		{"warn", WARN},
		{"info", INFO},
		{"Debug", DEBUG},
		{"FATAL", FATAL},
		{"hoge", WARN},
	}
	for _, td := range testdata {
		gotLevel := LogLevel(td.giveLevel)
		if gotLevel != td.wontLevel {
			t.Errorf(`LogLevel("%s") did not match wont %d, got %d`, td.giveLevel, td.wontLevel, gotLevel)
		}
	}
}

func TestLogf(t *testing.T) {
	testdata := []struct {
		giveLevel    Level
		giveFormat   string
		params       []interface{}
		wontSuffixes []string
	}{
		{DEBUG, "this is %s", []interface{}{"test"}, []string{"this is test", "this is test", "this is test", "this is test"}},
		{INFO, "this is %s", []interface{}{"test"}, []string{"", "this is test", "this is test", "this is test"}},
		{WARN, "this is %s", []interface{}{"test"}, []string{"", "", "this is test", "this is test"}},
		{FATAL, "this is %s", []interface{}{"test"}, []string{"", "", "", "this is test"}},
	}
	for _, td := range testdata {
		list := []struct {
			label string
			f     func(format string, v ...interface{})
		}{
			{"Debugf", Debugf},
			{"Infof", Infof},
			{"Warnf", Warnf},
			{"Fatalf", Fatalf},
		}
		SetLevel(td.giveLevel)
		for i, suffix := range td.wontSuffixes {
			sw := &strings.Builder{}
			log.SetOutput(sw)
			list[i].f(td.giveFormat, td.params...)
			gotString := strings.TrimSpace(sw.String())
			if (suffix == "" && gotString != "") || !strings.HasSuffix(gotString, suffix) {
				t.Errorf(`%s("%s") (level %s): result did not match, wont suffix "%s", but "%s"`, list[i].label, td.giveFormat, td.giveLevel, suffix, gotString)
			}
		}
	}
}

func TestLog(t *testing.T) {
	testdata := []struct {
		giveLevel    Level
		giveMessage  string
		wontSuffixes []string
	}{
		{DEBUG, "this is test", []string{"this is test", "this is test", "this is test", "this is test"}},
		{INFO, "this is test", []string{"", "this is test", "this is test", "this is test"}},
		{WARN, "this is test", []string{"", "", "this is test", "this is test"}},
		{FATAL, "this is test", []string{"", "", "", "this is test"}},
	}
	for _, td := range testdata {
		list := []struct {
			label string
			f     func(message string)
		}{
			{"Debug", Debug},
			{"Info", Info},
			{"Warn", Warn},
			{"Fatal", Fatal},
		}
		SetLevel(td.giveLevel)
		for i, suffix := range td.wontSuffixes {
			sw := &strings.Builder{}
			log.SetOutput(sw)
			list[i].f(td.giveMessage)
			gotString := strings.TrimSpace(sw.String())
			if (suffix == "" && gotString != "") || !strings.HasSuffix(gotString, suffix) {
				t.Errorf(`%s: (%s) result did not match, wont suffix "%s", but "%s"`, list[i].label, td.giveLevel, suffix, gotString)
			}
		}
	}
}
func TestSetLevel(t *testing.T) {
	SetLevel(WARN)
	if GetLevel() != WARN {
		t.Errorf("Default log level wont %s, got %s", WARN, GetLevel())
	}

	SetLevel(DEBUG)
	if GetLevel() != DEBUG {
		t.Errorf("Default log level wont %s, got %s", DEBUG, GetLevel())
	}

	SetLevel(FATAL + 1)
	if GetLevel() != DEBUG {
		t.Errorf("Default log level wont %s, got %s", DEBUG, GetLevel())
	}
}

func TestStringer(t *testing.T) {
	if DEBUG.String() != "DEBUG" {
		t.Errorf("DEBUG.String() did not match, wont DEBUG, got %s", DEBUG.String())
	}
	if INFO.String() != "INFO" {
		t.Errorf("DEBUG.String() did not match, wont INFO, got %s", INFO.String())
	}
	if WARN.String() != "WARN" {
		t.Errorf("DEBUG.String() did not match, wont WARN, got %s", WARN.String())
	}
	if FATAL.String() != "FATAL" {
		t.Errorf("DEBUG.String() did not match, wont FATAL, got %s", FATAL.String())
	}
	if Level(-1).String() != "UNKNOWN" {
		t.Errorf("UNKNOWN.String() did not match, wont UNKNOWN, got %s", Level(-1).String())
	}
}
