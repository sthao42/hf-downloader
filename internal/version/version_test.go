package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Get() != "1.0.0" {
		t.Errorf("expected '1.0.0', got %q", Get())
	}

	info := Info()
	if info == "" {
		t.Errorf("expected non-empty info string")
	}

	bInfo := BuildInfo()
	if bInfo["version"] != "1.0.0" {
		t.Errorf("expected '1.0.0' in build info, got %q", bInfo["version"])
	}
}
