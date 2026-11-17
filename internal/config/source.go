package config

import (
	"encoding/json"
	"os"
)

// Source says where the effective configuration came from.
//
//	SourceDefaults  no config file existed; built-in defaults apply and saving is safe
//	SourceFile      a config file was read and parsed successfully
//	SourceInvalid   a config file existed but could not be read or parsed
//
// SourceInvalid must not be treated as "defaults": the file on disk may still
// hold the user's real settings, so a save could destroy them. Callers that
// recover from errors should warn and refuse to write rather than persist
// defaults over an unreadable file.
type Source int

const (
	SourceDefaults Source = iota
	SourceFile
	SourceInvalid
)

func (s Source) String() string {
	switch s {
	case SourceDefaults:
		return "defaults"
	case SourceFile:
		return "file"
	case SourceInvalid:
		return "invalid"
	default:
		return "unknown"
	}
}

// Loaded is the result of reading the configuration, with its provenance made
// explicit so callers can tell "no file yet" from "unreadable file".
type Loaded struct {
	Config *Config
	Source Source
	Path   string
	err    error
}

// Err returns the load error when Source is SourceInvalid.
func (l Loaded) Err() error { return l.err }

// Defaults reports whether the config is the built-in default rather than
// something read from disk.
func (l Loaded) Defaults() bool { return l.Source == SourceDefaults }

// SafeToSave reports whether writing this config would not overwrite data that
// failed to load.
func (l Loaded) SafeToSave() bool { return l.Source != SourceInvalid }

// LoadWithSource reads the configuration and reports where it came from. Unlike
// Load it distinguishes a missing file (defaults, safe to save) from a file it
// could not read or parse (SourceInvalid, not safe to overwrite).
func LoadWithSource() Loaded {
	p, err := path()
	if err != nil {
		return Loaded{Config: &Config{TVIP: defaultTVIP}, Source: SourceInvalid, err: err}
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Loaded{Config: &Config{TVIP: defaultTVIP}, Source: SourceDefaults, Path: p}
		}
		return Loaded{Config: &Config{TVIP: defaultTVIP}, Source: SourceInvalid, Path: p, err: err}
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return Loaded{Config: &Config{TVIP: defaultTVIP}, Source: SourceInvalid, Path: p, err: err}
	}
	if cfg.TVIP == "" {
		cfg.TVIP = defaultTVIP
	}
	return Loaded{Config: cfg, Source: SourceFile, Path: p}
}
