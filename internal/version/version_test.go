package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Get() != "1.1.1" {
		t.Errorf("expected '1.1.1', got %q", Get())
	}

	info := Info()
	if info == "" {
		t.Errorf("expected non-empty info string")
	}

	bInfo := BuildInfo()
	if bInfo["version"] != "1.1.1" {
		t.Errorf("expected '1.1.1' in build info, got %q", bInfo["version"])
	}
}
