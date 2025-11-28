package maptype

// Source types
type ConfigRequest struct {
	Name     string
	Settings map[string]SettingRequest
}

type SettingRequest struct {
	Value   string
	Enabled bool
}

// Target types
type Config struct {
	Name     string
	Settings map[string]Setting
}

type Setting struct {
	Value   string
	Enabled bool
}
