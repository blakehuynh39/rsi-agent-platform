package dbread

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	pg_query "github.com/pganalyze/pg_query_go/v6"
	"google.golang.org/protobuf/encoding/protojson"
)

type SQLValidation struct {
	OK          bool              `json:"ok"`
	SQLSHA256   string            `json:"sql_sha256"`
	Preview     string            `json:"preview"`
	ErrorCode   string            `json:"error_code,omitempty"`
	Message     string            `json:"message,omitempty"`
	RepairHints map[string]string `json:"repair_hints,omitempty"`
}

func ValidateSQLSafety(sql string) SQLValidation {
	sql = strings.TrimSpace(sql)
	sum := sha256.Sum256([]byte(sql))
	result := SQLValidation{
		SQLSHA256: "sha256:" + hex.EncodeToString(sum[:]),
		Preview:   truncateSQL(sql, 1800),
	}
	if sql == "" {
		result.ErrorCode = "empty_sql"
		result.Message = "SQL is required"
		return result
	}
	stmts, err := pg_query.SplitWithParser(sql, true)
	if err != nil {
		result.ErrorCode = "parse_error"
		result.Message = err.Error()
		return result
	}
	if len(stmts) != 1 {
		result.ErrorCode = "multi_statement"
		result.Message = "Only one SQL statement is allowed"
		return result
	}
	tree, err := pg_query.Parse(sql)
	if err != nil {
		result.ErrorCode = "parse_error"
		result.Message = err.Error()
		return result
	}
	if len(tree.GetStmts()) != 1 {
		result.ErrorCode = "multi_statement"
		result.Message = "Only one SQL statement is allowed"
		return result
	}
	raw := tree.GetStmts()[0]
	stmt := raw.GetStmt()
	selectStmt := stmt.GetSelectStmt()
	if selectStmt == nil {
		result.ErrorCode = "not_select"
		result.Message = "Only SELECT or read-only WITH statements are allowed"
		return result
	}
	if selectStmt.GetIntoClause() != nil {
		result.ErrorCode = "select_into"
		result.Message = "SELECT INTO is not allowed"
		return result
	}
	if len(selectStmt.GetLockingClause()) > 0 {
		result.ErrorCode = "row_lock"
		result.Message = "SELECT FOR UPDATE/SHARE locking clauses are not allowed"
		return result
	}
	if err := validateReadOnlySelect(selectStmt); err != nil {
		result.ErrorCode = "unsafe_ast"
		result.Message = err.Error()
		return result
	}
	if err := validateASTSafety(tree); err != nil {
		code := "unsafe_ast"
		if strings.Contains(err.Error(), "function") {
			code = "unsafe_function"
		}
		if strings.Contains(err.Error(), "catalog") {
			code = "unsafe_catalog"
		}
		result.ErrorCode = code
		result.Message = err.Error()
		return result
	}
	result.OK = true
	return result
}

func validateASTSafety(tree *pg_query.ParseResult) error {
	raw, err := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(tree)
	if err != nil {
		return fmt.Errorf("failed to inspect SQL AST: %w", err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("failed to inspect SQL AST: %w", err)
	}
	return walkAST(document)
}

func walkAST(value any) error {
	switch node := value.(type) {
	case map[string]any:
		if raw, ok := node["FuncCall"]; ok {
			name := funcNameFromJSON(raw)
			if isUnsafeFunction(name) {
				return fmt.Errorf("SQL calls blocked function %q", name)
			}
		}
		if raw, ok := node["RangeVar"]; ok {
			schema, relation := rangeVarFromJSON(raw)
			if isBlockedCatalog(schema, relation) {
				return fmt.Errorf("SQL reads blocked catalog metadata %q", relationName(schema, relation))
			}
		}
		for _, child := range node {
			if err := walkAST(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range node {
			if err := walkAST(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func funcNameFromJSON(value any) string {
	node, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	parts, ok := node["funcname"].([]any)
	if !ok {
		return ""
	}
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		partNode, ok := part.(map[string]any)
		if !ok {
			continue
		}
		stringNode, ok := partNode["String"].(map[string]any)
		if !ok {
			continue
		}
		if value, ok := stringNode["sval"].(string); ok {
			names = append(names, strings.ToLower(value))
		}
	}
	return strings.Join(names, ".")
}

func rangeVarFromJSON(value any) (string, string) {
	node, ok := value.(map[string]any)
	if !ok {
		return "", ""
	}
	schema, _ := node["schemaname"].(string)
	relation, _ := node["relname"].(string)
	return strings.ToLower(schema), strings.ToLower(relation)
}

func isUnsafeFunction(name string) bool {
	parts := strings.Split(name, ".")
	base := name
	if len(parts) > 0 {
		base = parts[len(parts)-1]
	}
	base = strings.ToLower(strings.TrimSpace(base))
	switch base {
	case "pg_sleep", "set_config", "nextval", "postgres_fdw_handler", "lo_import", "lo_export":
		return true
	}
	return strings.HasPrefix(base, "dblink") ||
		strings.HasPrefix(base, "pg_advisory_") ||
		strings.HasPrefix(base, "http_")
}

func isBlockedCatalog(schema, relation string) bool {
	switch relation {
	case "pg_authid", "pg_shadow", "pg_user_mapping", "pg_auth_members":
		return true
	}
	if schema == "information_schema" {
		return relation == "enabled_roles" ||
			relation == "administrable_role_authorizations" ||
			strings.HasPrefix(relation, "role_")
	}
	return false
}

func relationName(schema, relation string) string {
	if schema == "" {
		return relation
	}
	return schema + "." + relation
}

func validateReadOnlySelect(stmt *pg_query.SelectStmt) error {
	if stmt == nil {
		return fmt.Errorf("missing select statement")
	}
	if stmt.GetIntoClause() != nil {
		return fmt.Errorf("SELECT INTO is not allowed")
	}
	if len(stmt.GetLockingClause()) > 0 {
		return fmt.Errorf("row locking clauses are not allowed")
	}
	if with := stmt.GetWithClause(); with != nil {
		for _, cteNode := range with.GetCtes() {
			cte := cteNode.GetCommonTableExpr()
			if cte == nil {
				return fmt.Errorf("invalid CTE")
			}
			cteSelect := cte.GetCtequery().GetSelectStmt()
			if cteSelect == nil {
				return fmt.Errorf("data-modifying CTE %q is not allowed", cte.GetCtename())
			}
			if err := validateReadOnlySelect(cteSelect); err != nil {
				return err
			}
		}
	}
	if stmt.GetLarg() != nil {
		if err := validateReadOnlySelect(stmt.GetLarg()); err != nil {
			return err
		}
	}
	if stmt.GetRarg() != nil {
		if err := validateReadOnlySelect(stmt.GetRarg()); err != nil {
			return err
		}
	}
	return nil
}

func truncateSQL(sql string, max int) string {
	if len(sql) <= max {
		return sql
	}
	if max < 16 {
		for i := max; i > 0; i-- {
			if utf8.ValidString(sql[:i]) {
				return sql[:i]
			}
		}
		return ""
	}
	cutPoint := max - 16
	for i := cutPoint; i >= 0 && i > cutPoint-4; i-- {
		if utf8.ValidString(sql[:i]) {
			return sql[:i] + "\n-- truncated --"
		}
	}
	if cutPoint > 0 {
		return sql[:cutPoint] + "\n-- truncated --"
	}
	return ""
}
