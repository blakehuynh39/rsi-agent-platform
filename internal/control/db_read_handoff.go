package control

import (
	"time"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type dbReadResponseScope struct {
	ConversationID string
	WorkflowID     string
	TraceID        string
	ChannelID      string
	ThreadTS       string
	NotBefore      time.Time
}

func latestDBReadResponseOwner(store storepkg.Store, scope dbReadResponseScope) (storepkg.DBReadRequest, bool) {
	var latest storepkg.DBReadRequest
	found := false
	for _, request := range store.ListDBReadRequestsByScope(scope.ConversationID, scope.WorkflowID, scope.TraceID, scope.ChannelID, scope.ThreadTS, scope.NotBefore) {
		if !dbReadStateOwnsResponse(request.State) {
			continue
		}
		if !found || request.CreatedAt.After(latest.CreatedAt) {
			latest = request
			found = true
		}
	}
	return latest, found
}

func dbReadStateOwnsResponse(state storepkg.DBReadState) bool {
	return state != storepkg.DBReadStateValidationFailed
}
