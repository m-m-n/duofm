# Implementation Plan: Advanced File Filtering with SQL-like Syntax

## Overview

Implement a SQL-like WHERE clause filtering system for duofm that enables users to filter files based on various attributes (name, size, modification time, type, permissions, etc.) using familiar SQL syntax. This complements the existing incremental search (`/`) and regex search (`Ctrl+F`).

## Objectives

- Provide SQL-like WHERE clause syntax for filtering files by multiple attributes
- Support comparison operators, pattern matching (LIKE, ILIKE), and logical operators (AND, OR, NOT)
- Support human-readable size literals (KiB, MiB, GB, etc.) and date/time functions
- Integrate seamlessly with the existing search/filter architecture
- Provide helpful error messages with position and hints for syntax errors

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)
- Existing duofm development environment set up

### Dependencies
- No external dependencies (uses Go standard library only)
- Internal: `internal/fs` (FileEntry), `internal/ui` (Minibuffer, SearchState, Model)

### Knowledge Requirements
- Go's lexer/parser patterns (recursive descent parser)
- Bubble Tea architecture (Model-Update-View)
- SQL WHERE clause syntax and operator precedence

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Key Libraries**:
  - Lip Gloss - TUI styling
  - Go standard library - time, strings, regexp

### Design Approach

Layered architecture with clear separation of concerns:

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

### Component Interaction

Data flow from user input to filtered results:

```
User Input (Ctrl+G)
    |
    v
Raw query string (e.g., "size > 1GiB")
    |
    v
Lexer: Tokenize into token stream
    |
    v
Parser: Build Abstract Syntax Tree (AST)
    |
    v
Evaluator: For each FileEntry, evaluate AST -> bool
    |
    v
Filtered file list
```

## Implementation Phases

### Phase 1: Lexer Implementation

**Goal**: Tokenize SQL-like query strings into a stream of typed tokens

**Files to Create**:
- `internal/filter/token.go` - Token type definitions and constants
- `internal/filter/lexer.go` - Lexer implementation
- `internal/filter/lexer_test.go` - Comprehensive lexer tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TokenType | Define all token categories | None | Enumeration of all valid token types |
| Token | Hold token value, type, and position | None | Immutable token representation |
| Lexer | Convert input string to token stream | Valid UTF-8 input | Token slice or error with position |

**Processing Flow**:
```
1. Initialize lexer with input string
   |-- Set position to 0
   |-- Prepare token accumulator
2. While not at end of input
   |-- Skip whitespace
   |-- Identify token type based on current character
   |   |-- Letter -> identifier or keyword
   |   |-- Digit -> number (with optional size/duration suffix)
   |   |-- Quote -> string literal
   |   |-- Symbol -> operator or punctuation
   |-- Consume token characters
   |-- Record token with position
3. Append EOF token
4. Return token stream
```

**Implementation Steps**:

1. **Define Token Types**
   - Define enumeration for all token categories
   - Include: identifiers, keywords (AND, OR, NOT, LIKE, etc.), operators, literals, punctuation

2. **Implement Core Lexer**
   - Character-by-character scanning with position tracking
   - Whitespace handling

3. **Implement Token Recognition**
   - Identifiers and keywords (case-insensitive keyword matching)
   - String literals with single quotes (support `''` escape for embedded single quotes)
   - Numbers with optional size units (KiB, MiB, GB, etc.)
   - Duration suffixes (d, h, m)
   - Operators (=, !=, <>, <, >, <=, >=)
   - Parentheses and comma

4. **Implement Error Reporting**
   - Track position for error messages
   - Report unterminated strings, invalid characters

**Dependencies**:
- Requires: None (foundation layer)
- Blocks: Phase 2 (Parser)

**Testing Approach**:

*Unit Tests*:
- Token recognition for each token type
- Position tracking accuracy
- Error cases (unterminated strings, invalid tokens)
- Edge cases (empty input, whitespace-only)

| Test Category | Example Cases |
|---------------|---------------|
| Simple tokens | `size`, `>`, `100` |
| Size units | `1GiB`, `100MB`, `1KiB` |
| Durations | `7d`, `24h`, `30m` |
| String literals | `'%.txt'`, `'test'` |
| Escaped quotes | `'file''s name.txt'` -> `file's name.txt` |
| Operators | `=`, `!=`, `<>`, `LIKE`, `ILIKE` |
| Keywords | `AND`, `OR`, `NOT`, `IN`, `IS`, `NULL` |
| Complex | `size > 1GiB AND name LIKE '%.txt'` |
| Errors | `'unterminated`, `'file's name'` (unescaped quote), `@invalid` |

**Acceptance Criteria**:
- [ ] All token types correctly recognized
- [ ] Position tracking accurate for error reporting
- [ ] Size units (KiB, MiB, GiB, TiB, KB, MB, GB, TB) correctly lexed
- [ ] Duration literals (Nd, Nh, Nm) correctly lexed
- [ ] String literals with single quotes handled
- [ ] Escaped single quotes (`''`) correctly processed as single `'`
- [ ] Keywords case-insensitive
- [ ] Error messages include position

**Estimated Effort**: Medium (3-4 days)

**Risks and Mitigation**:
- **Risk**: Unicode handling in identifiers
  - **Mitigation**: Restrict identifiers to ASCII for column names

---

### Phase 2: Parser Implementation

**Goal**: Parse token stream into Abstract Syntax Tree (AST) following SQL operator precedence

**Files to Create**:
- `internal/filter/ast.go` - AST node type definitions
- `internal/filter/parser.go` - Recursive descent parser
- `internal/filter/parser_test.go` - Parser tests
- `internal/filter/errors.go` - Error types with position and hints

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Expression | Interface for all AST nodes | None | Common evaluation contract |
| BinaryExpr | Represent binary operations (AND, OR, comparisons) | Two child expressions | Evaluatable binary expression |
| UnaryExpr | Represent unary operations (NOT) | One child expression | Evaluatable unary expression |
| ColumnRef | Reference to a file attribute column | Valid column name | Column accessor |
| Literal | Constant value (string, number, timestamp) | None | Typed constant value |
| FunctionCall | Function invocation (now, year, lower, etc.) | Valid function name, args | Callable function node |
| InExpr | IN/NOT IN list membership | Column and value list | List membership test |
| IsNullExpr | IS NULL/IS NOT NULL check | Column reference | Null check node |
| Parser | Convert token stream to AST | Valid token stream | AST root or error |
| QueryError | Structured error with position and hint | Error condition | User-friendly error message |

**Processing Flow** (Grammar-based):
```
1. Parse query
   |-- Skip optional WHERE keyword
   |-- Parse expression
2. Parse expression (OR level - lowest precedence)
   |-- Parse AND expression
   |-- While OR token, parse next AND expression
3. Parse AND expression
   |-- Parse NOT expression
   |-- While AND token, parse next NOT expression
4. Parse NOT expression
   |-- If NOT token, parse primary and negate
   |-- Else parse primary
5. Parse primary
   |-- If left paren, parse grouped expression
   |-- Else parse comparison or special form
6. Parse comparison
   |-- Parse operand (column, literal, function)
   |-- If comparison operator, parse right operand
   |-- If LIKE/ILIKE/NOT LIKE, parse pattern
   |-- If IN/NOT IN, parse value list
   |-- If IS NULL/IS NOT NULL, create null check
   |-- If boolean column alone, return as condition
```

**Implementation Steps**:

1. **Define AST Node Types**
   - Define Expression interface with evaluation contract
   - Implement concrete node types for each expression kind

2. **Implement Recursive Descent Parser**
   - Entry point handling optional WHERE keyword
   - Expression parsing following precedence: NOT > AND > OR

3. **Implement Comparison Parsing**
   - Comparison operators (=, !=, <>, <, >, <=, >=)
   - LIKE/ILIKE/NOT LIKE with pattern validation
   - IN/NOT IN with value list parsing
   - IS NULL/IS NOT NULL

4. **Implement Operand Parsing**
   - Column references with validation
   - Literals (strings, numbers, dates)
   - Function calls with argument parsing
   - Parenthesized expressions

5. **Implement Error Handling**
   - Create QueryError type with position, message, hint
   - Generate helpful hints for common mistakes

**Dependencies**:
- Requires: Phase 1 (Lexer)
- Blocks: Phase 3 (Evaluator)

**Testing Approach**:

*Unit Tests*:
- Parse each expression type in isolation
- Verify AST structure correctness
- Test operator precedence
- Test error reporting

| Test Category | Example Cases |
|---------------|---------------|
| Simple comparison | `size > 100` |
| AND expression | `size > 100 AND type = 'file'` |
| OR expression | `ext = 'txt' OR ext = 'md'` |
| NOT expression | `NOT isdir` |
| Parentheses | `(size > 1GiB OR name LIKE '%.mp4') AND NOT isdir` |
| LIKE patterns | `name LIKE '%.txt'`, `name ILIKE '%test%'` |
| IN lists | `ext IN ('jpg', 'png', 'gif')` |
| IS NULL | `ext IS NULL`, `ext IS NOT NULL` |
| Functions | `year(mtime) = 2024`, `now() - 7d` |
| Errors | Missing operand, unbalanced parens, unknown column |

**Acceptance Criteria**:
- [ ] All comparison operators parsed correctly
- [ ] Operator precedence follows SQL standard (NOT > AND > OR)
- [ ] Parentheses grouping works correctly
- [ ] LIKE/ILIKE patterns accepted
- [ ] IN lists parsed correctly
- [ ] IS NULL/IS NOT NULL parsed
- [ ] Function calls with arguments parsed
- [ ] Error messages include position and helpful hints
- [ ] Unknown columns reported with valid column list

**Estimated Effort**: Medium (3-4 days)

**Risks and Mitigation**:
- **Risk**: Complex grammar leading to parser bugs
  - **Mitigation**: Extensive test coverage, follow EBNF grammar strictly
- **Risk**: Poor error messages
  - **Mitigation**: Design error types upfront, test error scenarios

---

### Phase 3: Evaluator Implementation

**Goal**: Evaluate AST against FileEntry to produce boolean result

**Files to Create**:
- `internal/filter/evaluator.go` - Expression evaluation logic
- `internal/filter/evaluator_test.go` - Evaluator tests
- `internal/filter/functions.go` - Built-in function implementations
- `internal/filter/functions_test.go` - Function tests
- `internal/filter/coercion.go` - Type coercion utilities

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Evaluator | Evaluate AST against FileEntry | Valid AST, FileEntry | Boolean result or error |
| ColumnAccessor | Extract column value from FileEntry | Valid column name | Typed value (string, int64, time.Time, bool) |
| TypeCoercer | Convert values between types | Source value | Target type value or error |
| PatternMatcher | Evaluate LIKE/ILIKE patterns | Pattern string, target string | Boolean match result |
| FunctionRegistry | Execute built-in functions | Function name, arguments | Function result |

**Type Coercion Rules**:
- Type mismatch in comparison: Return error (e.g., `size = 'abc'`)
- NULL comparison: Always return false (except IS NULL/IS NOT NULL)
- Implicit string-to-number: Not supported, explicit format required
- Timestamp comparison: String literals auto-parsed as ISO 8601
- Timezone handling: All date/time literals interpreted in local timezone (system default)
- now() function: Returns current local time

**Processing Flow**:
```
1. Receive AST and FileEntry
2. Evaluate expression recursively
   |-- BinaryExpr: Evaluate left, right, apply operator
   |   |-- AND: Short-circuit on first false
   |   |-- OR: Short-circuit on first true
   |   |-- Comparison: Coerce types, compare values
   |-- UnaryExpr (NOT): Evaluate operand, negate
   |-- ColumnRef: Extract value from FileEntry
   |-- Literal: Return constant value
   |-- FunctionCall: Evaluate args, execute function
   |-- InExpr: Check value membership in list
   |-- IsNullExpr: Check if column value is null
3. Return boolean result
```

**Implementation Steps**:

1. **Implement Column Value Extraction**
   - Map column names to FileEntry fields
   - Derive computed columns (type, ext, isfile)
   - Handle NULL for ext column

2. **Implement Comparison Evaluation**
   - Type-aware comparison (numbers, strings, timestamps)
   - NULL handling (comparisons with NULL return false)

3. **Implement Pattern Matching**
   - Convert SQL LIKE pattern to Go regex
   - Handle % (any string) and _ (single char) wildcards
   - Case-sensitive (LIKE) and case-insensitive (ILIKE) variants

4. **Implement Type Coercion**
   - Size literal to bytes (handle KiB/MiB/GiB/TiB, KB/MB/GB/TB)
   - Duration to time offset (d, h, m)
   - String to timestamp (ISO 8601)

5. **Implement Built-in Functions**
   - now() - current timestamp
   - year(timestamp), month(timestamp), day(timestamp) - extract date parts
   - lower(string), upper(string) - case conversion

6. **Implement Logical Operators**
   - AND with short-circuit evaluation
   - OR with short-circuit evaluation
   - NOT negation

**Dependencies**:
- Requires: Phase 2 (Parser)
- Blocks: Phase 4 (UI Integration)

**Testing Approach**:

*Unit Tests*:
- Column value extraction for all columns
- Comparison operators for each type
- Pattern matching with various wildcards
- Function execution with edge cases
- Type coercion accuracy

| Test Category | Example Cases |
|---------------|---------------|
| Size comparisons | `size > 1073741824` (1GiB), `size > 1000000000` (1GB) |
| String equality | `name = 'test.txt'`, `type = 'dir'` |
| Pattern matching | `name LIKE '%.txt'`, `name LIKE 'report_%'`, `name LIKE '___.go'` |
| Case-insensitive | `name ILIKE '%TEST%'` |
| Date comparison | `mtime > '2024-01-01'`, `mtime > now() - 7d` |
| Date functions | `year(mtime) = 2024`, `month(mtime) = 12` |
| String functions | `lower(name) = 'readme.md'`, `upper(ext) = 'TXT'` |
| Boolean columns | `isdir`, `NOT isfile`, `issymlink` |
| IN operator | `ext IN ('jpg', 'png')` |
| NULL handling | `ext IS NULL`, `ext IS NOT NULL` |
| Logical combinations | `size > 1GiB AND year(mtime) = 2024` |

**Acceptance Criteria**:
- [ ] All columns correctly extracted from FileEntry
- [ ] Size units converted accurately (binary and decimal)
- [ ] LIKE patterns with % and _ work correctly
- [ ] ILIKE performs case-insensitive matching
- [ ] Date literals parsed and compared correctly
- [ ] now() returns current time
- [ ] Date extraction functions work correctly
- [ ] NULL comparisons return false
- [ ] IS NULL/IS NOT NULL work for ext column
- [ ] Short-circuit evaluation for AND/OR

**Estimated Effort**: Medium (2-3 days)

**Risks and Mitigation**:
- **Risk**: Timezone issues with date comparisons
  - **Mitigation**: Use local timezone consistently, document behavior
- **Risk**: Performance with complex patterns
  - **Mitigation**: Benchmark pattern matching, consider caching compiled regex

---

### Phase 4: UI Integration

**Goal**: Integrate SQL-like filter into duofm UI with Ctrl+G keybinding

**Files to Create**:
- `internal/filter/filter.go` - Main filter function (public API)
- `internal/filter/filter_test.go` - Integration tests

**Files to Modify**:
- `internal/ui/search.go`:
  - Add SearchModeSQLLike constant
  - Add String() case for new mode
- `internal/ui/model_update.go`:
  - Add Ctrl+G key handler
  - Handle Enter to execute filter
  - Handle error display
- `internal/ui/pane.go` or `internal/ui/pane_filter.go`:
  - Call FilterSQLLike when SearchModeSQLLike is active
- `internal/ui/help_dialog.go`:
  - Add Ctrl+G to Display & Search section
  - Add SQL-like Filter section with documentation

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| SearchModeSQLLike | New search mode constant | None | Integrated with existing modes |
| FilterSQLLike | Public filter function | Entries, query string | Filtered entries or error |
| CompiledQuery | Cached parsed query | Valid query string | Reusable filter predicate |
| UI Error Display | Show syntax errors | QueryError | Error shown to user |

**Processing Flow**:
```
1. User presses Ctrl+G
   |-- Set search mode to SearchModeSQLLike
   |-- Show minibuffer with "WHERE " prompt
2. User types query
   |-- Standard minibuffer editing (no real-time filtering)
3. User presses Enter
   |-- Parse and evaluate query
   |-- If error, display error message
   |-- If success, filter entries and display result
4. User presses Esc
   |-- Cancel filter mode
   |-- Restore original entry list
```

**Implementation Steps**:

1. **Create Filter Public API**
   - FilterSQLLike function accepting entries and query
   - CompileQuery for query validation
   - CompiledQuery.Match for per-entry filtering

2. **Add SearchModeSQLLike**
   - Add constant to SearchMode enum
   - Update String() method

3. **Add Ctrl+G Key Handler**
   - Activate SQL-like filter mode
   - Set minibuffer prompt to "WHERE "
   - Handle Enter to execute
   - Handle Esc to cancel

4. **Integrate with Pane Filtering**
   - Call FilterSQLLike when mode is active
   - Display filtered results
   - Show error if parsing fails

5. **Update Help Dialog**
   - Add Ctrl+G to keybindings section
   - Add SQL-like Filter reference section

**Dependencies**:
- Requires: Phase 3 (Evaluator)
- Blocks: Phase 5 (Polish)

**Testing Approach**:

*Unit Tests*:
- FilterSQLLike function with various queries
- Error handling for invalid queries

*Integration Tests*:
- End-to-end filtering through UI
- Key handling (Ctrl+G, Enter, Esc)
- Error display

| Test Category | Example Cases |
|---------------|---------------|
| Filter function | Filter by size, name, type combinations |
| Empty query | Clear filter when query is empty |
| Error handling | Invalid syntax shows error message |
| UI flow | Ctrl+G opens, Enter executes, Esc cancels |

**Acceptance Criteria**:
- [ ] Ctrl+G activates SQL-like filter
- [ ] Minibuffer shows "WHERE " prompt
- [ ] Enter executes filter and displays results
- [ ] Esc cancels filter and restores list
- [ ] Error messages displayed clearly
- [ ] Empty query clears filter
- [ ] Help dialog updated with SQL-like filter documentation

**Estimated Effort**: Small (2 days)

**Risks and Mitigation**:
- **Risk**: Conflict with existing keybindings
  - **Mitigation**: Ctrl+G is unused, verify no conflicts
- **Risk**: Error display not visible
  - **Mitigation**: Use existing error display mechanism or status bar

---

### Phase 5: Polish and Testing

**Goal**: Complete testing, performance optimization, and documentation

**Files to Create/Modify**:
- `internal/filter/benchmark_test.go` - Performance benchmarks
- Additional edge case tests across all test files

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Benchmarks | Measure parsing and filtering performance | All components complete | Performance metrics |
| Edge Case Tests | Cover unusual inputs | All components complete | Comprehensive coverage |
| Documentation | Inline code documentation | All components complete | Well-documented code |

**Processing Flow**:
```
1. Run existing tests to verify baseline
2. Add edge case tests
   |-- Unicode file names
   |-- Very long queries
   |-- Deeply nested expressions
   |-- Boundary values
3. Run benchmarks
   |-- Query parsing time
   |-- Filter execution time (10K, 100K files)
4. Optimize if needed
   |-- Cache compiled queries
   |-- Optimize pattern matching
5. Verify NFR compliance
   |-- Parsing < 10ms
   |-- 10K files < 100ms
```

**Implementation Steps**:

1. **Add Edge Case Tests**
   - Unicode in file names and patterns
   - Very long queries (>1000 chars)
   - Size of 0 bytes
   - Files with no extension
   - Hidden files (starting with .)
   - Symbolic links
   - Deeply nested parentheses
   - Many OR conditions

2. **Implement Benchmarks**
   - Query parsing benchmark
   - Filter 10,000 files benchmark
   - Filter 100,000 files benchmark

3. **Performance Optimization**
   - Compile query once, reuse for all entries
   - Cache now() value per filter operation
   - Short-circuit logical operators

4. **Verify Test Coverage**
   - Target: >90% for parser, >80% for evaluator
   - Run coverage report

5. **Final Documentation**
   - Inline code comments
   - Package documentation

**Dependencies**:
- Requires: Phases 1-4 complete

**Testing Approach**:

*Benchmark Tests*:
- Parse simple query
- Parse complex query
- Filter 10K entries
- Filter 100K entries

*Edge Case Tests*:
| Test Category | Example Cases |
|---------------|---------------|
| Empty/whitespace | `""`, `"   "` |
| Unicode | File name with Japanese, emoji |
| Long query | >1000 character query |
| Deep nesting | `((((a AND b) OR c) AND d) OR e)` |
| Many conditions | `a OR b OR c OR ... OR z` |
| Boundary | `size = 0`, `mtime = '1970-01-01'` |

**Acceptance Criteria**:
- [ ] Query parsing < 10ms
- [ ] Filter 10,000 files < 100ms
- [ ] Filter 100,000 files < 1s
- [ ] Parser test coverage > 90%
- [ ] All edge cases handled gracefully
- [ ] No panics or crashes with invalid input

**Estimated Effort**: Small (2 days)

**Risks and Mitigation**:
- **Risk**: Performance does not meet targets
  - **Mitigation**: Profile and optimize hot paths, consider algorithm changes

---

## Complete File Structure

```
internal/
+-- filter/
|   +-- token.go              # Token type definitions
|   +-- lexer.go              # Tokenizer implementation
|   +-- lexer_test.go         # Lexer tests
|   +-- ast.go                # AST node definitions
|   +-- parser.go             # Recursive descent parser
|   +-- parser_test.go        # Parser tests
|   +-- evaluator.go          # Expression evaluator
|   +-- evaluator_test.go     # Evaluator tests
|   +-- functions.go          # Built-in functions (now, year, lower, etc.)
|   +-- functions_test.go     # Function tests
|   +-- coercion.go           # Type coercion utilities
|   +-- errors.go             # Error types (QueryError)
|   +-- filter.go             # Main filter function (public API)
|   +-- filter_test.go        # Integration tests
|   +-- benchmark_test.go     # Performance benchmarks
+-- ui/
    +-- search.go             # Modify: Add SearchModeSQLLike
    +-- model_update.go       # Modify: Add Ctrl+G handler
    +-- pane.go               # Modify: Add FilterSQLLike call (or pane_filter.go)
    +-- help_dialog.go        # Modify: Add SQL-like filter documentation
```

**File Descriptions**:

| File | Responsibility |
|------|----------------|
| token.go | Define TokenType enum and Token struct |
| lexer.go | Convert input string to token stream |
| ast.go | Define Expression interface and node types |
| parser.go | Parse tokens into AST following SQL grammar |
| evaluator.go | Evaluate AST against FileEntry |
| functions.go | Implement now(), year(), month(), day(), lower(), upper() |
| coercion.go | Type conversion for sizes, durations, timestamps |
| errors.go | QueryError with position, message, hint |
| filter.go | Public API: FilterSQLLike(), CompileQuery() |
| search.go | SearchMode enum with SQLLike variant |
| model_update.go | Ctrl+G key handling, filter execution |
| help_dialog.go | Help content with SQL-like filter reference |

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Mock FileEntry for evaluator tests

**Test Coverage Goals**:
- Parser/Lexer: >90%
- Evaluator: >80%
- Functions: >80%
- Overall: >85%

**Key Test Areas**:

1. **Lexer** (`internal/filter/`)
   - All token types recognized
   - Size unit parsing
   - Duration parsing
   - Error positions

2. **Parser** (`internal/filter/`)
   - Operator precedence
   - All expression types
   - Error messages with hints

3. **Evaluator** (`internal/filter/`)
   - Column extraction
   - Type coercion
   - Pattern matching
   - NULL handling

4. **Functions** (`internal/filter/`)
   - Date/time functions
   - String functions
   - Edge cases

### Integration Testing

**Scenarios**:
1. End-to-end query execution through FilterSQLLike
2. UI key handling flow
3. Error propagation

**Approach**:
- Create test FileEntry fixtures
- Verify filter results match expectations
- Test error display

### Performance Testing

**Benchmarks**:
- Query parsing
- Filter 10K entries
- Filter 100K entries
- Memory allocation

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Press Ctrl+G, type `size > 0`, press Enter
- [ ] Verify only non-empty files shown
- [ ] Press Ctrl+G, type invalid query, verify error displayed
- [ ] Press Esc to cancel filter
- [ ] Test with actual large directory (>1000 files)
- [ ] Verify help screen shows SQL-like filter section

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/charmbracelet/bubbletea | (existing) | TUI framework |
| github.com/charmbracelet/lipgloss | (existing) | Styling |

No new external dependencies required.

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: Lexer (no dependencies)
2. Phase 2: Parser (depends on Lexer)
3. Phase 3: Evaluator (depends on Parser)
4. Phase 4: UI Integration (depends on Evaluator)
5. Phase 5: Polish (depends on all phases)

**Component Dependencies**:
- `filter/parser.go` depends on `filter/lexer.go`, `filter/ast.go`
- `filter/evaluator.go` depends on `filter/ast.go`, `filter/functions.go`
- `filter/filter.go` depends on `filter/parser.go`, `filter/evaluator.go`
- `ui/model_update.go` depends on `filter/filter.go`

## Risk Assessment

### Technical Risks

1. **Parser Complexity**
   - **Risk**: Grammar edge cases cause bugs
   - **Likelihood**: Medium
   - **Impact**: Medium
   - **Mitigation**: Follow EBNF strictly, extensive testing

2. **Performance with Large Directories**
   - **Risk**: Filtering >10K files takes too long
   - **Likelihood**: Low
   - **Impact**: Medium
   - **Mitigation**: Compile query once, short-circuit evaluation

3. **Date/Time Edge Cases**
   - **Risk**: Timezone and leap year issues
   - **Likelihood**: Medium
   - **Impact**: Low
   - **Mitigation**: Use local timezone, test edge cases

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Adding features beyond spec (BETWEEN, GLOB, etc.)
   - **Mitigation**: Strict adherence to spec, defer to Open Questions

2. **Integration Conflicts**
   - **Risk**: Changes break existing search functionality
   - **Mitigation**: New search mode, existing modes unchanged

## Performance Considerations

1. **Query Parsing**
   - Parse once, evaluate many
   - Avoid re-parsing for each file

2. **Evaluation**
   - Short-circuit AND/OR
   - Cache now() per filter operation
   - Lazy column value extraction

3. **Pattern Matching**
   - Compile regex once per LIKE pattern
   - Consider caching compiled patterns

## Security Considerations

1. **Input Validation**
   - All user input validated by parser
   - No arbitrary code execution
   - Bounded recursion depth

2. **Path Safety**
   - Filter operates on FileEntry, not paths
   - No path traversal possible

3. **Resource Limits**
   - Query length limit (implicit in UI)
   - Stack depth limit in parser

## Open Questions

### From Specification:
- [ ] Should we support BETWEEN operator? (e.g., `size BETWEEN 1M AND 1G`)
- [ ] Should we support GLOB pattern matching in addition to LIKE?
- [ ] Should we add query history?
- [ ] Should we allow saving named filters?

### Implementation-Specific:
- [ ] Should pattern regex be cached across filter operations?
- [ ] How to handle very long error messages in minibuffer?

## Future Enhancements

Items deferred to later phases or releases (from spec Open Questions):
- Query history navigation
- Preset/saved filters
- BETWEEN operator
- GLOB pattern matching

## Success Metrics

### Functional Completeness
- [ ] All functional requirements (FR1-FR12) implemented
- [ ] All user stories (US1-US6) satisfied
- [ ] All acceptance criteria met

### Quality Metrics
- [ ] Parser test coverage > 90%
- [ ] No critical bugs in manual testing
- [ ] Code follows Go best practices

### Performance Metrics
- [ ] Query parsing < 10ms
- [ ] Filter 10,000 files < 100ms
- [ ] UI response < 50ms

### User Experience
- [ ] Error messages include position and hints
- [ ] Help screen documents all features
- [ ] Syntax familiar to SQL users

## References

- **Specification**: `doc/tasks/advanced-filtering/SPEC.md`
- **Requirements**: `doc/tasks/advanced-filtering/要件定義書.md`
- **FileEntry**: `internal/fs/types.go`
- **Search**: `internal/ui/search.go`
- **Minibuffer**: `internal/ui/minibuffer.go`
- **Help Dialog**: `internal/ui/help_dialog.go`
- **SQL LIKE**: https://www.postgresql.org/docs/current/functions-matching.html
- **IEC 80000-13**: https://en.wikipedia.org/wiki/Binary_prefix

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review plan against specification
   - Address open questions
   - Confirm approach

2. **Begin Implementation**
   - Start with Phase 1 (Lexer)
   - Follow TDD approach
   - Commit incrementally

3. **Verification**
   - Use VERIFICATION.md for testing checklist
   - Verify each phase before proceeding
