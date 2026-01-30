package db

import (
	"strings"
	"sync"
)

// reservedKeywords 存储各数据库的保留关键字集合
var reservedKeywords = struct {
	sync.RWMutex
	data map[string]map[string]bool // clientType -> keywords map
}{
	data: map[string]map[string]bool{
		MySQL:      buildMySQLKeywords(),
		Postgres:   buildPostgresKeywords(),
		Sqlite:     buildSQLiteKeywords(),
		ClickHouse: buildClickHouseKeywords(),
		TiDB:       nil, // 复用 MySQL 的
		GaussDB:    nil, // 复用 Postgres 的
	},
}

// quoteCache 缓存字段名是否需要转义的判断结果
var quoteCache = struct {
	sync.RWMutex
	data map[string]map[string]string // clientType -> fieldName -> quotedName
}{
	data: make(map[string]map[string]string),
}

// buildMySQLKeywords 构建 MySQL 8.0 保留关键字集合
func buildMySQLKeywords() map[string]bool {
	return map[string]bool{
		"accessible": true, "add": true, "all": true, "alter": true,
		"analyze": true, "and": true, "as": true, "asc": true,
		"asensitive": true, "before": true, "between": true, "bigint": true,
		"binary": true, "blob": true, "both": true, "by": true,
		"call": true, "cascade": true, "case": true, "change": true,
		"char": true, "character": true, "check": true, "collate": true,
		"column": true, "condition": true, "constraint": true, "continue": true,
		"convert": true, "create": true, "cross": true, "cube": true,
		"cume_dist": true, "current_date": true, "current_time": true,
		"current_timestamp": true, "current_user": true, "cursor": true,
		"database": true, "databases": true, "day_hour": true,
		"day_microsecond": true, "day_minute": true, "day_second": true,
		"dec": true, "decimal": true, "declare": true, "default": true,
		"delayed": true, "delete": true, "dense_rank": true, "desc": true,
		"describe": true, "deterministic": true, "distinct": true,
		"distinctrow": true, "div": true, "double": true, "drop": true,
		"dual": true, "each": true, "else": true, "elseif": true,
		"empty": true, "enclosed": true, "escaped": true, "except": true,
		"exists": true, "exit": true, "explain": true, "false": true,
		"fetch": true, "first_value": true, "float": true, "float4": true,
		"float8": true, "for": true, "force": true, "foreign": true,
		"from": true, "fulltext": true, "function": true, "generated": true,
		"get": true, "grant": true, "group": true, "grouping": true,
		"groups": true, "having": true, "high_priority": true,
		"hour_microsecond": true, "hour_minute": true, "hour_second": true,
		"if": true, "ignore": true, "in": true, "index": true,
		"infile": true, "inner": true, "inout": true, "insensitive": true,
		"insert": true, "int": true, "int1": true, "int2": true,
		"int3": true, "int4": true, "int8": true, "integer": true,
		"interval": true, "into": true, "io_after_gtids": true,
		"io_before_gtids": true, "is": true, "iterate": true, "join": true,
		"json_table": true, "key": true, "keys": true, "kill": true,
		"lag": true, "last_value": true, "lateral": true, "lead": true,
		"leading": true, "leave": true, "left": true, "like": true,
		"limit": true, "linear": true, "lines": true, "load": true,
		"localtime": true, "localtimestamp": true, "lock": true, "long": true,
		"longblob": true, "longtext": true, "loop": true, "low_priority": true,
		"master_bind": true, "master_ssl_verify_server_cert": true,
		"match": true, "maxvalue": true, "mediumblob": true,
		"mediumint": true, "mediumtext": true, "middleint": true,
		"minute_microsecond": true, "minute_second": true, "mod": true,
		"modifies": true, "natural": true, "not": true, "no_write_to_binlog": true,
		"nth_value": true, "ntile": true, "null": true, "numeric": true,
		"of": true, "on": true, "optimize": true, "optimizer_costs": true,
		"option": true, "optionally": true, "or": true, "order": true,
		"out": true, "outer": true, "outfile": true, "over": true,
		"partition": true, "percent_rank": true, "precision": true,
		"primary": true, "procedure": true, "purge": true, "range": true,
		"rank": true, "read": true, "reads": true, "read_write": true,
		"real": true, "recursive": true, "references": true, "regexp": true,
		"release": true, "rename": true, "repeat": true, "replace": true,
		"require": true, "resignal": true, "restrict": true, "return": true,
		"revoke": true, "right": true, "rlike": true, "row": true,
		"rows": true, "row_number": true, "schema": true, "schemas": true,
		"second_microsecond": true, "select": true, "sensitive": true,
		"separator": true, "set": true, "show": true, "signal": true,
		"smallint": true, "spatial": true, "specific": true, "sql": true,
		"sqlexception": true, "sqlstate": true, "sqlwarning": true,
		"sql_big_result": true, "sql_calc_found_rows": true,
		"sql_small_result": true, "ssl": true, "starting": true,
		"stored": true, "straight_join": true, "system": true, "table": true,
		"terminated": true, "then": true, "tinyblob": true, "tinyint": true,
		"tinytext": true, "to": true, "trailing": true, "trigger": true,
		"true": true, "undo": true, "union": true, "unique": true,
		"unlock": true, "unsigned": true, "update": true, "usage": true,
		"use": true, "using": true, "utc_date": true, "utc_time": true,
		"utc_timestamp": true, "values": true, "varbinary": true,
		"varchar": true, "varcharacter": true, "varying": true,
		"virtual": true, "when": true, "where": true, "while": true,
		"window": true, "with": true, "write": true, "xor": true,
		"year_month": true, "zerofill": true,
	}
}

// buildPostgresKeywords 构建 PostgreSQL 15 保留关键字集合
func buildPostgresKeywords() map[string]bool {
	return map[string]bool{
		"all": true, "analyse": true, "analyze": true, "and": true,
		"any": true, "array": true, "as": true, "asc": true,
		"asymmetric": true, "both": true, "case": true, "cast": true,
		"check": true, "collate": true, "column": true, "constraint": true,
		"create": true, "cross": true, "current_date": true,
		"current_role": true, "current_time": true,
		"current_timestamp": true, "current_user": true, "default": true,
		"deferrable": true, "desc": true, "distinct": true, "do": true,
		"else": true, "end": true, "except": true, "false": true,
		"fetch": true, "for": true, "foreign": true, "from": true,
		"grant": true, "group": true, "having": true, "in": true,
		"initially": true, "inner": true, "intersect": true, "into": true,
		"lateral": true, "leading": true, "limit": true, "localtime": true,
		"localtimestamp": true, "natural": true, "not": true, "null": true,
		"of": true, "off": true, "offset": true, "on": true,
		"only": true, "or": true, "order": true, "outer": true,
		"over": true, "overlaps": true, "placing": true, "primary": true,
		"references": true, "returning": true, "right": true, "select": true,
		"session_user": true, "similar": true, "some": true, "symmetric": true,
		"table": true, "then": true, "to": true, "trailing": true,
		"true": true, "union": true, "unique": true, "user": true,
		"using": true, "variadic": true, "when": true, "where": true,
		"window": true, "with": true,
	}
}

// buildSQLiteKeywords 构建 SQLite 3 保留关键字集合
func buildSQLiteKeywords() map[string]bool {
	return map[string]bool{
		"abort": true, "action": true, "add": true, "after": true,
		"all": true, "alter": true, "analyze": true, "and": true,
		"as": true, "asc": true, "attach": true, "autoincrement": true,
		"before": true, "begin": true, "between": true, "by": true,
		"cascade": true, "case": true, "cast": true, "check": true,
		"collate": true, "column": true, "commit": true, "conflict": true,
		"constraint": true, "create": true, "cross": true, "current_date": true,
		"current_time": true, "current_timestamp": true, "database": true,
		"default": true, "deferrable": true, "deferred": true,
		"delete": true, "desc": true, "detach": true, "distinct": true,
		"drop": true, "each": true, "else": true, "end": true,
		"escape": true, "except": true, "exclusive": true, "exists": true,
		"explain": true, "fail": true, "for": true, "foreign": true,
		"from": true, "full": true, "glob": true, "group": true,
		"having": true, "if": true, "ignore": true, "immediate": true,
		"in": true, "index": true, "indexed": true, "initially": true,
		"inner": true, "insert": true, "instead": true, "intersect": true,
		"into": true, "is": true, "isnull": true, "join": true,
		"key": true, "left": true, "like": true, "limit": true,
		"match": true, "natural": true, "no": true, "not": true,
		"notnull": true, "null": true, "of": true, "offset": true,
		"on": true, "or": true, "order": true, "outer": true,
		"plan": true, "pragma": true, "primary": true, "query": true,
		"raise": true, "recursive": true, "references": true,
		"regexp": true, "reindex": true, "release": true, "rename": true,
		"replace": true, "restrict": true, "right": true, "rollback": true,
		"row": true, "savepoint": true, "select": true, "set": true,
		"table": true, "temp": true, "temporary": true, "then": true,
		"to": true, "transaction": true, "trigger": true, "union": true,
		"unique": true, "update": true, "using": true, "vacuum": true,
		"values": true, "view": true, "virtual": true, "when": true,
		"where": true, "with": true, "without": true,
	}
}

// buildClickHouseKeywords 构建 ClickHouse 保留关键字集合
func buildClickHouseKeywords() map[string]bool {
	return map[string]bool{
		"add": true, "after": true, "all": true, "alter": true,
		"and": true, "any": true, "array": true, "as": true,
		"asc": true, "assert": true, "at": true, "authorization": true,
		"before": true, "between": true, "both": true, "by": true,
		"cache": true, "cascade": true, "case": true, "cast": true,
		"change": true, "check": true, "cluster": true, "collate": true,
		"column": true, "constraint": true, "create": true, "cross": true,
		"cube": true, "current": true, "database": true, "databases": true,
		"day": true, "deallocate": true, "declare": true, "default": true,
		"delay": true, "delete": true, "desc": true, "describe": true,
		"distinct": true, "distributed": true, "drop": true, "dual": true,
		"else": true, "end": true, "engine": true, "escape": true,
		"except": true, "exchange": true, "exists": true, "extract": true,
		"false": true, "fetch": true, "first": true, "float": true,
		"for": true, "format": true, "from": true, "full": true,
		"function": true, "global": true, "grant": true, "group": true,
		"grouping": true, "having": true, "hour": true, "if": true,
		"ilike": true, "in": true, "index": true, "inner": true,
		"insert": true, "interval": true, "into": true, "is": true,
		"isnull": true, "join": true, "key": true, "last": true,
		"lateral": true, "left": true, "like": true, "limit": true,
		"local": true, "merge": true, "minute": true, "month": true,
		"natural": true, "no": true, "not": true, "notnull": true,
		"null": true, "nulls": true, "offset": true, "on": true,
		"only": true, "or": true, "order": true, "outer": true,
		"out": true, "over": true, "partition": true, "precision": true,
		"primary": true, "procedure": true, "purge": true, "quart": true,
		"range": true, "reads": true, "recursive": true, "refresh": true,
		"rename": true, "repeat": true, "replace": true, "restrict": true,
		"return": true, "revoke": true, "right": true, "rollback": true,
		"rollup": true, "row": true, "rows": true, "second": true,
		"select": true, "session": true, "set": true, "show": true,
		"some": true, "start": true, "substring": true, "table": true,
		"tables": true, "temporary": true, "then": true, "time": true,
		"timestamp": true, "to": true, "trailing": true, "transaction": true,
		"true": true, "truncate": true, "union": true, "unique": true,
		"update": true, "using": true, "user": true, "validate": true,
		"values": true, "view": true, "virtual": true, "when": true,
		"where": true, "window": true, "with": true, "within": true,
		"without": true, "work": true, "write": true, "year": true,
	}
}

// getReservedKeywords 获取指定数据库类型的保留关键字集合
func getReservedKeywords(clientType string) map[string]bool {
	reservedKeywords.RLock()
	defer reservedKeywords.RUnlock()

	// 处理复用关键字的情况
	keywords, ok := reservedKeywords.data[clientType]
	if !ok || keywords == nil {
		switch clientType {
		case TiDB:
			return reservedKeywords.data[MySQL]
		case GaussDB:
			return reservedKeywords.data[Postgres]
		default:
			return make(map[string]bool)
		}
	}
	return keywords
}

// isReservedKeyword 检查字段名是否为保留关键字
func isReservedKeyword(fieldName, clientType string) bool {
	if fieldName == "" {
		return false
	}

	// 转换为小写进行比较（SQL 不区分大小写）
	fieldNameLower := strings.ToLower(fieldName)

	keywords := getReservedKeywords(clientType)
	return keywords[fieldNameLower]
}

// containsSpecialChars 检查字段名是否包含需要引用的特殊字符
func containsSpecialChars(fieldName string) bool {
	if fieldName == "" {
		return false
	}

	// 检查首字符是否为数字
	if fieldName[0] >= '0' && fieldName[0] <= '9' {
		return true
	}

	// 检查是否包含特殊字符
	for i := 0; i < len(fieldName); i++ {
		c := fieldName[i]
		// 允许: 字母、数字、下划线
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return true
	}

	return false
}

// needsQuoting 检查字段名是否需要引用（考虑保留关键字和特殊字符）
func needsQuoting(fieldName, clientType string) bool {
	// 检查是否为保留关键字
	if isReservedKeyword(fieldName, clientType) {
		return true
	}

	// 检查是否包含特殊字符
	if containsSpecialChars(fieldName) {
		return true
	}

	return false
}

// applyQuotes 根据数据库类型应用正确的引号
func applyQuotes(fieldName, clientType string) string {
	switch clientType {
	case MySQL, TiDB, ClickHouse:
		return "`" + fieldName + "`"
	case Postgres, GaussDB, Sqlite:
		return "\"" + fieldName + "\""
	default:
		// 默认使用反引号
		return "`" + fieldName + "`"
	}
}

// shouldQuoteFieldName 检查是否应该引用字段名（带缓存）
func shouldQuoteFieldName(fieldName, clientType string) (string, bool) {
	if fieldName == "" {
		return fieldName, false
	}

	// 检查是否已经引用
	if strings.HasPrefix(fieldName, "`") || strings.HasPrefix(fieldName, "\"") {
		return fieldName, false
	}

	// 检查缓存
	quoteCache.RLock()
	if cache, ok := quoteCache.data[clientType]; ok {
		if quoted, exists := cache[fieldName]; exists {
			quoteCache.RUnlock()
			return quoted, quoted != fieldName
		}
	}
	quoteCache.RUnlock()

	// 检查是否需要引用
	if !needsQuoting(fieldName, clientType) {
		// 缓存不需要引用的结果
		quoteCache.Lock()
		if quoteCache.data[clientType] == nil {
			quoteCache.data[clientType] = make(map[string]string)
		}
		quoteCache.data[clientType][fieldName] = fieldName
		quoteCache.Unlock()
		return fieldName, false
	}

	// 应用引号
	quoted := applyQuotes(fieldName, clientType)

	// 缓存结果
	quoteCache.Lock()
	if quoteCache.data[clientType] == nil {
		quoteCache.data[clientType] = make(map[string]string)
	}
	quoteCache.data[clientType][fieldName] = quoted
	quoteCache.Unlock()

	return quoted, true
}
