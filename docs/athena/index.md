# Athena Migration Documentation Index

The documentation gate is complete when each dossier contains the current architecture, target architecture, comparison and risks, public interfaces, internal components, migration strategy, dependency graph, sequence diagram, data flow, acceptance criteria, testing strategy, benchmark strategy, and documentation obligations.

1. [Architecture Decision Record](architecture.md)
2. [Repository Intelligence](01-repository-intelligence.md)
3. [Knowledge System](02-knowledge-system.md)
4. [Context Engine](03-context-engine.md)
5. [Brain](04-brain.md)
6. [Execution](05-execution.md)
7. [Verification](06-verification.md)
8. [User Interfaces](07-user-interfaces.md)

The first implementation capability is repository inventory with durable file hashes, reviewed in [Repository Inventory Review](reviews/01-repository-inventory.md). The second is Tree-sitter symbol extraction with durable parse/symbol facts, reviewed in [Tree-sitter Symbol Facts Review](reviews/02-tree-sitter-symbol-facts.md). Both are intentionally read-only and do not replace an OpenCode runtime subsystem.
