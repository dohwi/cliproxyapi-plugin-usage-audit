package plugin

import "testing"

func TestParsePluginConfig(t *testing.T) {
	cases := []struct {
		in     string
		dir    string
		strict bool
	}{
		{"", "", false},
		{"enabled: true\npriority: 0\n", "", false},
		{"enabled: true\ndata_dir: /srv/ua\nstrict_decoding: true\n", "/srv/ua", true},
		{"data_dir: \"/a b\"\n", "/a b", false},
		{"nested:\n  data_dir: /no\n", "", false},
		{`{"data_dir":"/j","strict_decoding":true}`, "/j", true},
	}
	for _, c := range cases {
		got := parsePluginConfig([]byte(c.in))
		if got.dataDir != c.dir || got.strict != c.strict {
			t.Errorf("%q: got %+v", c.in, got)
		}
	}
}

func TestRegistrationAdvertisesConfigFields(t *testing.T) {
	if len(pluginConfigFields) < 2 {
		t.Fatal("config fields missing")
	}
}
