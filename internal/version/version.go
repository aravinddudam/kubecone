package version

// Version is the release number. Override at build time with -ldflags.
var Version = "0.2.0"

// Commit is the git SHA when set via -ldflags. Empty in source builds.
var Commit = ""

// String is what `kubecone version` and --version print.
// A placeholder "dev" commit is omitted so the banner is just the version.
func String() string {
	if Commit == "" || Commit == "dev" {
		return Version
	}
	return Version + " (" + Commit + ")"
}
