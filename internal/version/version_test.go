package version

import "testing"

func TestStringOmitsDevCommit(t *testing.T) {
	origV, origC := Version, Commit
	t.Cleanup(func() {
		Version, Commit = origV, origC
	})

	Version = "0.2.0"
	Commit = ""
	if String() != "0.2.0" {
		t.Fatalf("empty commit: %q", String())
	}

	Commit = "dev"
	if String() != "0.2.0" {
		t.Fatalf("dev commit: %q", String())
	}

	Commit = "abc1234"
	if String() != "0.2.0 (abc1234)" {
		t.Fatalf("sha commit: %q", String())
	}
}
