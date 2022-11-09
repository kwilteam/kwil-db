package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"ksl"
	"ksl/kslparse"
	"ksl/sqlclient"
	_ "ksl/sqldriver"
	"ksl/sqlspec"
)

func main() {
	client, err := sqlclient.Open(context.Background(), "postgres://localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to open client: %v", err)
	}
	defer client.Close()

	differ := sqlspec.NewDiffer()
	planner := sqlspec.NewPlanner()

	from := realmFromClient(client)
	// from := realmFromFile("data/from.kwil")
	to := realmFromFile("data/test.kwil")

	fmt.Println("from")
	sqlspec.Marshal(os.Stdout, from)
	fmt.Println("to")
	sqlspec.Marshal(os.Stdout, to)

	changes, err := differ.RealmDiff(from, to)
	if err != nil {
		log.Fatalf("failed to diff realms: %v", err)
	}

	if len(changes) == 0 {
		log.Println("Schema is synced, no changes to be made")
		return
	}

	plan, err := planner.PlanChanges(changes)
	if err != nil {
		log.Fatalf("failed to plan changes: %v", err)
	}

	sqlspec.PrintPlan(plan)
	// if err := client.ApplyChanges(context.Background(), changes); err != nil {
	// 	log.Fatalf("failed to apply changes: %v", err)
	// }
}

func realmFromClient(client *sqlclient.Client) *sqlspec.Realm {
	targetOpts := &sqlspec.InspectRealmOption{}
	if client.URL.Schema != "" {
		targetOpts.Schemas = append(targetOpts.Schemas, client.URL.Schema)
	}

	inspectedRealm, err := client.InspectRealm(context.Background(), targetOpts)
	if err != nil {
		log.Fatalf("failed to inspect target: %v", err)
	}
	return inspectedRealm
}

func realmFromFile(file string) *sqlspec.Realm {
	parser := kslparse.NewParser()
	_, diags := parser.ParseFile(file)

	if diags.HasErrors() {
		ksl.NewDiagnosticTextWriter(os.Stdout, parser.Sources(), 120, true).WriteDiagnostics(diags)
		os.Exit(1)
	}

	r, diags := sqlspec.Decode(parser.FileSet())
	if diags.HasErrors() {
		ksl.NewDiagnosticTextWriter(os.Stdout, parser.Sources(), 120, true).WriteDiagnostics(diags)
		os.Exit(1)
	}
	return r
}
