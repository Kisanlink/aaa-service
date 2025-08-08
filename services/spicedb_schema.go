package services

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	authzed "github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ensureBaseSchema makes sure a minimal, extensible schema exists in SpiceDB.
// If no schema is present, it writes the base schema.
// If some expected namespaces are missing, it appends them and writes the merged schema.
func ensureBaseSchema(ctx context.Context, client *authzed.Client, logger *zap.Logger) error {
	// Read current schema
	current := ""
	readResp, err := client.ReadSchema(ctx, &authzedpb.ReadSchemaRequest{})
	if err == nil && readResp != nil {
		current = readResp.SchemaText
	}

	// Build the desired base schema
	desired := buildBaseSchema()

	// If current is empty, write desired
	if strings.TrimSpace(current) == "" {
		_, err := client.WriteSchema(ctx, &authzedpb.WriteSchemaRequest{Schema: desired})
		if err != nil {
			if logger != nil {
				logger.Error("Failed to write base SpiceDB schema", zap.Error(err))
			}
			return fmt.Errorf("write base spicedb schema: %w", err)
		}
		if logger != nil {
			logger.Info("Wrote base SpiceDB schema")
		}
		return nil
	}

	// If current exists, ensure required definitions exist; append any missing
	merged := mergeMissingNamespaces(current, desired)
	if merged != current { // some additions made
		_, err := client.WriteSchema(ctx, &authzedpb.WriteSchemaRequest{Schema: merged})
		if err != nil {
			if logger != nil {
				logger.Error("Failed to update SpiceDB schema with missing namespaces", zap.Error(err))
			}
			return fmt.Errorf("update spicedb schema: %w", err)
		}
		if logger != nil {
			logger.Info("Updated SpiceDB schema with missing namespaces")
		}
	}
	return nil
}

// buildBaseSchema returns a minimal yet extensible schema covering core AAA resources.
//
//go:embed schema/spicedb_schema.zed
var embeddedBaseSchema string

func buildBaseSchema() string {
	return embeddedBaseSchema
}

// mergeMissingNamespaces appends any definitions from desired that are missing in current.
func mergeMissingNamespaces(current, desired string) string {
	out := current
	// Collect desired definition headers and bodies
	desiredDefs := splitDefinitions(desired)
	for name, body := range desiredDefs {
		header := "definition " + name + " {"
		if !strings.Contains(out, header) {
			if !strings.HasSuffix(out, "\n\n") {
				out += "\n\n"
			}
			out += body
			if !strings.HasSuffix(out, "\n") {
				out += "\n"
			}
		}
	}
	return out
}

// splitDefinitions extracts definition blocks keyed by fully-qualified name (e.g., "aaa/user").
func splitDefinitions(schema string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(schema, "definition ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// part starts with: "aaa/user { ... }\n..."
		// Find the name up to first space or '{'
		nameEnd := strings.IndexAny(part, " {\n\t")
		if nameEnd <= 0 {
			continue
		}
		name := strings.TrimSpace(part[:nameEnd])
		body := "definition " + part
		// Ensure each block ends with a newline
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		result[name] = body
	}
	return result
}

// init attempts a best-effort bootstrap so that early DB manager connection to SpiceDB succeeds.
func init() { // nolint: gochecknoinits
	// Best-effort load .env early to ensure we have SpiceDB creds before DB manager connects
	_ = godotenv.Load()
	endpoint := os.Getenv("DB_SPICEDB_ENDPOINT")
	token := os.Getenv("DB_SPICEDB_TOKEN")
	if endpoint == "" || token == "" {
		return
	}
	// Create a lightweight client and write schema if needed
	client, err := authzed.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(token),
	)
	if err != nil {
		// Best effort; avoid panicking during init
		return
	}
	// Try multiple times in case SpiceDB isn't fully ready yet
	deadline := time.Now().Add(15 * time.Second)
	for attempt := 1; time.Now().Before(deadline); attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := ensureBaseSchema(ctx, client, nil)
		cancel()
		if err == nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
}

// EnsureSpiceDBSchema connects to SpiceDB with the provided endpoint/token and ensures
// the base schema exists, appending any missing namespaces. This is safe to call
// repeatedly and should be invoked before any components that require SpiceDB connectivity.
func EnsureSpiceDBSchema(ctx context.Context, endpoint, token string, logger *zap.Logger) error {
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(token) == "" {
		return nil
	}
	client, err := authzed.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(token),
	)
	if err != nil {
		return fmt.Errorf("create spicedb client: %w", err)
	}
	// Small timeout if none provided
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	return ensureBaseSchema(ctx, client, logger)
}

// ExtendNamespace creates or updates a resource namespace with the given actions.
// It appends a new definition if missing; if present, it does not remove existing permissions
// but can be enhanced later to merge.
func ExtendNamespace(ctx context.Context, client *authzed.Client, logger *zap.Logger, resource string, actions []string, hierarchical bool) error {
	name := fmt.Sprintf("aaa/%s", resource)
	// Build desired block
	var b strings.Builder
	b.WriteString("definition ")
	b.WriteString(name)
	b.WriteString(" {\n")
	if hierarchical {
		b.WriteString("    relation owner: aaa/user\n")
		b.WriteString("    relation parent: aaa/resource\n")
		// Basic inherited permissions
		for _, act := range actions {
			if act == "read" || act == "view" {
				b.WriteString("    permission view = owner + parent->view\n")
				b.WriteString("    permission read = view\n")
				continue
			}
			if act == "update" || act == "edit" {
				b.WriteString("    permission edit = owner + parent->edit\n")
				b.WriteString("    permission update = edit\n")
				continue
			}
			if act == "delete" {
				b.WriteString("    permission delete = owner + parent->delete\n")
				continue
			}
			if act == "manage" {
				b.WriteString("    permission manage = owner + parent->manage\n")
				continue
			}
			b.WriteString("    permission ")
			b.WriteString(act)
			b.WriteString(" = owner + parent->")
			b.WriteString(act)
			b.WriteString("\n")
		}
	} else {
		// Flat permissions referencing self
		for _, act := range actions {
			b.WriteString("    permission ")
			b.WriteString(act)
			b.WriteString(" = self\n")
		}
	}
	b.WriteString("}\n")

	desired := b.String()

	// Read current schema
	current := ""
	readResp, err := client.ReadSchema(ctx, &authzedpb.ReadSchemaRequest{})
	if err == nil && readResp != nil {
		current = readResp.SchemaText
	}
	merged := mergeMissingNamespaces(current, desired)
	if merged == current {
		// Already present
		return nil
	}
	_, err = client.WriteSchema(ctx, &authzedpb.WriteSchemaRequest{Schema: merged})
	if err != nil {
		if logger != nil {
			logger.Error("Failed to extend SpiceDB schema", zap.String("namespace", name), zap.Error(err))
		}
		return fmt.Errorf("extend spicedb schema: %w", err)
	}
	if logger != nil {
		logger.Info("Extended SpiceDB schema", zap.String("namespace", name))
	}
	return nil
}
