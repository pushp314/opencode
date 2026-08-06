# TECH STACK & DEVELOPMENT RULES

Before starting any task, assume the following technology stack unless I explicitly override it.

Backend & Core
- Python 3.13+
- FastAPI for backend services, APIs, and internal agent services when required
- Typer for the CLI
- Pydantic v2 for validation
- AsyncIO as the default concurrency model

Databases
- PostgreSQL for structured data
- Qdrant for vector embeddings and semantic search
- Neo4j for knowledge graphs, repository relationships, and dependency graphs
- Redis for caching, queues, and session/state management
- DuckDB or SQLite for local analytics, logs, and lightweight storage

AI & LLM
- Local LLMs via Ollama or vLLM
- Hugging Face Transformers when direct model loading is needed
- Sentence Transformers for embeddings
- Hybrid RAG (Vector + BM25 + Graph Retrieval)
- MCP (Model Context Protocol) support from the beginning

Code Intelligence
- Tree-sitter for parsing source code
- AST-based code analysis
- GitPython for repository analysis
- NetworkX for graph algorithms when Neo4j is unnecessary

Task Execution
- Docker for isolated execution
- Pytest for testing
- Ruff for linting
- MyPy for type checking
- Pre-commit hooks
- Git-based version tracking

Architecture
- Modular, plugin-based architecture
- SOLID principles
- Clean Architecture
- Dependency Injection where appropriate
- Async-first design
- Extensible interfaces over tightly coupled implementations

Development Rules
- Reuse existing code whenever possible.
- Never introduce unnecessary dependencies.
- Prefer production-grade solutions over quick fixes.
- Every implementation should be scalable, maintainable, testable, and extensible.
- Explain architectural decisions before implementing significant changes.
- Write clean, typed, documented code following modern Python best practices.


FOR PERFORMANCE-CRITICAL COMPONENTS USE GO LANGUAGE