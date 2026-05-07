package control

import "testing"

func TestDBReadIdempotencyKeyEscapesFieldBoundaries(t *testing.T) {
	left := dbReadIdempotencyKey(dbReadQueryRequest{
		ConversationID: "a:b",
		ThreadTS:       "",
		Target:         "depin-stage",
		Requester:      "hermes",
		Purpose:        "query",
	}, "sha256:abc")
	right := dbReadIdempotencyKey(dbReadQueryRequest{
		ConversationID: "a",
		ThreadTS:       "b",
		Target:         "depin-stage",
		Requester:      "hermes",
		Purpose:        "query",
	}, "sha256:abc")
	if left == right {
		t.Fatalf("expected distinct idempotency keys for distinct field boundaries")
	}
}
