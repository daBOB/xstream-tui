# xstream-tui Documentation

Welcome to the xstream-tui documentation! This folder contains comprehensive guides for understanding, developing, and maintaining the IPTV terminal user interface application.

## Quick Navigation

### I just joined the project - where do I start?
1. Read **[Project Overview & PDR](./project-overview-pdr.md)** - Understand what we're building and why
2. Skim **[Codebase Summary](./codebase-summary.md)** - Get familiar with the current state
3. Review **[System Architecture](./system-architecture.md)** - Understand how components interact
4. Check **[Code Standards](./code-standards.md)** - Follow the established patterns

### I'm implementing a new feature
1. Check **[Code Standards](./code-standards.md)** - Naming, patterns, testing requirements
2. Review **[System Architecture](./system-architecture.md)** - Where does it fit?
3. Read relevant sections of **[Codebase Summary](./codebase-summary.md)**
4. Consult phase plans in `../plans/` for requirements

### I'm debugging a problem
1. Check **[System Architecture](./system-architecture.md)** - Data flow and error handling
2. Review the relevant component section in **[Codebase Summary](./codebase-summary.md)**
3. Check **[Code Standards](./code-standards.md)** - Error handling patterns
4. Search issue history in `../plans/reports/`

### I'm reviewing code
1. Use **[Code Standards](./code-standards.md)** - Code review checklist at the end
2. Reference **[System Architecture](./system-architecture.md)** for design patterns
3. Check **[Codebase Summary](./codebase-summary.md)** for component boundaries

---

## Documentation Files

### [codebase-summary.md](./codebase-summary.md)
**Current State of the Codebase**

Provides overview of the actual implementation:
- Project structure and directory layout
- Core components (entry point, TUI model)
- Dependencies with versions
- Build targets and development workflow
- Development patterns and conventions
- Phase status and roadmap

**Best for:** Understanding what exists right now, finding files, checking versions

**Key Sections:**
- Directory Structure
- Core Components
- Build & Development
- Phase Status

---

### [project-overview-pdr.md](./project-overview-pdr.md)
**Product Development Requirements**

Complete project specification and requirements:
- Project vision and goals
- 6 functional requirements (FR-1 to FR-6)
- 5 non-functional requirements (NFR-1 to NFR-5)
- Success metrics and acceptance criteria
- 6-phase implementation roadmap
- Risk assessment and mitigation strategies
- Dependency and constraint analysis

**Best for:** Understanding the big picture, what we're building, when we're done

**Key Sections:**
- Functional Requirements
- Non-Functional Requirements
- Implementation Phases
- Success Metrics
- Acceptance Criteria

---

### [code-standards.md](./code-standards.md)
**Development Guidelines & Coding Standards**

Rules and patterns that all code must follow:
- Naming conventions (files, types, functions, variables)
- Coding style and formatting
- Architectural patterns (Elm, Dependency Injection)
- Error handling strategies
- Testing standards and test patterns
- File size guidelines
- Pre-commit requirements
- Code review checklist

**Best for:** Writing code correctly, code review, understanding our patterns

**Key Sections:**
- Naming Conventions
- Code Standards
- Architecture Patterns
- Testing Standards
- Code Review Checklist

---

### [system-architecture.md](./system-architecture.md)
**Technical Architecture & Design**

Detailed description of how the system works:
- Three-layer architecture overview
- Layer descriptions and responsibilities
- Component details and interactions
- Data flow diagrams
- State machine visualization
- Message types and events
- Concurrency model
- Error handling strategy
- Performance and security considerations

**Best for:** Understanding how components interact, design decisions, data flow

**Key Sections:**
- Architecture Overview
- Layer Descriptions
- Data Flow
- State Machine
- Error Handling
- Performance Considerations

---

## Quick Reference

### Development Workflow
```
1. Read requirements in project-overview-pdr.md
2. Check architecture in system-architecture.md
3. Review code-standards.md before coding
4. Implement following naming/pattern conventions
5. Write tests (>70% coverage)
6. Run: make lint, make test
7. Get code reviewed using code-standards.md checklist
```

### File Organization
```
xstream-tui/
├── cmd/xstream-tui/main.go         # Entry point
├── internal/
│   ├── xc/                         # Data layer
│   ├── tui/                        # Presentation layer
│   ├── player/                     # Playback layer
│   └── config/                     # Configuration
├── Makefile                        # Build targets
├── go.mod / go.sum                 # Dependencies
└── docs/                           # This documentation
    ├── codebase-summary.md         # Current state
    ├── project-overview-pdr.md     # Requirements
    ├── code-standards.md           # Development rules
    ├── system-architecture.md      # Technical design
    └── README.md                   # This file
```

### Key Commands
```bash
make build           # Build executable
make run            # Run application
make test           # Run tests
make test-coverage  # Generate coverage report
make lint           # Run go vet
make clean          # Remove build artifacts
make tidy           # Tidy dependencies
```

### Key Technologies
- **Language:** Go 1.25.5
- **TUI Framework:** Bubble Tea v1.3.10
- **Styling:** Lipgloss v1.1.0
- **Components:** Bubbles v0.21.0
- **Configuration:** godotenv v1.5.1
- **Players:** mpv, VLC (external)

---

## Document Status

| Document | Status | Last Updated | Coverage |
|----------|--------|--------------|----------|
| codebase-summary.md | Complete | 2025-12-14 | Phase 1 |
| project-overview-pdr.md | Complete | 2025-12-14 | All phases |
| code-standards.md | Complete | 2025-12-14 | All code |
| system-architecture.md | Complete | 2025-12-14 | All phases |
| design-guidelines.md | Pending | - | Phase 3+ |
| deployment-guide.md | Pending | - | Phase 6 |
| project-roadmap.md | Pending | - | Phase 6 |

---

## How to Update Documentation

### When to Update
- After implementing a feature
- When architecture changes
- When requirements change
- After each phase completion
- When adding new patterns

### How to Update
1. Identify which document(s) are affected
2. Make changes while staying consistent with other docs
3. Update cross-references if needed
4. Add to the phase report in `../plans/reports/`
5. Commit with meaningful message

### Cross-Referencing
Link between documents using:
```markdown
[Document Name](./document-name.md)
[Section in Doc](./document-name.md#section-header)
```

---

## Contribution Guidelines

### Code Documentation
Every code change should consider:
- Package documentation (e.g., `// Package xc provides...`)
- Exported types/functions (e.g., `// NewClient creates...`)
- Complex logic (explain WHY in comments)
- Error types and handling

### Documentation Changes
When updating docs:
- Ensure accuracy against actual code
- Keep examples current and tested
- Maintain consistent terminology
- Update cross-references
- Note the update date and version

### Standards Compliance
All code must follow code-standards.md:
- Use correct naming conventions
- Follow file size guidelines
- Write tests (>70% coverage)
- Include error handling
- Pass code review checklist

---

## Common Questions

**Q: Where do I find the requirements?**
A: [project-overview-pdr.md](./project-overview-pdr.md) - Functional and non-functional requirements

**Q: How should I name my variables?**
A: [code-standards.md - Naming Conventions](./code-standards.md#naming-conventions)

**Q: What's the architecture?**
A: [system-architecture.md](./system-architecture.md) - Three-layer pattern with diagrams

**Q: How do I run tests?**
A: `make test` or `make test-coverage` - See [codebase-summary.md](./codebase-summary.md#build--development)

**Q: Where do tests go?**
A: See [code-standards.md - Test File Organization](./code-standards.md#test-file-organization)

**Q: How do I know when I'm done?**
A: Check [project-overview-pdr.md - Acceptance Criteria](./project-overview-pdr.md#acceptance-criteria-summary)

---

## Related Resources

### In This Repository
- Phase implementation plans: `../plans/251213-1633-iptv-tui-implementation/`
- Phase reports: `../plans/reports/`
- Research docs: `../plans/251213-1633-iptv-tui-implementation/research/`

### External References
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Guide](https://github.com/charmbracelet/lipgloss)
- [Go Standards](https://golang.org/doc/effective_go)
- [Xtream Codes API](https://github.com/E-Corp-4/Xtream-Codes)

---

## Document History

| Date | Author | Changes |
|------|--------|---------|
| 2025-12-14 | docs-manager | Initial creation - Phase 1 complete |

---

**Last Updated:** 2025-12-14
**Phase:** 1 (Project Setup) - Complete
**Status:** Ready for Phase 2
