package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pushp314/opencode/athena/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "athena:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printUsage(stdout)
		return nil
	}
	switch args[0] {
	case "index":
		return index(ctx, args[1:], stdout)
	case "symbols":
		return symbols(ctx, args[1:], stdout)
	case "imports":
		return imports(ctx, args[1:], stdout)
	case "edges":
		return edges(ctx, args[1:], stdout)
	}
	printUsage(stderr)
	return fmt.Errorf("unknown command %q", args[0])
}

func index(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("repo", "", "Git repository to index")
	database := flags.String("database", defaultDatabasePath(), "SQLite database path")
	jsonOutput := flags.Bool("json", false, "write a JSON report")
	parseFacts := flags.Bool("parse", false, "extract and store symbol facts")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse index flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("index does not accept positional arguments")
	}
	if *root == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("read working directory: %w", err)
		}
		*root = workingDirectory
	}
	if err := os.MkdirAll(filepath.Dir(*database), 0o700); err != nil {
		return fmt.Errorf("create Athena data directory: %w", err)
	}
	store, err := repository.OpenSQLite(*database)
	if err != nil {
		return err
	}
	defer store.Close()
	report, err := repository.NewIndexer(store).Index(ctx, repository.Ref{Root: *root}, repository.IndexOptions{Parse: *parseFacts})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return json.NewEncoder(stdout).Encode(report)
	}
	_, err = fmt.Fprintf(stdout, "snapshot=%s files=%d changed=%d skipped=%d reused=%t parsed=%d parse_errors=%d symbols=%d\n",
		report.SnapshotID, report.Files, report.Changed, report.Skipped, report.Reused, report.Parsed, report.ParseErrors, report.Symbols)
	return err
}

func symbols(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("symbols", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("repo", "", "Git repository to query")
	database := flags.String("database", defaultDatabasePath(), "SQLite database path")
	jsonOutput := flags.Bool("json", false, "write a JSON array")
	name := flags.String("name", "", "filter by symbol name")
	kind := flags.String("kind", "", "filter by symbol kind")
	path := flags.String("path", "", "filter by file path")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse symbols flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("symbols does not accept positional arguments")
	}
	if *root == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("read working directory: %w", err)
		}
		*root = workingDirectory
	}
	repositoryID, err := repository.RepositoryID(*root)
	if err != nil {
		return err
	}
	store, err := repository.OpenSQLite(*database)
	if err != nil {
		return err
	}
	defer store.Close()
	query := repository.NewQuery(store)
	result, err := query.Symbols(ctx, repositoryID, repository.SymbolQuery{Name: *name, Kind: *kind, Path: *path})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return json.NewEncoder(stdout).Encode(result)
	}
	for _, symbol := range result {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\t%s\t%d-%d\t%d-%d\n", symbol.Path, symbol.Name, symbol.Kind,
			symbol.StartByte, symbol.EndByte, symbol.StartLine, symbol.EndLine); err != nil {
			return err
		}
	}
	return nil
}

func imports(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("imports", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("repo", "", "Git repository to query")
	database := flags.String("database", defaultDatabasePath(), "SQLite database path")
	jsonOutput := flags.Bool("json", false, "write a JSON array")
	kind := flags.String("kind", "", "filter by import kind")
	spec := flags.String("spec", "", "filter by module specifier")
	path := flags.String("path", "", "filter by file path")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse imports flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("imports does not accept positional arguments")
	}
	if *root == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("read working directory: %w", err)
		}
		*root = workingDirectory
	}
	repositoryID, err := repository.RepositoryID(*root)
	if err != nil {
		return err
	}
	store, err := repository.OpenSQLite(*database)
	if err != nil {
		return err
	}
	defer store.Close()
	result, err := repository.NewQuery(store).Imports(ctx, repositoryID, repository.ImportQuery{Kind: *kind, Spec: *spec, Path: *path})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return json.NewEncoder(stdout).Encode(result)
	}
	for _, imp := range result {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\t%s\t%d-%d\t%d-%d\n", imp.Path, imp.Spec, imp.Kind,
			imp.StartByte, imp.EndByte, imp.StartLine, imp.EndLine); err != nil {
			return err
		}
	}
	return nil
}

func edges(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("edges", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("repo", "", "Git repository to query")
	database := flags.String("database", defaultDatabasePath(), "SQLite database path")
	jsonOutput := flags.Bool("json", false, "write a JSON array")
	path := flags.String("path", "", "filter by source file path")
	kind := flags.String("kind", "", "filter by import kind")
	to := flags.String("to", "", "filter by target file path")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse edges flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("edges does not accept positional arguments")
	}
	if *root == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("read working directory: %w", err)
		}
		*root = workingDirectory
	}
	repositoryID, err := repository.RepositoryID(*root)
	if err != nil {
		return err
	}
	store, err := repository.OpenSQLite(*database)
	if err != nil {
		return err
	}
	defer store.Close()
	result, err := repository.NewQuery(store).Edges(ctx, repositoryID, repository.EdgeQuery{Path: *path, Kind: *kind, To: *to})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return json.NewEncoder(stdout).Encode(result)
	}
	for _, edge := range result {
		if _, err := fmt.Fprintf(stdout, "%s --%s(%d) (%s)--> %s\n", edge.SourcePath, edge.Kind, edge.StartByte, edge.Spec, edge.TargetPath); err != nil {
			return err
		}
	}
	return nil
}

func defaultDatabasePath() string {
	if configured := os.Getenv("ATHENA_DATA_DIR"); configured != "" {
		return filepath.Join(configured, "athena.db")
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		return "athena.db"
	}
	return filepath.Join(directory, "Athena", "athena.db")
}

func printUsage(writer io.Writer) {
	fmt.Fprint(writer, `Athena Engineering Brain foundation

Usage:
  athena index [--repo PATH] [--database PATH] [--json] [--parse]
  athena symbols [--repo PATH] [--database PATH] [--json] [--name NAME] [--kind KIND] [--path PATH]
  athena imports [--repo PATH] [--database PATH] [--json] [--kind KIND] [--spec SPEC] [--path PATH]
  athena edges [--repo PATH] [--database PATH] [--json] [--path PATH] [--kind KIND] [--to PATH]

The index command is read-only. It inventories Git-visible text files and stores
deterministic SHA-256 snapshots in local SQLite. With --parse it also extracts
and stores symbol and import facts for supported languages, and resolves
cross-file import edges for relative module and include references; unchanged
content is never re-parsed. ATHENA_DATA_DIR changes the default database dir.

The symbols command prints the durable symbols of the latest snapshot. Symbol
kinds: function, method, class, struct, interface, enum, type, const, var.

The imports command prints the durable module/header references of the latest
snapshot. Import kinds: import, export, include, require, use.

The edges command prints resolved cross-file import edges of the latest
snapshot. Edges only exist when the specifier resolves to a file in the
repository; external references are not graph edges.
Supported parse languages: Go, TypeScript/TSX, JavaScript, Python, Rust, Java,
C, and C++.
`)
}
