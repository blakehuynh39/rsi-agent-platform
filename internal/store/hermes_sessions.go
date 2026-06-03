package store

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

type HermesSessionListPage struct {
	Items                 []HermesSessionListItem
	Total                 int
	LedgerActivityByTrace map[string]HermesLedgerActivity
}

type HermesSessionListItem struct {
	ID                string
	Type              string
	Source            string
	Model             string
	RawTitle          string
	SessionTitle      string
	TitleTraceID      string
	TriggerTitle      string
	Preview           string
	ActiveCaseSummary string
	StartedAt         time.Time
	EndedAt           *time.Time
	LastActive        time.Time
	ConversationID    string
	TraceID           string
	ParentSessionID   string
	CaseID            string
	TriggerEventID    string
	ThreadKey         string
	WorkflowKind      string
	Status            string
	LastVerdict       string
	MessageCount      int
	ToolCallCount     int
	TraceCount        int
	OpenTraceCount    int
	ProposalCount     int
}

type HermesLiveLedgerEvent struct {
	TraceID        string
	ConversationID string
	Kind           string
	Status         string
	RecordedAt     time.Time
}

type HermesLedgerActivity struct {
	TraceID        string
	ConversationID string
	RecordedAt     time.Time
}

type hermesLedgerTraceCandidate struct {
	TraceID    string
	RecordedAt time.Time
}

func (p *PostgresStore) ListHermesSessionsPage(limit int, offset int) HermesSessionListPage {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	window := limit + offset
	if window < limit {
		window = limit
	}

	// A top-N trace by max(trace base activity, ledger activity) must be in the
	// top-N set for at least one side, so the page can be composed from bounded
	// trace-summary and ledger-recency candidates instead of a global ledger CTE.
	ledgerTraceCandidates := p.recentHermesLedgerTraceCandidates(window)
	ledgerCandidateJSON := hermesLedgerTraceCandidatesJSON(ledgerTraceCandidates)

	rows, err := p.db.Query(`
			with ledger_trace_candidates as (
				select trace_id, recorded_at
				from jsonb_to_recordset($4::jsonb) as c(trace_id text, recorded_at timestamptz)
				where trace_id <> '' and recorded_at is not null
			),
			base_trace_candidates as (
				select trace_id
				from trace_summary
				order by
					(case
						when ended_at is null or ended_at <= timestamptz '0001-01-02 00:00:00+00'
						then started_at
						else ended_at
					end) desc,
					trace_id asc
				limit $3
			),
		trace_candidate_ids as (
			select trace_id from base_trace_candidates
			union
			select trace_id from ledger_trace_candidates
		),
		trace_ranked as (
			select
				t.trace_id as id,
				'trace'::text as session_type,
				'trace'::text as source,
			coalesce(nullif(t.workflow_kind, ''), 'rsi-trace') as model,
			coalesce(nullif(e.normalized_problem_statement, ''), nullif(t.workflow_kind, ''), t.trace_id) as raw_title,
			coalesce(e.normalized_problem_statement, '') as trigger_title,
			t.trace_id as title_trace_id,
				t.started_at as started_at,
				case when t.ended_at is null or t.ended_at <= timestamptz '0001-01-02 00:00:00+00' then null else t.ended_at end as ended_at,
				greatest(
					case when t.ended_at is null or t.ended_at <= timestamptz '0001-01-02 00:00:00+00' then t.started_at else t.ended_at end,
					coalesce(l.recorded_at, timestamptz '0001-01-01 00:00:00+00')
				) as last_active,
				coalesce(t.conversation_id, '') as conversation_id,
				t.trace_id as trace_id,
				coalesce(t.conversation_id, '') as parent_session_id,
			coalesce(t.case_id, '') as case_id,
			coalesce(t.trigger_event_id, '') as trigger_event_id,
			t.thread_key as thread_key,
			t.workflow_kind as workflow_kind,
			t.status as status,
			coalesce(t.last_verdict, '') as last_verdict,
			(t.event_count + t.reasoning_step_count + t.tool_call_count)::int as message_count,
			t.tool_call_count::int as tool_call_count,
			0::int as trace_count,
			0::int as open_trace_count,
				0::int as proposal_count,
				''::text as active_case_summary
			from trace_candidate_ids ci
			join trace_summary t on t.trace_id = ci.trace_id
			left join event_envelope e on e.id = t.trigger_event_id
			left join lateral (
				select le.recorded_at
				from execution_ledger_event le
				where le.trace_id = t.trace_id
				order by le.recorded_at desc, le.execution_id desc, le.seq desc, le.id desc
				limit 1
			) l on true
			order by last_active desc, t.trace_id asc
			limit $3
		),
		recent_trace_parents as (
			select conversation_id, max(last_active) as latest_trace_at
			from trace_ranked
			where conversation_id <> ''
			group by conversation_id
		),
		conversation_candidate_ids as (
			select id
			from (
				select c.id
				from conversation c
				order by c.updated_at desc, c.id asc
				limit $3
			) recent_conversations
			union
			select conversation_id as id
			from recent_trace_parents
		),
		conversation_candidates as (
			select
				c.id,
				'conversation'::text as session_type,
				coalesce(nullif(c.source, ''), 'conversation') as source,
				'rsi-platform'::text as model,
				coalesce(nullif(c.title, ''), nullif(c.external_key, ''), c.id) as raw_title,
				''::text as trigger_title,
				null::text as title_trace_id,
				c.created_at as started_at,
				null::timestamptz as ended_at,
				greatest(c.updated_at, coalesce(rp.latest_trace_at, c.updated_at)) as last_active,
				c.id as conversation_id,
				''::text as trace_id,
				''::text as parent_session_id,
				''::text as case_id,
				''::text as trigger_event_id,
				c.external_key as thread_key,
				''::text as workflow_kind,
				c.status as status,
				''::text as last_verdict,
				0::int as message_count,
				0::int as tool_call_count,
				0::int as trace_count,
				0::int as open_trace_count,
				0::int as proposal_count,
				coalesce(ac.summary, '') as active_case_summary
			from conversation_candidate_ids ci
			join conversation c on c.id = ci.id
			left join recent_trace_parents rp on rp.conversation_id = c.id
			left join case_record ac on ac.id = c.active_case_id
			order by last_active desc, c.id asc
			limit $3
		),
		trace_candidates as (
			select * from trace_ranked
		),
		session_rows as (
			select * from conversation_candidates
			union all
			select * from trace_candidates
		),
		totals as (
			select ((select count(*) from conversation) + (select count(*) from trace_summary))::int as total
		)
	select
		id, session_type, source, model, raw_title, trigger_title, title_trace_id,
		started_at, ended_at, last_active, conversation_id, trace_id, parent_session_id,
		case_id, trigger_event_id, thread_key, workflow_kind, status, last_verdict, message_count,
		tool_call_count, trace_count, open_trace_count, proposal_count, active_case_summary,
		totals.total
	from session_rows
		cross join totals
		order by last_active desc, id asc
		limit $1 offset $2
		`, limit, offset, window, ledgerCandidateJSON)
	if err != nil {
		return HermesSessionListPage{}
	}
	defer rows.Close()

	items := []HermesSessionListItem{}
	total := 0
	titleTraceIDs := []string{}
	conversationIDs := []string{}
	traceIDs := []string{}
	for rows.Next() {
		item, titleTraceID, rowTotal, err := scanHermesSessionListItem(rows)
		if err != nil {
			return HermesSessionListPage{}
		}
		total = rowTotal
		items = append(items, item)
		if titleTraceID != "" {
			titleTraceIDs = append(titleTraceIDs, titleTraceID)
		}
		if item.Type == "conversation" && item.ConversationID != "" {
			conversationIDs = append(conversationIDs, item.ConversationID)
		}
		if item.Type == "trace" && item.TraceID != "" {
			traceIDs = append(traceIDs, item.TraceID)
		}
	}
	if err := rows.Err(); err != nil {
		return HermesSessionListPage{}
	}
	if len(items) == 0 {
		return HermesSessionListPage{
			Items:                 []HermesSessionListItem{},
			Total:                 p.hermesSessionTotal(),
			LedgerActivityByTrace: map[string]HermesLedgerActivity{},
		}
	}

	firstTraces := p.firstTraceByConversation(conversationIDs)
	for _, traceID := range firstTraces {
		titleTraceIDs = append(titleTraceIDs, traceID)
	}
	titles := p.sessionTitlesByTraceID(titleTraceIDs)
	entryCounts := p.conversationEntryCounts(conversationIDs)
	proposalCounts := p.proposalCounts(conversationIDs)
	traceCounts := p.traceCounts(conversationIDs)
	previews := p.latestConversationEntryBodies(conversationIDs)
	canonicalToolIndexes := p.hermesCanonicalToolIndexes(traceIDs)
	projectedLedgerToolCounts := p.hermesProjectedLedgerToolCallCounts(traceIDs, canonicalToolIndexes)
	latestLedgerActivity := p.LatestHermesLedgerActivityByTraceID(traceIDs)
	for i := range items {
		if title := titles[items[i].TraceID]; title != "" {
			items[i].SessionTitle = title
		}
		if items[i].Type == "conversation" {
			items[i].TitleTraceID = firstTraces[items[i].ConversationID]
			if title := titles[items[i].TitleTraceID]; title != "" {
				items[i].SessionTitle = title
			}
			items[i].MessageCount = entryCounts[items[i].ConversationID]
			items[i].ProposalCount = proposalCounts[items[i].ConversationID]
			if counts, ok := traceCounts[items[i].ConversationID]; ok {
				items[i].TraceCount = counts.total
				items[i].OpenTraceCount = counts.open
			}
			items[i].Preview = previews[items[i].ConversationID]
		} else if items[i].Type == "trace" {
			if latest := latestLedgerActivity[items[i].TraceID]; !latest.RecordedAt.IsZero() {
				if items[i].EndedAt != nil && latest.RecordedAt.After(*items[i].EndedAt) {
					items[i].EndedAt = &latest.RecordedAt
				}
				if items[i].ConversationID == "" && latest.ConversationID != "" {
					items[i].ConversationID = latest.ConversationID
					items[i].ParentSessionID = latest.ConversationID
				}
			}
			canonicalCount := items[i].ToolCallCount
			if index := canonicalToolIndexes[items[i].TraceID]; index.count > canonicalCount {
				canonicalCount = index.count
			}
			toolCallCount := canonicalCount + projectedLedgerToolCounts[items[i].TraceID]
			items[i].MessageCount += toolCallCount - items[i].ToolCallCount
			items[i].ToolCallCount = toolCallCount
		}
	}
	return HermesSessionListPage{
		Items:                 items,
		Total:                 total,
		LedgerActivityByTrace: latestLedgerActivity,
	}
}

func (p *PostgresStore) hermesSessionTotal() int {
	var total int
	if err := p.db.QueryRow(`select (select count(*) from conversation) + (select count(*) from trace_summary)`).Scan(&total); err != nil {
		return 0
	}
	return total
}

func (p *PostgresStore) recentHermesLedgerTraceCandidates(limit int) []hermesLedgerTraceCandidate {
	if limit <= 0 {
		return nil
	}
	pageSize := limit * 8
	if pageSize < 512 {
		pageSize = 512
	}
	if pageSize > 5000 {
		pageSize = 5000
	}

	seen := map[string]bool{}
	out := make([]hermesLedgerTraceCandidate, 0, limit)
	var cursor hermesLedgerTraceCursor
	for len(out) < limit {
		items, nextCursor, ok := p.scanRecentHermesLedgerTraceCandidatesPage(pageSize, cursor)
		if !ok || len(items) == 0 {
			break
		}
		for _, item := range items {
			traceID := strings.TrimSpace(item.traceID)
			if traceID == "" || seen[traceID] {
				continue
			}
			seen[traceID] = true
			out = append(out, hermesLedgerTraceCandidate{
				TraceID:    traceID,
				RecordedAt: item.recordedAt,
			})
			if len(out) >= limit {
				return out
			}
		}
		if len(items) < pageSize {
			break
		}
		cursor = nextCursor
	}
	return out
}

type hermesLedgerTraceCursor struct {
	valid       bool
	traceID     string
	executionID string
	id          string
	seq         int
	recordedAt  time.Time
}

func (p *PostgresStore) scanRecentHermesLedgerTraceCandidatesPage(limit int, cursor hermesLedgerTraceCursor) ([]hermesLedgerTraceCursor, hermesLedgerTraceCursor, bool) {
	if limit <= 0 {
		return nil, hermesLedgerTraceCursor{}, true
	}
	rows, err := p.db.Query(`
		select trace_id, execution_id, seq, id, recorded_at
		from execution_ledger_event
		where trace_id <> ''
			and (
				$1 = false
				or recorded_at < $2
				or (recorded_at = $2 and trace_id > $3)
				or (recorded_at = $2 and trace_id = $3 and execution_id < $4)
				or (recorded_at = $2 and trace_id = $3 and execution_id = $4 and seq < $5)
				or (recorded_at = $2 and trace_id = $3 and execution_id = $4 and seq = $5 and id < $6)
			)
		order by recorded_at desc, trace_id asc, execution_id desc, seq desc, id desc
		limit $7
	`, cursor.valid, cursor.recordedAt, cursor.traceID, cursor.executionID, cursor.seq, cursor.id, limit)
	if err != nil {
		return nil, hermesLedgerTraceCursor{}, false
	}
	defer rows.Close()

	items := []hermesLedgerTraceCursor{}
	nextCursor := hermesLedgerTraceCursor{}
	for rows.Next() {
		var item hermesLedgerTraceCursor
		if err := rows.Scan(&item.traceID, &item.executionID, &item.seq, &item.id, &item.recordedAt); err != nil {
			return nil, hermesLedgerTraceCursor{}, false
		}
		item.valid = true
		item.traceID = strings.TrimSpace(item.traceID)
		item.executionID = strings.TrimSpace(item.executionID)
		item.id = strings.TrimSpace(item.id)
		items = append(items, item)
		nextCursor = item
	}
	if err := rows.Err(); err != nil {
		return nil, hermesLedgerTraceCursor{}, false
	}
	return items, nextCursor, true
}

func hermesLedgerTraceCandidatesJSON(items []hermesLedgerTraceCandidate) string {
	if len(items) == 0 {
		return "[]"
	}
	type candidateRow struct {
		TraceID    string    `json:"trace_id"`
		RecordedAt time.Time `json:"recorded_at"`
	}
	rows := make([]candidateRow, 0, len(items))
	for _, item := range items {
		traceID := strings.TrimSpace(item.TraceID)
		if traceID == "" || item.RecordedAt.IsZero() {
			continue
		}
		rows = append(rows, candidateRow{
			TraceID:    traceID,
			RecordedAt: item.RecordedAt,
		})
	}
	if len(rows) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func scanHermesSessionListItem(scanner interface{ Scan(dest ...any) error }) (HermesSessionListItem, string, int, error) {
	var item HermesSessionListItem
	var titleTraceID sql.NullString
	var endedAt sql.NullTime
	total := 0
	err := scanner.Scan(
		&item.ID,
		&item.Type,
		&item.Source,
		&item.Model,
		&item.RawTitle,
		&item.TriggerTitle,
		&titleTraceID,
		&item.StartedAt,
		&endedAt,
		&item.LastActive,
		&item.ConversationID,
		&item.TraceID,
		&item.ParentSessionID,
		&item.CaseID,
		&item.TriggerEventID,
		&item.ThreadKey,
		&item.WorkflowKind,
		&item.Status,
		&item.LastVerdict,
		&item.MessageCount,
		&item.ToolCallCount,
		&item.TraceCount,
		&item.OpenTraceCount,
		&item.ProposalCount,
		&item.ActiveCaseSummary,
		&total,
	)
	if err != nil {
		return HermesSessionListItem{}, "", 0, err
	}
	if endedAt.Valid {
		item.EndedAt = &endedAt.Time
	}
	item.TitleTraceID = titleTraceID.String
	return item, titleTraceID.String, total, nil
}

func (p *PostgresStore) sessionTitlesByTraceID(traceIDs []string) map[string]string {
	traceIDs = compactStrings(traceIDs)
	out := map[string]string{}
	if len(traceIDs) == 0 {
		return out
	}
	query := `select distinct on (trace_id) trace_id, coalesce(nullif(summary, ''), nullif(decision, '')) as title
		from reasoning_step
		where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `)
			and lower(step_type) = 'session_title'
			and coalesce(nullif(summary, ''), nullif(decision, '')) is not null
		order by trace_id asc, created_at desc, id desc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var traceID string
		var title string
		if err := rows.Scan(&traceID, &title); err != nil {
			return map[string]string{}
		}
		out[strings.TrimSpace(traceID)] = title
	}
	if err := rows.Err(); err != nil {
		return map[string]string{}
	}
	return out
}

func (p *PostgresStore) firstTraceByConversation(conversationIDs []string) map[string]string {
	conversationIDs = compactStrings(conversationIDs)
	out := map[string]string{}
	if len(conversationIDs) == 0 {
		return out
	}
	query := `select distinct on (conversation_id) conversation_id, trace_id
		from trace_summary
		where conversation_id in (` + sqlPlaceholders(len(conversationIDs), 1) + `)
		order by conversation_id asc, started_at asc, trace_id asc`
	rows, err := p.db.Query(query, stringsToAny(conversationIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID string
		var traceID string
		if err := rows.Scan(&conversationID, &traceID); err != nil {
			return map[string]string{}
		}
		out[conversationID] = traceID
	}
	if err := rows.Err(); err != nil {
		return map[string]string{}
	}
	return out
}

func (p *PostgresStore) conversationEntryCounts(conversationIDs []string) map[string]int {
	conversationIDs = compactStrings(conversationIDs)
	out := map[string]int{}
	if len(conversationIDs) == 0 {
		return out
	}
	query := `select conversation_id, count(*)::int from conversation_entry where conversation_id in (` + sqlPlaceholders(len(conversationIDs), 1) + `) group by conversation_id`
	rows, err := p.db.Query(query, stringsToAny(conversationIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID string
		var count int
		if err := rows.Scan(&conversationID, &count); err != nil {
			return map[string]int{}
		}
		out[conversationID] = count
	}
	if err := rows.Err(); err != nil {
		return map[string]int{}
	}
	return out
}

func (p *PostgresStore) proposalCounts(conversationIDs []string) map[string]int {
	conversationIDs = compactStrings(conversationIDs)
	out := map[string]int{}
	if len(conversationIDs) == 0 {
		return out
	}
	query := `select conversation_id, count(*)::int from proposal where conversation_id in (` + sqlPlaceholders(len(conversationIDs), 1) + `) group by conversation_id`
	rows, err := p.db.Query(query, stringsToAny(conversationIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID string
		var count int
		if err := rows.Scan(&conversationID, &count); err != nil {
			return map[string]int{}
		}
		out[conversationID] = count
	}
	if err := rows.Err(); err != nil {
		return map[string]int{}
	}
	return out
}

type traceCountPair struct {
	total int
	open  int
}

func (p *PostgresStore) traceCounts(conversationIDs []string) map[string]traceCountPair {
	conversationIDs = compactStrings(conversationIDs)
	out := map[string]traceCountPair{}
	if len(conversationIDs) == 0 {
		return out
	}
	query := `select conversation_id, count(*)::int, (count(*) filter (where status in ('queued','running','needs-human','in-review','replayed')))::int from trace_summary where conversation_id in (` + sqlPlaceholders(len(conversationIDs), 1) + `) group by conversation_id`
	rows, err := p.db.Query(query, stringsToAny(conversationIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID string
		var counts traceCountPair
		if err := rows.Scan(&conversationID, &counts.total, &counts.open); err != nil {
			return map[string]traceCountPair{}
		}
		out[conversationID] = counts
	}
	if err := rows.Err(); err != nil {
		return map[string]traceCountPair{}
	}
	return out
}

func (p *PostgresStore) latestConversationEntryBodies(conversationIDs []string) map[string]string {
	conversationIDs = compactStrings(conversationIDs)
	out := map[string]string{}
	if len(conversationIDs) == 0 {
		return out
	}
	query := `select distinct on (conversation_id) conversation_id, body
		from conversation_entry
		where conversation_id in (` + sqlPlaceholders(len(conversationIDs), 1) + `)
			and btrim(body) <> ''
		order by conversation_id asc, created_at desc, id desc`
	rows, err := p.db.Query(query, stringsToAny(conversationIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID string
		var body string
		if err := rows.Scan(&conversationID, &body); err != nil {
			return map[string]string{}
		}
		out[conversationID] = body
	}
	if err := rows.Err(); err != nil {
		return map[string]string{}
	}
	return out
}

func (p *PostgresStore) LatestHermesLedgerActivityByTraceID(traceIDs []string) map[string]HermesLedgerActivity {
	traceIDs = compactStrings(traceIDs)
	out := map[string]HermesLedgerActivity{}
	if len(traceIDs) == 0 {
		return out
	}
	query := `select distinct on (e.trace_id)
			e.trace_id,
			coalesce(t.conversation_id, '') as conversation_id,
			e.recorded_at
		from execution_ledger_event e
		left join trace_summary t on t.trace_id = e.trace_id
		where e.trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `)
		order by e.trace_id asc, e.recorded_at desc, e.execution_id desc, e.seq desc, e.id desc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var item HermesLedgerActivity
		if err := rows.Scan(&item.TraceID, &item.ConversationID, &item.RecordedAt); err != nil {
			return map[string]HermesLedgerActivity{}
		}
		item.TraceID = strings.TrimSpace(item.TraceID)
		item.ConversationID = strings.TrimSpace(item.ConversationID)
		if item.TraceID != "" {
			out[item.TraceID] = item
		}
	}
	if err := rows.Err(); err != nil {
		return map[string]HermesLedgerActivity{}
	}
	return out
}

type hermesCanonicalToolIndex struct {
	count int
	ids   map[string]bool
	names map[string]bool
}

type hermesLedgerToolCountEvent struct {
	traceID  string
	id       string
	kind     string
	status   string
	seq      int
	name     string
	stableID string
}

type hermesLedgerToolCountState struct {
	byKey      map[string]bool
	lastByName map[string]string
}

func (p *PostgresStore) hermesCanonicalToolIndexes(traceIDs []string) map[string]hermesCanonicalToolIndex {
	traceIDs = compactStrings(traceIDs)
	out := map[string]hermesCanonicalToolIndex{}
	if len(traceIDs) == 0 {
		return out
	}
	query := `select trace_id, id, tool_call_id, tool_name
		from tool_call_record
		where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `)
		order by trace_id asc, created_at asc, id asc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var traceID string
		var id string
		var toolCallID string
		var toolName string
		if err := rows.Scan(&traceID, &id, &toolCallID, &toolName); err != nil {
			return map[string]hermesCanonicalToolIndex{}
		}
		traceID = strings.TrimSpace(traceID)
		if traceID == "" {
			continue
		}
		index := out[traceID]
		if index.ids == nil {
			index.ids = map[string]bool{}
		}
		if index.names == nil {
			index.names = map[string]bool{}
		}
		index.count++
		for _, value := range []string{id, toolCallID} {
			if value = strings.TrimSpace(value); value != "" {
				index.ids[value] = true
			}
		}
		if toolName = strings.TrimSpace(toolName); toolName != "" {
			index.names[toolName] = true
		}
		out[traceID] = index
	}
	if err := rows.Err(); err != nil {
		return map[string]hermesCanonicalToolIndex{}
	}
	return out
}

func (p *PostgresStore) hermesProjectedLedgerToolCallCounts(traceIDs []string, canonical map[string]hermesCanonicalToolIndex) map[string]int {
	traceIDs = compactStrings(traceIDs)
	if len(traceIDs) == 0 {
		return map[string]int{}
	}
	query := `select
			trace_id,
			id,
			kind,
			status,
			seq,
			coalesce(nullif(payload->>'tool_name', ''), nullif(payload->>'name', ''), nullif(payload->>'tool', ''), nullif(payload->>'transport_tool_name', '')) as tool_name,
			coalesce(nullif(payload->>'tool_call_id', ''), nullif(payload->>'tool_id', ''), nullif(payload->>'call_id', ''), nullif(payload->>'id', '')) as stable_id
		from execution_ledger_event
		where trace_id in (` + sqlPlaceholders(len(traceIDs), 1) + `)
			and (lower(kind) like '%tool%' or btrim(coalesce(payload->>'tool_name', '')) <> '')
		order by trace_id asc, recorded_at asc, seq asc, id asc`
	rows, err := p.db.Query(query, stringsToAny(traceIDs)...)
	if err != nil {
		return map[string]int{}
	}
	defer rows.Close()
	events := []hermesLedgerToolCountEvent{}
	for rows.Next() {
		var item hermesLedgerToolCountEvent
		var name sql.NullString
		var stableID sql.NullString
		if err := rows.Scan(&item.traceID, &item.id, &item.kind, &item.status, &item.seq, &name, &stableID); err != nil {
			return map[string]int{}
		}
		item.traceID = strings.TrimSpace(item.traceID)
		item.id = strings.TrimSpace(item.id)
		item.kind = strings.TrimSpace(item.kind)
		item.status = strings.TrimSpace(item.status)
		item.name = trimmedNullString(name)
		item.stableID = trimmedNullString(stableID)
		events = append(events, item)
	}
	if err := rows.Err(); err != nil {
		return map[string]int{}
	}
	return countProjectedLedgerToolCallsByTrace(events, canonical)
}

func countProjectedLedgerToolCallsByTrace(events []hermesLedgerToolCountEvent, canonical map[string]hermesCanonicalToolIndex) map[string]int {
	out := map[string]int{}
	states := map[string]*hermesLedgerToolCountState{}
	for _, item := range events {
		traceID := strings.TrimSpace(item.traceID)
		name := strings.TrimSpace(item.name)
		if traceID == "" || name == "" {
			continue
		}
		stableID := strings.TrimSpace(item.stableID)
		index := canonical[traceID]
		if stableID != "" && index.ids[stableID] {
			continue
		}
		if stableID == "" && index.names[name] {
			continue
		}
		state := states[traceID]
		if state == nil {
			state = &hermesLedgerToolCountState{
				byKey:      map[string]bool{},
				lastByName: map[string]string{},
			}
			states[traceID] = state
		}
		key := stableID
		if key == "" {
			if !ledgerToolCountEventIsStart(item) {
				key = state.lastByName[name]
			}
			if key == "" {
				key = firstNonEmpty(strings.TrimSpace(item.id), name)
			}
		}
		if key == "" {
			continue
		}
		if !state.byKey[key] {
			state.byKey[key] = true
			out[traceID]++
		}
		state.lastByName[name] = key
	}
	return out
}

func (p *PostgresStore) ListRecentHermesLiveLedgerEvents(since time.Time) []HermesLiveLedgerEvent {
	rows, err := p.db.Query(`
		select
			e.trace_id,
			coalesce(t.conversation_id, '') as conversation_id,
			e.kind,
			e.status,
			e.recorded_at
		from execution_ledger_event e
		left join trace_summary t on t.trace_id = e.trace_id
		where e.trace_id <> ''
			and e.recorded_at >= $1
		order by e.trace_id asc, e.recorded_at asc, e.execution_id asc, e.seq asc, e.id asc
	`, since)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []HermesLiveLedgerEvent{}
	for rows.Next() {
		var item HermesLiveLedgerEvent
		if err := rows.Scan(&item.TraceID, &item.ConversationID, &item.Kind, &item.Status, &item.RecordedAt); err != nil {
			return nil
		}
		item.TraceID = strings.TrimSpace(item.TraceID)
		item.ConversationID = strings.TrimSpace(item.ConversationID)
		item.Kind = strings.TrimSpace(item.Kind)
		item.Status = strings.TrimSpace(item.Status)
		if item.TraceID != "" {
			out = append(out, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}

func ledgerToolCountEventIsStart(item hermesLedgerToolCountEvent) bool {
	kind := strings.ToLower(strings.TrimSpace(item.kind))
	status := strings.ToLower(strings.TrimSpace(item.status))
	if strings.Contains(kind, "progress") {
		return false
	}
	return strings.Contains(status, "start") ||
		strings.Contains(status, "running") ||
		strings.Contains(kind, "start")
}

func trimmedNullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}
