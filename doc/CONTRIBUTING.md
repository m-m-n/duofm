# Contributing to duofm

This guide provides detailed guidelines for contributing to duofm, including documentation standards, development workflow, and coding conventions.

## Documentation Guidelines

### Overview

duofm uses a structured documentation approach to maintain clarity and organization throughout the development process.

### Documentation Structure

```
doc/
├── specification/
│   └── SPEC.md              # Overall project specification
├── tasks/
│   ├── dual-pane-ui/
│   │   └── SPEC.md          # Feature-specific spec
│   ├── file-operations/
│   │   └── SPEC.md          # Feature-specific spec
│   └── ...
└── CONTRIBUTING.md          # This file
```

## Specification Documents

### Overall Specification: `doc/specification/SPEC.md`

The overall specification document describes the project-wide architecture and design decisions.

**Contents should include:**
- Project vision and goals
- High-level architecture
- Core concepts and terminology
- System-wide requirements
- Data models and structures
- Technology stack decisions
- Cross-cutting concerns (logging, error handling, etc.)

**When to update:**
- When adding major features that affect the overall architecture
- When making significant technology decisions
- When defining new core concepts
- During initial project planning

**Template:**
```markdown
# duofm - Overall Specification

## Vision
[Project vision and goals]

## Architecture
[High-level architecture diagram and explanation]

## Core Concepts
[Key concepts and terminology]

## Technology Stack
[Languages, frameworks, libraries with rationale]

## Data Models
[Core data structures]

## Cross-Cutting Concerns
[Logging, error handling, configuration, etc.]
```

### Feature Specifications: `doc/tasks/FEATURE_NAME/SPEC.md`

Feature specifications describe individual features in detail. **The specification is the Single Source of Truth (SSOT)** - all implementation plans and code are derivatives.

**SSOT Principles:**
- Specification describes WHAT and WHY, not HOW
- No implementation code or architecture decisions (those belong in IMPLEMENTATION.md)
- Changes flow: Specification → Implementation Plan → Code

**Contents should include (in priority order):**
1. Feature overview and objectives
2. Domain rules and business constraints (highest priority)
3. User stories or use cases
4. Functional requirements (numbered: FR1.1, FR1.2, ...)
5. Non-functional requirements (numbered: NFR1.1, NFR1.2, ...)
6. Interface contracts (input/output, preconditions/postconditions)
7. Test scenarios
8. Success criteria
9. Dependencies and constraints

**DO NOT include:**
- Implementation code or code snippets
- File/directory structure
- Architecture decisions
- Library-specific API calls

**When to create:**
- Before starting implementation of a significant feature
- When planning features that require design decisions
- When features need to be reviewed before implementation
- For features that involve multiple components

**Naming convention:**
- Use lowercase with hyphens: `dual-pane-ui`, `file-operations`, `keyboard-shortcuts`
- Be descriptive but concise
- Avoid version numbers in directory names

**Template:**
```markdown
# Feature: [Feature Name]

## Overview
[Brief description of the feature]

## Domain Rules
[Business rules and constraints - CRITICAL: highest priority]
- [Invariant 1]
- [Constraint 1]

## Objectives
- [Objective 1]
- [Objective 2]

## User Stories
- As a [user type], I want to [action], so that [benefit]

## Functional Requirements
- FR1.1: [Requirement - describe behavior, not implementation]
- FR1.2: [Requirement]

## Non-Functional Requirements
- NFR1.1: [Performance/security/usability constraint]

## Interface Contract

### Input/Output Specification
[Data types, formats, valid ranges]

### Preconditions/Postconditions
[State requirements before and after operations]

### State Transitions
[Use Mermaid diagrams for state machines]

### Error Conditions
[Expected error behaviors - describe at behavioral level]

## Dependencies
[External requirements, prerequisites]

## Test Scenarios
- [ ] Scenario 1: [description]
- [ ] Scenario 2: [description]

## Success Criteria
- [ ] Criteria 1
- [ ] Criteria 2

## Constraints
[Technical/business limitations]

## Open Questions
- [ ] Question 1
- [ ] Question 2
```

## Development Workflow

**SSOT Change Flow:** Specification → Implementation Plan → Code

Changes MUST flow from specification to implementation, never the reverse. If implementation reveals issues, update the specification FIRST, then update the plan and code.

### 1. Specification Phase (SSOT)
1. Create or update `doc/specification/SPEC.md` for architectural changes
2. Create `doc/tasks/FEATURE_NAME/SPEC.md` for new features
3. Review specifications with team (or self-review for solo projects)
4. **Specification is approved before proceeding**

### 2. Planning Phase
1. Create `doc/tasks/FEATURE_NAME/IMPLEMENTATION.md` based on specification
2. Verify plan covers all specification requirements
3. Plan includes file structure, architecture, and implementation steps

### 3. Implementation Phase
1. Create feature branch: `git checkout -b feature/FEATURE_NAME`
2. Implement according to plan (which derives from specification)
3. Write tests alongside code
4. If issues found: **Update specification FIRST**, then plan, then code

### 4. Review Phase
1. Self-review changes
2. Run all tests and linting
3. Verify implementation matches specification
4. Commit with descriptive messages

### 5. Integration Phase
1. Merge feature branch to main
2. Archive or update task specification
3. Update overall specification if needed

## Code Conventions

### Go Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use meaningful variable and function names
- Write godoc comments for exported functions

### File Organization

```
internal/
├── ui/
│   ├── pane.go          # UI component
│   ├── pane_test.go     # Unit tests
│   └── ...
├── fs/
│   ├── operations.go
│   ├── operations_test.go
│   └── ...
```

### Testing Guidelines

**Write tests for:**
- All business logic
- File system operations (use temporary directories)
- UI component logic (mock rendering where possible)

**Test structure:**
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Commit Messages

Use conventional commit format:
```
feat: add dual pane navigation
fix: correct file copy error handling
docs: update SPEC.md with cache strategy
test: add tests for file operations
refactor: simplify UI rendering logic
```

## Best Practices

### Documentation
- Keep specifications up-to-date with implementation
- Use diagrams where helpful (ASCII art is fine)
- Include rationale for important decisions
- Reference issues or discussions when applicable

### Code
- Keep functions small and focused
- Separate UI logic from business logic
- Handle errors explicitly
- Use context for cancellation
- Comment non-obvious code

### Testing
- Aim for high test coverage
- Test edge cases and error conditions
- Use table-driven tests for multiple scenarios
- Keep tests fast and independent

## Questions or Suggestions?

For questions about contributing or suggestions for improving these guidelines, please:
- Open an issue on the project repository
- Update this document directly (if you're a maintainer)
- Discuss in project communication channels
