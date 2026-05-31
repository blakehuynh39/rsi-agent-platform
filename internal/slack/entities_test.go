package slack

import "testing"

func TestExtractEntityRefsCapturesSlackMentionsWithoutDupes(t *testing.T) {
	refs := ExtractEntityRefs("Hello <@U12345678|blake>, check <#C12345678|depin-backend> and #C12345678 plus @U12345678.")
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %#v", refs)
	}
	if refs[0].Kind != EntityUser || refs[0].ID != "U12345678" || refs[0].Label != "blake" {
		t.Fatalf("unexpected user ref %#v", refs[0])
	}
	if refs[1].Kind != EntityChannel || refs[1].ID != "C12345678" || refs[1].Label != "depin-backend" {
		t.Fatalf("unexpected channel ref %#v", refs[1])
	}
}
