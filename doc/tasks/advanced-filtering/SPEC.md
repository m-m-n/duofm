# Feature: Advanced File Filtering with SQL-like Syntax

## Overview

Add a powerful SQL-like filtering capability to duofm that allows users to filter files based on various attributes (name, size, modification time, type, permissions, etc.) using familiar SQL WHERE clause syntax. This feature complements the existing incremental search (`/`) and regex search (`Ctrl+F`) with a more expressive query language.

## Objectives

- Provide SQL-like WHERE clause syntax for filtering files by multiple attributes
- Support comparison operators (=, !=, <, >, <=, >=), pattern matching (LIKE, ILIKE), and logical operators (AND, OR, NOT)
- Support human-readable size literals (KB, MiB, GB, etc.) and date/time functions
- Integrate seamlessly with the existing search/filter architecture
- Provide helpful error messages with position and hints for syntax errors

## User Stories

### US1: Filter by File Size
As a system administrator, I want to filter files larger than 1GB, so that I can identify large files consuming disk space.

**Acceptance Criteria:**
- [ ] User can press Ctrl+G to open SQL-like filter
- [ ] User can type `size > 1GiB` and press Enter
- [ ] Only files larger than 1 GiB are displayed
- [ ] Size units (KiB, MiB, GiB, TiB, KB, MB, GB, TB) are supported

### US2: Filter by Modification Date
As a developer, I want to find files modified in the last 7 days, so that I can track recent changes.

**Acceptance Criteria:**
- [ ] User can type `mtime > now() - 7d`
- [ ] Only recently modified files are displayed
- [ ] ISO 8601 date format is supported: `mtime > '2024-01-15'`

### US3: Combine Multiple Conditions
As a power user, I want to combine conditions with AND/OR/NOT, so that I can create precise filters.

**Acceptance Criteria:**
- [ ] User can type `size > 1GiB AND year(mtime) = 2024`
- [ ] User can use parentheses for grouping: `(size > 1GiB OR name LIKE '%.mp4') AND NOT type = 'dir'`
- [ ] Operator precedence follows SQL standard: NOT > AND > OR

### US4: Pattern Matching
As a user, I want to filter files by name patterns, so that I can find files with specific naming conventions.

**Acceptance Criteria:**
- [ ] LIKE operator supports `%` (any string) and `_` (single char)
- [ ] ILIKE operator for case-insensitive matching
- [ ] NOT LIKE for exclusion patterns

### US5: Clear Error Messages
As a user, when I make a syntax error, I want a helpful message showing where the error is.

**Acceptance Criteria:**
- [ ] Error message includes position of the error
- [ ] Error message includes hint about what was expected
- [ ] User can correct the error without restarting

### US6: In-App Help Reference
As a user learning the SQL-like filter syntax, I want to see a quick reference in the help screen.

**Acceptance Criteria:**
- [ ] Ctrl+G keybinding shown in Display & Search section
- [ ] SQL-like Filter section with columns, operators, examples
- [ ] Size units and date format documented
- [ ] Wildcard note (% not *) included

## Technical Requirements

### Functional Requirements

**FR1: Query Syntax**
- FR1.1: Parser shall accept WHERE clause syntax (WHERE keyword optional for pasting convenience)
- FR1.2: Keywords and column names shall be case-insensitive
- FR1.3: String literals shall be enclosed in single quotes
- FR1.4: Empty query shall clear the filter
- FR1.5: Single quotes within string literals shall be escaped by doubling (`''`), following SQL standard (e.g., `name = 'file''s name.txt'`)

**FR2: Supported Columns**
- FR2.1: `name` (string) - File name
- FR2.2: `size` (int64) - File size in bytes
- FR2.3: `mtime` (timestamp) - Modification time
- FR2.4: `type` (string) - 'file', 'dir', 'symlink'
- FR2.5: `ext` (string) - File extension (**without** leading dot: 'txt', 'go', 'rs'). NULL for files starting with dot only (e.g., `.gitignore`). Only the last extension (e.g., `archive.tar.gz` → 'gz')
- FR2.6: `perm` (string) - Permission string (e.g., "-rwxr-xr-x")
- FR2.7: `owner` (string) - Owner name (always has a value: username, UID string, "unknown", or "N/A")
- FR2.8: `group` (string) - Group name (always has a value: group name, GID string, "unknown", or "N/A")
- FR2.9: `isdir` (bool) - Is directory
- FR2.10: `isfile` (bool) - Is regular file
- FR2.11: `issymlink` (bool) - Is symbolic link

**FR3: Comparison Operators**
- FR3.1: `=` Equal
- FR3.2: `!=`, `<>` Not equal
- FR3.3: `<` Less than
- FR3.4: `>` Greater than
- FR3.5: `<=` Less than or equal
- FR3.6: `>=` Greater than or equal

**FR4: Pattern Matching Operators**
- FR4.1: `LIKE` - SQL pattern matching (case-sensitive)
- FR4.2: `ILIKE` - Case-insensitive pattern matching
- FR4.3: `NOT LIKE` - Negated pattern matching
- FR4.4: `%` wildcard matches zero or more characters
- FR4.5: `_` wildcard matches exactly one character
- FR4.6: Shell wildcards (`*`, `?`) are NOT supported

**FR5: Logical Operators**
- FR5.1: `AND` - Logical conjunction
- FR5.2: `OR` - Logical disjunction
- FR5.3: `NOT` - Logical negation
- FR5.4: Parentheses `()` for grouping
- FR5.5: Operator precedence: NOT > AND > OR

**FR6: Size Literals**
- FR6.1: Binary units (1024-based): KiB, MiB, GiB, TiB
- FR6.2: Decimal units (1000-based): KB, MB, GB, TB
- FR6.3: Bare numbers treated as bytes

**FR7: Date/Time Literals and Functions**
- FR7.1: ISO 8601 format: `'2024-01-15'`, `'2024-01-15 10:30:00'`
- FR7.2: `now()` - Current timestamp
- FR7.3: `now() - Nd` - N days ago
- FR7.4: `now() - Nh` - N hours ago
- FR7.5: `now() - Nm` - N minutes ago
- FR7.6: `year(column)` - Extract year
- FR7.7: `month(column)` - Extract month
- FR7.8: `day(column)` - Extract day
- FR7.9: Date/time literals shall be interpreted in the local timezone (system default)
- FR7.10: `now()` returns the current local time

**FR8: String Functions**
- FR8.1: `lower(column)` - Convert to lowercase
- FR8.2: `upper(column)` - Convert to uppercase

**FR9: Additional Operators**
- FR9.1: `IN (value1, value2, ...)` - Value in list
- FR9.2: `NOT IN (value1, value2, ...)` - Value not in list
- FR9.3: `IS NULL` - Value is null (**ext column only**)
- FR9.4: `IS NOT NULL` - Value is not null (**ext column only**)

**FR10: NULL Handling**
- FR10.1: **Only the `ext` column can be NULL** (for files with no extension or dot-only files like `.gitignore`)
- FR10.2: All other columns (name, size, mtime, type, perm, owner, group, isdir, isfile, issymlink) always have values
- FR10.3: owner/group always have values (username, UID/GID string, "unknown", or "N/A")
- FR10.4: Comparisons with NULL return false
- FR10.5: Use `IS NULL`/`IS NOT NULL` for NULL checks on ext column

**FR11: Error Handling**
- FR11.1: Syntax errors shall include position
- FR11.2: Syntax errors shall include helpful hints
- FR11.3: Unknown columns shall be reported
- FR11.4: Type mismatches shall be reported

**FR12: Help Screen Integration**
- FR12.1: Add `Ctrl+G : SQL-like filter (advanced)` to Display & Search keybindings
- FR12.2: Add new "SQL-like Filter" section with:
  - Available columns list
  - Supported operators
  - Usage examples
  - Size unit reference (KiB/MiB/GiB vs KB/MB/GB)
  - Date format reference
  - Wildcard note (use % not *)
- FR12.3: Implementation location: `internal/ui/help_dialog.go` buildContent()

### Non-Functional Requirements

**NFR1 - Performance:**
- Query parsing: < 10ms
- Filter execution: < 100ms for 10,000 files
- UI response: < 50ms

**NFR2 - Usability:**
- Error messages shall be user-friendly with actionable hints
- Query syntax shall be familiar to SQL users
- Existing search features remain unchanged

**NFR3 - Maintainability:**
- Parser test coverage: > 90%
- Each operator shall have independent test cases
- Code shall be well-documented

## Implementation Approach

### Architecture

**Layered Architecture:**
```
+--------------------------------------------------+
|                   UI Layer                        |
|  - SQLFilterMinibuffer (input handling)          |
|  - Error display                                  |
|  - Keybinding (Ctrl+G)                           |
+--------------------------------------------------+
|                  Parser Layer                     |
|  - Lexer (tokenization)                          |
|  - Parser (AST generation)                        |
|  - Error reporter                                 |
+--------------------------------------------------+
|                Evaluator Layer                    |
|  - Expression evaluator                           |
|  - Type coercion                                  |
|  - Function implementation                        |
+--------------------------------------------------+
|               Integration Layer                   |
|  - FilterSQLLike function                         |
|  - SearchMode extension                           |
+--------------------------------------------------+
```

### Component Design

#### Lexer (Tokenizer)

```go
type TokenType int

const (
    TokenEOF TokenType = iota
    TokenIdent        // column names, keywords
    TokenString       // 'string literal'
    TokenNumber       // 123, 1.5
    TokenSizeUnit     // K, M, G, KiB, MiB, etc.
    TokenDuration     // 7d, 1h, 30m
    TokenOperator     // =, !=, <, >, <=, >=, <>, LIKE, ILIKE
    TokenLogical      // AND, OR, NOT
    TokenLParen       // (
    TokenRParen       // )
    TokenComma        // ,
    TokenMinus        // -
)

type Token struct {
    Type    TokenType
    Value   string
    Pos     int  // Position in input
}

type Lexer struct {
    input   string
    pos     int
    tokens  []Token
}

func (l *Lexer) Tokenize() ([]Token, error)
```

#### Parser (AST Generator)

```go
// AST Node Types
type Expression interface {
    Evaluate(entry fs.FileEntry) (interface{}, error)
}

type BinaryExpr struct {
    Left     Expression
    Operator string
    Right    Expression
}

type UnaryExpr struct {
    Operator string
    Operand  Expression
}

type ColumnRef struct {
    Name string
}

type Literal struct {
    Value interface{} // string, int64, time.Time, bool
}

type FunctionCall struct {
    Name string
    Args []Expression
}

type InExpr struct {
    Column  Expression
    Values  []Expression
    Negated bool
}

type Parser struct {
    tokens  []Token
    pos     int
}

func (p *Parser) Parse() (Expression, error)
```

#### Evaluator

```go
type Evaluator struct {
    entry fs.FileEntry
}

func (e *Evaluator) Eval(expr Expression) (bool, error)

// Column value getters
func (e *Evaluator) getColumnValue(name string) (interface{}, error)

// Type coercion
func coerceToInt64(v interface{}) (int64, error)
func coerceToString(v interface{}) (string, error)
func coerceToTime(v interface{}) (time.Time, error)
func coerceToBool(v interface{}) (bool, error)
```

### Data Flow

```
User Input (Ctrl+G)
    |
    v
+-------------------+
| "size > 1GiB"     |  Raw input
+-------------------+
    |
    v
+-------------------+
| Lexer             |  Tokenization
+-------------------+
    |
    v
+-------------------+
| [IDENT:size]      |  Token stream
| [OP:>]            |
| [NUM:1]           |
| [SIZE:GiB]        |
+-------------------+
    |
    v
+-------------------+
| Parser            |  AST generation
+-------------------+
    |
    v
+-------------------+
| BinaryExpr{       |  AST
|   Left: Column    |
|   Op: ">"         |
|   Right: Literal  |
| }                 |
+-------------------+
    |
    v
+-------------------+
| Evaluator         |  For each FileEntry
+-------------------+
    |
    v
+-------------------+
| true/false        |  Include in result?
+-------------------+
```

### Grammar (EBNF)

```ebnf
query      = [ "WHERE" ] expression ;

expression = or_expr ;

or_expr    = and_expr { "OR" and_expr } ;

and_expr   = not_expr { "AND" not_expr } ;

not_expr   = [ "NOT" ] primary ;

primary    = comparison
           | "(" expression ")"
           | column "IN" "(" value_list ")"
           | column "NOT" "IN" "(" value_list ")"
           | column "IS" "NULL"
           | column "IS" "NOT" "NULL"
           | column  (* boolean column like isdir *)
           ;

comparison = operand compare_op operand ;

compare_op = "=" | "!=" | "<>" | "<" | ">" | "<=" | ">="
           | "LIKE" | "ILIKE" | "NOT" "LIKE" ;

operand    = column
           | literal
           | function_call
           | "(" expression ")"
           ;

column     = identifier ;

literal    = string_literal
           | number [ size_unit ]
           | date_literal
           | "now" "(" ")" [ "-" duration ]
           ;

function_call = function_name "(" [ operand { "," operand } ] ")" ;

function_name = "year" | "month" | "day" | "lower" | "upper" | "ext" | "now" ;

value_list = literal { "," literal } ;

size_unit  = "KiB" | "MiB" | "GiB" | "TiB"
           | "KB" | "MB" | "GB" | "TB" ;

duration   = number ( "d" | "h" | "m" ) ;

string_literal = "'" { string_char | "''" } "'" ;  (* '' escapes single quote *)

string_char    = ? any character except "'" ? ;   (* unescaped single quote not allowed *)

(* Note: date_literal is lexed as string_literal; the evaluator parses ISO 8601 format when comparing with timestamp columns *)

date_format = YYYY-MM-DD [ " " HH:MM:SS ] ;
```

### API Design

#### Filter Function

```go
// FilterSQLLike filters entries using SQL-like WHERE clause
func FilterSQLLike(entries []fs.FileEntry, query string) ([]fs.FileEntry, error)

// ParseQuery parses a SQL-like WHERE clause and returns an AST
func ParseQuery(query string) (Expression, error)

// CompileQuery parses and validates a query, returning a compiled filter
func CompileQuery(query string) (*CompiledQuery, error)

type CompiledQuery struct {
    ast Expression
}

func (q *CompiledQuery) Match(entry fs.FileEntry) bool
```

#### Error Types

```go
type QueryError struct {
    Message  string
    Position int
    Hint     string
}

func (e *QueryError) Error() string {
    return fmt.Sprintf("Error at position %d: %s\nHint: %s",
        e.Position, e.Message, e.Hint)
}
```

#### Search Mode Extension

```go
const (
    SearchModeNone SearchMode = iota
    SearchModeIncremental  // Existing: /
    SearchModeRegex        // Existing: Ctrl+F
    SearchModeSQLLike      // New: Ctrl+G
)
```

### Dependencies

**Internal Dependencies:**
- `internal/fs`: FileEntry struct
- `internal/ui`: Minibuffer, SearchState, Model

**External Dependencies:**
- None (uses Go standard library only)

### File Structure

```
internal/
+-- filter/
|   +-- lexer.go              # Tokenizer
|   +-- lexer_test.go
|   +-- parser.go             # AST generator
|   +-- parser_test.go
|   +-- ast.go                # AST node definitions
|   +-- evaluator.go          # Expression evaluator
|   +-- evaluator_test.go
|   +-- functions.go          # Built-in functions
|   +-- functions_test.go
|   +-- errors.go             # Error types
|   +-- filter.go             # Main filter function
|   +-- filter_test.go
+-- ui/
    +-- search.go             # Add SearchModeSQLLike
    +-- model_update.go       # Add Ctrl+G handler
    +-- pane_filter.go        # Add FilterSQLLike call
```

## Test Scenarios

### Unit Tests

**Lexer Tests:**
- [ ] Tokenize simple comparison: `size > 100`
- [ ] Tokenize string literals: `name LIKE '%.txt'`
- [ ] Tokenize escaped single quotes: `name = 'file''s name.txt'` -> `file's name.txt`
- [ ] Tokenize size units: `1GiB`, `100MB`, `1KiB`
- [ ] Tokenize date literals: `'2024-01-15'`
- [ ] Tokenize duration: `now() - 7d`
- [ ] Tokenize complex query with parentheses
- [ ] Report error on unterminated string
- [ ] Report error on unescaped single quote in string: `name = 'file's name'` (should error, use `'file''s name'`)
- [ ] Report error on invalid token

**Parser Tests:**
- [ ] Parse simple comparison
- [ ] Parse AND expression
- [ ] Parse OR expression
- [ ] Parse NOT expression
- [ ] Parse nested parentheses
- [ ] Parse LIKE with pattern
- [ ] Parse IN list
- [ ] Parse IS NULL
- [ ] Parse function call
- [ ] Report error on missing operand
- [ ] Report error on unknown column
- [ ] Report error on unbalanced parentheses

**Evaluator Tests:**
- [ ] Evaluate size comparison with bytes
- [ ] Evaluate size comparison with KiB/MiB/GiB
- [ ] Evaluate size comparison with KB/MB/GB
- [ ] Evaluate string equality
- [ ] Evaluate LIKE with `%` wildcard
- [ ] Evaluate LIKE with `_` wildcard
- [ ] Evaluate ILIKE case-insensitivity
- [ ] Evaluate date comparison
- [ ] Evaluate now() - duration
- [ ] Evaluate year/month/day functions
- [ ] Evaluate lower/upper functions
- [ ] Evaluate ext function
- [ ] Evaluate AND/OR/NOT combinations
- [ ] Evaluate IN list
- [ ] Evaluate IS NULL/IS NOT NULL
- [ ] Evaluate boolean columns (isdir, isfile, issymlink)
- [ ] Handle NULL values correctly

**Integration Tests:**
- [ ] Filter by size > 1GiB
- [ ] Filter by mtime > now() - 7d
- [ ] Filter by name LIKE '%.txt'
- [ ] Filter with combined conditions
- [ ] Filter clears when query is empty
- [ ] Error message displays correctly

### E2E Tests

**E2E Test 1: Basic Size Filter**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "C-g"  # Ctrl+G
sleep 0.3
assert_contains "$CURRENT_SESSION" "WHERE" "SQL filter prompt visible"
send_keys "$CURRENT_SESSION" "s" "i" "z" "e" " " ">" " " "0"
send_keys "$CURRENT_SESSION" "Enter"
sleep 0.3
# Verify filter is applied
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 2: Cancel Filter**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "C-g"
send_keys "$CURRENT_SESSION" "s" "i" "z" "e" " " ">" " " "1" "G" "i" "B"
send_keys "$CURRENT_SESSION" "Escape"
sleep 0.3
assert_not_contains "$CURRENT_SESSION" "WHERE" "Filter cancelled"
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 3: Syntax Error**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "C-g"
send_keys "$CURRENT_SESSION" "s" "i" "z" "e" " " ">" ">"  # Invalid >>
send_keys "$CURRENT_SESSION" "Enter"
sleep 0.3
assert_contains "$CURRENT_SESSION" "Error" "Syntax error displayed"
stop_duofm "$CURRENT_SESSION"
```

### Edge Cases

- [ ] Empty query (should clear filter)
- [ ] Query with only whitespace
- [ ] Very long query (>1000 characters)
- [ ] Unicode in file names and patterns
- [ ] Size of 0 bytes
- [ ] Files with no extension
- [ ] Hidden files (starting with .)
- [ ] Symbolic links
- [ ] NULL owner/group (depending on filesystem)
- [ ] Date in different timezones
- [ ] Leap year dates
- [ ] Deeply nested parentheses
- [ ] Many OR conditions

### Performance Tests

- [ ] Parse query < 10ms
- [ ] Filter 10,000 files < 100ms
- [ ] Filter 100,000 files < 1s
- [ ] Memory usage stays bounded during filter

## Security Considerations

- **Input Validation:** All user input is validated by the parser
- **No Code Execution:** The filter only evaluates expressions, no arbitrary code execution
- **Bounded Recursion:** Parser has depth limit to prevent stack overflow
- **Memory Safety:** No unbounded allocations during parsing

## Error Handling

### Error Codes

| Code | Description | User Message |
|------|-------------|--------------|
| ERR_SYNTAX | Syntax error | "Syntax error at position {pos}: {detail}" |
| ERR_UNKNOWN_COLUMN | Unknown column reference | "Unknown column: {name}. Valid columns: name, size, mtime, type, ext, perm, owner, group, isdir, isfile, issymlink" |
| ERR_TYPE_MISMATCH | Type mismatch in comparison | "Cannot compare {type1} with {type2}" |
| ERR_INVALID_PATTERN | Invalid LIKE pattern | "Invalid pattern: {detail}" |
| ERR_INVALID_SIZE | Invalid size literal | "Invalid size: {value}. Valid units: KiB, MiB, GiB, TiB, KB, MB, GB, TB" |
| ERR_INVALID_DATE | Invalid date literal | "Invalid date: {value}. Use format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS" |
| ERR_INVALID_FUNCTION | Unknown function | "Unknown function: {name}. Valid functions: now, year, month, day, lower, upper, ext" |

### Error Flow

```
Parse Input
    |
    +-- Lexer Error --> Display error with position
    |
    +-- Parser Error --> Display error with hint
    |
    +-- Validation Error --> Display column/type error
    |
    +-- Success --> Apply filter
```

## Performance Optimization

### Performance Goals

- Query parsing: < 10ms
- 10,000 file filter: < 100ms
- Memory: O(n) where n = number of files

### Optimization Strategies

**Compile Once, Evaluate Many:**
- Parse query once, create reusable AST
- Avoid re-parsing for each file

**Short-Circuit Evaluation:**
- AND: Stop on first false
- OR: Stop on first true

**Column Value Caching:**
- Cache computed values (ext, type) within single evaluation
- Avoid redundant computation for same entry

**Lazy Evaluation for Functions:**
- Compute now() once per filter operation
- Extract date parts only when needed

## Success Criteria

- [ ] All functional requirements implemented and tested
- [ ] Parser test coverage > 90%
- [ ] Performance goals met (< 100ms for 10,000 files)
- [ ] Error messages are clear and actionable
- [ ] Integration with existing search is seamless
- [ ] Documentation is complete
- [ ] Code review completed

## Open Questions

- [ ] Should we support BETWEEN operator? (e.g., `size BETWEEN 1M AND 1G`)
- [ ] Should we support GLOB pattern matching in addition to LIKE?
- [ ] Should we add query history?
- [ ] Should we allow saving named filters?

## Implementation Phases

### Phase 1: Core Parser (High Priority)
**Goals:** Implement lexer and parser for basic expressions

**Deliverables:**
- Lexer with all token types
- Parser for comparison expressions
- AND/OR/NOT operators
- Parentheses grouping
- Unit tests for lexer and parser

**Estimated Effort:** 3-4 days

### Phase 2: Evaluator (High Priority)
**Goals:** Implement expression evaluation

**Deliverables:**
- Column value extraction from FileEntry
- Comparison operators
- LIKE/ILIKE pattern matching
- Size literal handling
- Unit tests for evaluator

**Estimated Effort:** 2-3 days

### Phase 3: Date/Time Support (Medium Priority)
**Goals:** Add date/time literals and functions

**Deliverables:**
- Date literal parsing
- now() function
- Relative time (now() - 7d)
- year/month/day functions
- Unit tests

**Estimated Effort:** 2 days

### Phase 4: UI Integration (High Priority)
**Goals:** Integrate with duofm UI

**Deliverables:**
- SearchModeSQLLike constant
- Ctrl+G keybinding
- Minibuffer with "(query): " prompt
- Error display
- Integration tests

**Estimated Effort:** 2 days

### Phase 5: Polish and Testing (High Priority)
**Goals:** Complete testing and documentation

**Deliverables:**
- E2E tests
- Performance benchmarks
- Error message improvements
- Documentation

**Estimated Effort:** 2 days

**Total Estimated Effort:** 11-13 days

## References

- duofm FileEntry: `/home/sakura/cache/worktrees/duofm/feature-advanced-filtering/internal/fs/types.go`
- duofm Search: `/home/sakura/cache/worktrees/duofm/feature-advanced-filtering/internal/ui/search.go`
- duofm Minibuffer: `/home/sakura/cache/worktrees/duofm/feature-advanced-filtering/internal/ui/minibuffer.go`
- SQL LIKE operator: https://www.postgresql.org/docs/current/functions-matching.html
- IEC 80000-13 (Binary prefixes): https://en.wikipedia.org/wiki/Binary_prefix

---

**Status:** Draft
