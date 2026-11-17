package player

import "strings"

// Kind is what a `tv play` argument refers to.
//
//	KindURL       a remote media URL; handed to the player unchanged
//	KindLocalFile a local path; served over HTTP first, then handed to the player
//
// The distinction decides whether `tv play` starts a file server, so it is
// derived once instead of being re-tested at each step.
type Kind int

const (
	KindURL Kind = iota
	KindLocalFile
)

// Target is one `tv play` argument resolved to its kind. For KindLocalFile,
// Path is absolute.
type Target struct {
	Kind Kind
	Path string
	URL  string
}

func (k Kind) String() string {
	switch k {
	case KindLocalFile:
		return "local-file"
	case KindURL:
		return "url"
	default:
		return "unknown"
	}
}

// IsLocal reports whether the target must be served before playback.
func (k Kind) IsLocal() bool { return k == KindLocalFile }

// ClassifyTarget decides whether arg names a local file or a remote URL. A
// directory is treated as a URL so it fails loudly at the player instead of
// being served as an unplayable path.
func ClassifyTarget(arg string, stat func(string) (bool, bool)) Target {
	isFile, ok := stat(arg)
	if ok && isFile {
		return Target{Kind: KindLocalFile, Path: arg}
	}
	return Target{Kind: KindURL, URL: arg}
}

// LooksLikeURL reports whether arg has a scheme prefix, used only for clearer
// error messages.
func LooksLikeURL(arg string) bool {
	return strings.Contains(arg, "://")
}
