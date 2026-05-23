package store

import (
	"testing"
	"time"
)

func TestKanbanMemoryStoreEnforcesProjectIsolationAndTransitions(t *testing.T) {
	state := NewMemoryStore()
	now := time.Now().UTC()
	left, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "alpha", Name: "Alpha", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create left project: %v", err)
	}
	right, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "beta", Name: "Beta", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create right project: %v", err)
	}
	if _, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "alpha", Name: "Alpha duplicate", Actor: KanbanActor{Type: "test", ID: "tester"}}, now); err == nil {
		t.Fatalf("duplicate project slug should be rejected")
	}
	leftTicket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: left.ID, Title: "Left ticket", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create left ticket: %v", err)
	}
	rightTicket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: right.ID, Title: "Right ticket", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create right ticket: %v", err)
	}
	rightCursor := state.LatestKanbanEventID(right.ID)
	if _, err := state.AddKanbanTicketLink(KanbanTicketLinkInput{FromTicketID: leftTicket.ID, ToTicketID: rightTicket.ID, Actor: KanbanActor{Type: "test", ID: "tester"}}, now); err == nil {
		t.Fatalf("expected cross-project link rejection")
	}
	if _, err := state.AddKanbanTicketLink(KanbanTicketLinkInput{FromTicketID: leftTicket.ID, ToTicketID: leftTicket.ID, Actor: KanbanActor{Type: "test", ID: "tester"}}, now); err == nil {
		t.Fatalf("expected self-link rejection")
	}
	secondLeftTicket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: left.ID, Title: "Second left ticket", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create second left ticket: %v", err)
	}
	if events := state.ListKanbanEvents(left.ID, rightCursor, 10); len(events) != 0 {
		t.Fatalf("cross-project cursor returned %d left event(s)", len(events))
	}
	firstLink, err := state.AddKanbanTicketLink(KanbanTicketLinkInput{FromTicketID: leftTicket.ID, ToTicketID: secondLeftTicket.ID, LinkType: "related", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}
	duplicateLink, err := state.AddKanbanTicketLink(KanbanTicketLinkInput{FromTicketID: leftTicket.ID, ToTicketID: secondLeftTicket.ID, LinkType: "related", Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err != nil {
		t.Fatalf("dedupe link: %v", err)
	}
	if duplicateLink.ID != firstLink.ID {
		t.Fatalf("duplicate link ID = %s, want existing %s", duplicateLink.ID, firstLink.ID)
	}
	emptyTitle := "   "
	if _, err := state.UpdateKanbanTicket(leftTicket.ID, KanbanTicketUpdateInput{Title: &emptyTitle, Actor: KanbanActor{Type: "test", ID: "tester"}}, now); err == nil {
		t.Fatalf("empty title update should be rejected")
	}
	if _, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: left.ID, Title: "Already done", Status: KanbanStatusDone, Actor: KanbanActor{Type: "test", ID: "tester"}}, now); err == nil {
		t.Fatalf("ticket create should not bypass status workflow")
	}
	inProgress := KanbanStatusInProgress
	updated, err := state.UpdateKanbanTicket(leftTicket.ID, KanbanTicketUpdateInput{Status: &inProgress, Actor: KanbanActor{Type: "test", ID: "tester"}}, now)
	if err == nil {
		t.Fatalf("triage -> in_progress should be rejected via todo first, got success %#v", updated)
	}
}

func TestKanbanStatusTransitionRules(t *testing.T) {
	if !KanbanStatusTransitionAllowed(KanbanStatusTriage, KanbanStatusTodo) {
		t.Fatalf("triage -> todo should be allowed")
	}
	if KanbanStatusTransitionAllowed(KanbanStatusTriage, KanbanStatusInProgress) {
		t.Fatalf("triage -> in_progress should require todo first")
	}
	if !KanbanStatusTransitionAllowed(KanbanStatusBlocked, KanbanStatusInProgress) {
		t.Fatalf("blocked -> in_progress should be allowed")
	}
	if !KanbanStatusTransitionAllowed(KanbanStatusInProgress, KanbanStatusTodo) {
		t.Fatalf("in_progress -> todo should be allowed")
	}
	if !KanbanStatusTransitionAllowed(KanbanStatusDone, KanbanStatusTodo) {
		t.Fatalf("done -> todo reopen should be allowed")
	}
	if KanbanStatusTransitionAllowed(KanbanStatusArchived, KanbanStatusDone) {
		t.Fatalf("archived should only restore to todo in V1")
	}
}

func TestKanbanMemoryStoreResolvesSlackProjectRoute(t *testing.T) {
	state := NewMemoryStore()
	now := time.Now().UTC()
	project, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, now)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(KanbanProjectSlackRouteInput{
		ProjectID: project.ID,
		TeamID:    "T123",
		ChannelID: "C123",
		Actor:     KanbanActor{Type: "test", ID: "tester"},
	}, now); err != nil {
		t.Fatalf("set slack route: %v", err)
	}
	resolved, ok := state.ResolveKanbanSlackProject("T123", "C123", "171000001.000200")
	if !ok || resolved.ID != project.ID {
		t.Fatalf("resolved project = %#v ok=%v, want %s", resolved, ok, project.ID)
	}
	if resolved, ok := state.ResolveKanbanSlackProject("T999", "C123", "171000001.000200"); ok {
		t.Fatalf("non-matching non-empty team resolved to %#v", resolved)
	}
	if latest := state.LatestKanbanEventID(project.ID); latest == "" || !state.KanbanEventExists(project.ID, latest) {
		t.Fatalf("latest event cursor did not resolve: %q", latest)
	}
	if events := state.ListKanbanEvents(project.ID, "missing-event", 10); len(events) != 0 {
		t.Fatalf("unknown cursor should return no historical events, got %d", len(events))
	}
	resolvedWithoutTeam, ok := state.ResolveKanbanSlackProject("", "C123", "171000001.000200")
	if !ok || resolvedWithoutTeam.ID != project.ID {
		t.Fatalf("unique team-scoped route without team = %#v ok=%v, want %s", resolvedWithoutTeam, ok, project.ID)
	}
	other, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "other", Name: "Other"}, now)
	if err != nil {
		t.Fatalf("create other project: %v", err)
	}
	if _, err := state.SetKanbanSlackProjectRoute(KanbanProjectSlackRouteInput{
		ProjectID: other.ID,
		TeamID:    "T999",
		ChannelID: "C123",
		Actor:     KanbanActor{Type: "test", ID: "tester"},
	}, now); err != nil {
		t.Fatalf("set second slack route: %v", err)
	}
	if resolved, ok := state.ResolveKanbanSlackProject("", "C123", "171000001.000200"); ok {
		t.Fatalf("ambiguous team-scoped route without team resolved to %#v", resolved)
	}
}

func TestKanbanMemoryStoreSlackSourceRefsRequireCompleteKeys(t *testing.T) {
	state := NewMemoryStore()
	now := time.Now().UTC()
	project, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, now)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	incompleteRef := KanbanTicketSourceRefInput{
		SourceType: "slack",
		ActionKind: "create_ticket",
		TeamID:     "T123",
		ChannelID:  "C123",
		ThreadTS:   "171000001.000100",
	}
	first, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "First", SourceRefs: []KanbanTicketSourceRefInput{incompleteRef}}, now)
	if err != nil {
		t.Fatalf("create first ticket: %v", err)
	}
	second, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Second", SourceRefs: []KanbanTicketSourceRefInput{incompleteRef}}, now)
	if err != nil {
		t.Fatalf("create second ticket: %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("incomplete Slack ref deduped tickets: %s", first.ID)
	}
	completeRef := incompleteRef
	completeRef.MessageTS = "171000001.000200"
	third, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Third", SourceRefs: []KanbanTicketSourceRefInput{completeRef}}, now)
	if err != nil {
		t.Fatalf("create third ticket: %v", err)
	}
	duplicate, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Duplicate", SourceRefs: []KanbanTicketSourceRefInput{completeRef}}, now)
	if err != nil {
		t.Fatalf("dedupe complete ref: %v", err)
	}
	if duplicate.ID != third.ID {
		t.Fatalf("complete Slack ref did not dedupe: got %s want %s", duplicate.ID, third.ID)
	}
	withoutTeam := completeRef
	withoutTeam.TeamID = ""
	mixedTeam, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Mixed team duplicate", SourceRefs: []KanbanTicketSourceRefInput{withoutTeam}}, now)
	if err != nil {
		t.Fatalf("dedupe empty-team variant: %v", err)
	}
	if mixedTeam.ID != third.ID {
		t.Fatalf("empty-team Slack ref variant did not dedupe: got %s want %s", mixedTeam.ID, third.ID)
	}
	secondMessageWithoutTeam := withoutTeam
	secondMessageWithoutTeam.MessageTS = "171000001.000300"
	noTeamTicket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "No team first", SourceRefs: []KanbanTicketSourceRefInput{secondMessageWithoutTeam}}, now)
	if err != nil {
		t.Fatalf("create no-team ticket: %v", err)
	}
	secondMessageWithTeam := secondMessageWithoutTeam
	secondMessageWithTeam.TeamID = "T123"
	teamVariant, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Team variant duplicate", SourceRefs: []KanbanTicketSourceRefInput{secondMessageWithTeam}}, now)
	if err != nil {
		t.Fatalf("dedupe team variant: %v", err)
	}
	if teamVariant.ID != noTeamTicket.ID {
		t.Fatalf("team Slack ref variant did not dedupe: got %s want %s", teamVariant.ID, noTeamTicket.ID)
	}
}

func TestKanbanDoneNoopPreservesCompletedAt(t *testing.T) {
	state := NewMemoryStore()
	now := time.Now().UTC()
	project, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, now)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	ticket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "New ticket"}, now)
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	todo := KanbanStatusTodo
	ticket, err = state.UpdateKanbanTicket(ticket.ID, KanbanTicketUpdateInput{Status: &todo}, now)
	if err != nil {
		t.Fatalf("move to todo: %v", err)
	}
	inProgress := KanbanStatusInProgress
	ticket, err = state.UpdateKanbanTicket(ticket.ID, KanbanTicketUpdateInput{Status: &inProgress}, now)
	if err != nil {
		t.Fatalf("move to in_progress: %v", err)
	}
	done := KanbanStatusDone
	ticket, err = state.UpdateKanbanTicket(ticket.ID, KanbanTicketUpdateInput{Status: &done}, now)
	if err != nil {
		t.Fatalf("move to done: %v", err)
	}
	if ticket.CompletedAt == nil {
		t.Fatalf("done ticket missing completed_at")
	}
	originalCompletedAt := *ticket.CompletedAt
	later := now.Add(5 * time.Minute)
	updated, err := state.UpdateKanbanTicket(ticket.ID, KanbanTicketUpdateInput{Status: &done}, later)
	if err != nil {
		t.Fatalf("noop done update: %v", err)
	}
	if updated.CompletedAt == nil || !updated.CompletedAt.Equal(originalCompletedAt) {
		t.Fatalf("completed_at changed on noop done update: got %v want %v", updated.CompletedAt, originalCompletedAt)
	}
}

func TestKanbanMemoryStoreSlackSourceRefOrphanDoesNotOverwriteIdempotency(t *testing.T) {
	state := NewMemoryStore()
	now := time.Now().UTC()
	project, err := state.CreateKanbanProject(KanbanProjectCreateInput{Slug: "platform", Name: "Platform"}, now)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	ref := KanbanTicketSourceRefInput{
		SourceType: "slack",
		ActionKind: "create_ticket",
		TeamID:     "T123",
		ChannelID:  "C123",
		ThreadTS:   "171000001.000100",
		MessageTS:  "171000001.000200",
	}
	ticket, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "First", SourceRefs: []KanbanTicketSourceRefInput{ref}}, now)
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	var sourceRefID string
	for id, item := range state.kanbanSourceRefs {
		if item.TicketID == ticket.ID {
			sourceRefID = id
			break
		}
	}
	if sourceRefID == "" {
		t.Fatalf("expected source ref for ticket %s", ticket.ID)
	}
	delete(state.kanbanTickets, ticket.ID)
	if _, err := state.CreateKanbanTicket(KanbanTicketCreateInput{ProjectID: project.ID, Title: "Second", SourceRefs: []KanbanTicketSourceRefInput{ref}}, now); err == nil {
		t.Fatalf("stale Slack source ref index should not be overwritten")
	}
	if got := state.kanbanSlackSourceRefByKey[KanbanSlackSourceRefKey(ref)]; got != sourceRefID {
		t.Fatalf("source ref index was overwritten: got %s want %s", got, sourceRefID)
	}
}
