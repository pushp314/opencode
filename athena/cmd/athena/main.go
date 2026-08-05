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
	if args[0] != "index" {
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
	return index(ctx, args[1:], stdout)
}

func index(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("repo", "", "Git repository to index")
	database := flags.String("database", defaultDatabasePath(), "SQLite database path")
	jsonOutput := flags.Bool("json", false, "write a JSON report")
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
	report, err := repository.NewIndexer(store).Index(ctx, repository.Ref{Root: *root})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return json.NewEncoder(stdout).Encode(report)
	}
	_, err = fmt.Fprintf(stdout, "snapshot=%s files=%d changed=%d skipped=%d reused=%t\n", report.SnapshotID, report.Files, report.Changed, report.Skipped, report.Reused)
	return err
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
  athena index [--repo PATH] [--database PATH] [--json]

The index command is read-only. It inventories Git-visible text files and stores
deterministic SHA-256 snapshots in local SQLite. ATHENA_DATA_DIR changes the
default database directory.
`)
}
