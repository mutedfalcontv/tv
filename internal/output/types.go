package output

// ErrorResponse is the stable error shape written to stderr in JSON mode.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WarningResponse is written to stderr when a non-fatal warning occurs.
type WarningResponse struct {
	Warning string `json:"warning"`
}

// AdbResult describes an adb connect or disconnect operation.
type AdbResult struct {
	Action string `json:"action"`
	IP     string `json:"ip"`
	Output string `json:"output"`
}

// AppListResult is the stable shape of `tv app list --json`.
type AppListResult struct {
	Packages []string `json:"packages"`
}

// AppActionResult describes an app launch or kill operation.
type AppActionResult struct {
	Action  string `json:"action"`
	Package string `json:"package"`
}

// ConfigResult is the stable shape of `tv config --json`.
type ConfigResult struct {
	ConfigFile    string `json:"config_file"`
	TVIP          string `json:"tv_ip"`
	DefaultPlayer string `json:"default_player"`
}

// InstallResult describes an APK installation.
type InstallResult struct {
	APK       string `json:"apk"`
	Installed bool   `json:"installed"`
	Output    string `json:"output,omitempty"`
}

// LogsResult is the stable shape of a non-streaming `tv logs --json` run.
type LogsResult struct {
	Package string   `json:"package,omitempty"`
	PID     int      `json:"pid,omitempty"`
	Logs    string   `json:"logs"`
	Lines   []string `json:"lines"`
}

// ClearResult describes a log buffer clear.
type ClearResult struct {
	Cleared bool `json:"cleared"`
}

// PlayResult is the stable shape of `tv play --json`.
type PlayResult struct {
	URL        string `json:"url"`
	Player     string `json:"player"`
	TVIP       string `json:"tv_ip"`
	Played     bool   `json:"played"`
	LocalFile  string `json:"local_file,omitempty"`
	ServeURL   string `json:"serve_url,omitempty"`
	ServeMode  string `json:"serve_mode,omitempty"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

// PlayerInfo describes one installed video player.
type PlayerInfo struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

// PlayersResult is the stable shape of `tv player list --json`.
type PlayersResult struct {
	Players       []PlayerInfo `json:"players"`
	DefaultPlayer string       `json:"default_player"`
}

// PlayerDefaultResult is the stable shape of `tv player default --json`.
type PlayerDefaultResult struct {
	DefaultPlayer string `json:"default_player"`
}

// RemoteResult describes a remote key press or text input.
type RemoteResult struct {
	Key     string `json:"key,omitempty"`
	Keycode string `json:"keycode,omitempty"`
	Text    string `json:"text,omitempty"`
	Sent    bool   `json:"sent"`
}

// ServeResult is emitted once by `tv serve --json` before it blocks.
type ServeResult struct {
	URL  string `json:"url"`
	Host string `json:"host"`
	Port int    `json:"port"`
}
