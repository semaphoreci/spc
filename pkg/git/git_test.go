package git

import "testing"

const fallbackSha = "deadbeef"

func TestSplitThreeDotRangeDefaultsMissingSegments(t *testing.T) {
	t.Setenv("SEMAPHORE_GIT_SHA", fallbackSha)

	base, head := splitThreeDotRange("feature...")
	if base != "feature" {
		t.Fatalf("expected base to be feature, got %q", base)
	}
	if head != fallbackSha {
		t.Fatalf("expected head to fall back to current sha, got %q", head)
	}

	base, head = splitThreeDotRange("...feature")
	if base != fallbackSha {
		t.Fatalf("expected base to fall back to current sha, got %q", base)
	}
	if head != "feature" {
		t.Fatalf("expected head to be feature, got %q", head)
	}

	base, head = splitThreeDotRange("feature")
	if base != fallbackSha || head != fallbackSha {
		t.Fatalf("expected both sides to fall back to current sha, got %q and %q", base, head)
	}
}

func TestSplitThreeDotRangeFallsBackToHeadLiteral(t *testing.T) {
	t.Setenv("SEMAPHORE_GIT_SHA", "")

	base, head := splitThreeDotRange("...")
	if base != "HEAD" || head != "HEAD" {
		t.Fatalf("expected both sides to fall back to HEAD, got %q and %q", base, head)
	}
}
