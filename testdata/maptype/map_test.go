package maptype

import "testing"

func TestMapFieldConversion(t *testing.T) {
	src := &ConfigRequest{
		Name: "AppConfig",
		Settings: map[string]SettingRequest{
			"debug":   {Value: "true", Enabled: true},
			"timeout": {Value: "30s", Enabled: false},
		},
	}

	dst := &Config{}
	ConvertConfigRequestToConfig(src, dst)

	if dst.Name != "AppConfig" {
		t.Errorf("Name = %q, want %q", dst.Name, "AppConfig")
	}

	if len(dst.Settings) != 2 {
		t.Fatalf("len(Settings) = %d, want 2", len(dst.Settings))
	}

	debug, ok := dst.Settings["debug"]
	if !ok {
		t.Fatal("Settings[\"debug\"] not found")
	}

	if debug.Value != "true" {
		t.Errorf("Settings[\"debug\"].Value = %q, want %q", debug.Value, "true")
	}

	if !debug.Enabled {
		t.Error("Settings[\"debug\"].Enabled should be true")
	}

	timeout, ok := dst.Settings["timeout"]
	if !ok {
		t.Fatal("Settings[\"timeout\"] not found")
	}

	if timeout.Value != "30s" {
		t.Errorf("Settings[\"timeout\"].Value = %q, want %q", timeout.Value, "30s")
	}
}

func TestMapFieldNilMap(t *testing.T) {
	src := &ConfigRequest{
		Name:     "AppConfig",
		Settings: nil,
	}

	dst := &Config{}
	ConvertConfigRequestToConfig(src, dst)

	if dst.Settings != nil {
		t.Error("Settings should be nil when src.Settings is nil")
	}
}

func TestMapFieldEmptyMap(t *testing.T) {
	src := &ConfigRequest{
		Name:     "AppConfig",
		Settings: map[string]SettingRequest{},
	}

	dst := &Config{}
	ConvertConfigRequestToConfig(src, dst)

	if dst.Settings == nil {
		t.Error("Settings should not be nil for empty map")
	}

	if len(dst.Settings) != 0 {
		t.Errorf("len(Settings) = %d, want 0", len(dst.Settings))
	}
}
