package agent

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigBindAddrParts(t *testing.T) {
	testCases := []struct {
		Value string
		IP    string
		Port  int
		Error bool
	}{
		{"0.0.0.0", "0.0.0.0", DefaultBindPort, false},
		{"0.0.0.0:1234", "0.0.0.0", 1234, false},
	}

	for _, tc := range testCases {
		c := &Config{BindAddr: tc.Value}
		ip, port, err := c.BindAddrParts()
		if tc.Error != (err != nil) {
			t.Errorf("Bad error: %s", err)
			continue
		}

		if tc.IP != ip {
			t.Errorf("%s: Got IP %#v", tc.Value, ip)
			continue
		}

		if tc.Port != port {
			t.Errorf("%s: Got port %d", tc.Value, port)
			continue
		}
	}
}

func TestConfigEventScripts(t *testing.T) {
	c := &Config{
		EventHandlers: []string{
			"foo.sh",
			"bar=blah.sh",
		},
	}

	result, err := c.EventScripts()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result) != 2 {
		t.Fatalf("bad: %#v", result)
	}

	expected := []EventScript{
		{"*", "", "foo.sh"},
		{"bar", "", "blah.sh"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}
}

func TestDecodeConfig(t *testing.T) {
	input := `{"node_name": "foo"}`
	config, err := DecodeConfig(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.NodeName != "foo" {
		t.Fatalf("bad: %#v", config)
	}
}

func TestMergeConfig(t *testing.T) {
	a := &Config{
		NodeName:      "foo",
		Role:          "bar",
		EventHandlers: []string{"foo"},
		StartJoin:     []string{"foo"},
	}

	b := &Config{
		NodeName:      "bname",
		EventHandlers: []string{"bar"},
		StartJoin:     []string{"bar"},
	}

	c := MergeConfig(a, b)

	if c.NodeName != "bname" {
		t.Fatalf("bad: %#v", c)
	}

	if c.Role != "bar" {
		t.Fatalf("bad: %#v", c)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual(c.EventHandlers, expected) {
		t.Fatalf("bad: %#v", c)
	}

	if !reflect.DeepEqual(c.StartJoin, expected) {
		t.Fatalf("bad: %#v", c)
	}
}

func TestReadConfigPaths_badPath(t *testing.T) {
	_, err := ReadConfigPaths([]string{"/i/shouldnt/exist/ever/rainbows"})
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestReadConfigPaths_file(t *testing.T) {
	tf, err := ioutil.TempFile("", "serf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(`{"node_name":"bar"}`))
	tf.Close()
	defer os.Remove(tf.Name())

	config, err := ReadConfigPaths([]string{tf.Name()})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.NodeName != "bar" {
		t.Fatalf("bad: %#v", config)
	}
}

func TestReadConfigPaths_dir(t *testing.T) {
	td, err := ioutil.TempDir("", "serf")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	err = ioutil.WriteFile(filepath.Join(td, "a.json"),
		[]byte(`{"node_name": "bar"}`), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(td, "b.json"),
		[]byte(`{"node_name": "baz"}`), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// A non-json file, shouldn't be read
	err = ioutil.WriteFile(filepath.Join(td, "c"),
		[]byte(`{"node_name": "bad"}`), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config, err := ReadConfigPaths([]string{td})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.NodeName != "baz" {
		t.Fatalf("bad: %#v", config)
	}
}
