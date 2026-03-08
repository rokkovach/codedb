package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jClient struct {
	driver neo4j.DriverWithContext
}

type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

func NewNeo4jClient(cfg Config) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(context.Background())
		return nil, fmt.Errorf("failed to verify connectivity: %w", err)
	}

	client := &Neo4jClient{driver: driver}

	if err := client.initializeSchema(ctx); err != nil {
		driver.Close(context.Background())
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

func (c *Neo4jClient) Close() error {
	return c.driver.Close(context.Background())
}

func (c *Neo4jClient) initializeSchema(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	queries := []string{
		"CREATE CONSTRAINT repository_id_unique IF NOT EXISTS FOR (r:Repository) REQUIRE r.id IS UNIQUE",
		"CREATE CONSTRAINT file_id_unique IF NOT EXISTS FOR (f:File) REQUIRE f.id IS UNIQUE",
		"CREATE CONSTRAINT symbol_id_unique IF NOT EXISTS FOR (s:Symbol) REQUIRE s.id IS UNIQUE",
		"CREATE CONSTRAINT commit_id_unique IF NOT EXISTS FOR (c:Commit) REQUIRE c.id IS UNIQUE",
		"CREATE CONSTRAINT workspace_id_unique IF NOT EXISTS FOR (w:Workspace) REQUIRE w.id IS UNIQUE",
	}

	for _, query := range queries {
		if _, err := session.Run(ctx, query, nil); err != nil {
			return fmt.Errorf("failed to create constraint: %w", err)
		}
	}

	return nil
}

func (c *Neo4jClient) CreateRepository(ctx context.Context, id, name string) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MERGE (r:Repository {id: $id})
		SET r.name = $name,
		    r.created_at = datetime(),
		    r.updated_at = datetime()
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"id":   id,
		"name": name,
	})

	return err
}

func (c *Neo4jClient) CreateFile(ctx context.Context, id, repoID, path string) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MATCH (r:Repository {id: $repoID})
		MERGE (f:File {id: $id})
		SET f.path = $path,
		    f.created_at = datetime(),
		    f.updated_at = datetime()
		MERGE (r)-[:CONTAINS]->(f)
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"id":     id,
		"repoID": repoID,
		"path":   path,
	})

	return err
}

func (c *Neo4jClient) CreateSymbol(ctx context.Context, id, fileID, name, kind, fullyQualifiedName string, lineStart int) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MATCH (f:File {id: $fileID})
		MERGE (s:Symbol {id: $id})
		SET s.name = $name,
		    s.kind = $kind,
		    s.fully_qualified_name = $fullyQualifiedName,
		    s.line_start = $lineStart,
		    s.created_at = datetime(),
		    s.updated_at = datetime()
		MERGE (f)-[:DEFINES]->(s)
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"id":                 id,
		"fileID":             fileID,
		"name":               name,
		"kind":               kind,
		"fullyQualifiedName": fullyQualifiedName,
		"lineStart":          lineStart,
	})

	return err
}

func (c *Neo4jClient) CreateCallRelationship(ctx context.Context, callerID, calleeID string) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MATCH (caller:Symbol {id: $callerID})
		MATCH (callee:Symbol {id: $calleeID})
		MERGE (caller)-[:CALLS]->(callee)
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"callerID": callerID,
		"calleeID": calleeID,
	})

	return err
}

func (c *Neo4jClient) CreateDependencyRelationship(ctx context.Context, fromFileID, toFileID string) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
		MATCH (from:File {id: $fromFileID})
		MATCH (to:File {id: $toFileID})
		MERGE (from)-[:DEPENDS_ON]->(to)
	`

	_, err := session.Run(ctx, query, map[string]interface{}{
		"fromFileID": fromFileID,
		"toFileID":   toFileID,
	})

	return err
}

type ImpactResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Distance int    `json:"distance"`
}

func (c *Neo4jClient) GetTransitiveDependents(ctx context.Context, symbolID string, maxDepth int) ([]ImpactResult, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (target:Symbol {id: $symbolID})
		MATCH (dependent:Symbol)-[:CALLS|REFERENCES*1..` + fmt.Sprintf("%d", maxDepth) + `]->(target)
		MATCH (f:File)-[:DEFINES]->(dependent)
		RETURN DISTINCT dependent.id, dependent.name, dependent.kind, f.path, length(shortestPath((dependent)-[*]->(target))) as distance
		ORDER BY distance, dependent.name
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"symbolID": symbolID,
	})
	if err != nil {
		return nil, err
	}

	var dependents []ImpactResult
	for result.Next(ctx) {
		record := result.Record()
		dependents = append(dependents, ImpactResult{
			ID:       record.Values[0].(string),
			Name:     record.Values[1].(string),
			Kind:     record.Values[2].(string),
			Path:     record.Values[3].(string),
			Distance: int(record.Values[4].(int64)),
		})
	}

	return dependents, result.Err()
}

func (c *Neo4jClient) GetTransitiveDependencies(ctx context.Context, symbolID string, maxDepth int) ([]ImpactResult, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (source:Symbol {id: $symbolID})
		MATCH (source)-[:CALLS|REFERENCES*1..` + fmt.Sprintf("%d", maxDepth) + `]->(dependency:Symbol)
		MATCH (f:File)-[:DEFINES]->(dependency)
		RETURN DISTINCT dependency.id, dependency.name, dependency.kind, f.path, length(shortestPath((source)-[*]->(dependency))) as distance
		ORDER BY distance, dependency.name
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"symbolID": symbolID,
	})
	if err != nil {
		return nil, err
	}

	var dependencies []ImpactResult
	for result.Next(ctx) {
		record := result.Record()
		dependencies = append(dependencies, ImpactResult{
			ID:       record.Values[0].(string),
			Name:     record.Values[1].(string),
			Kind:     record.Values[2].(string),
			Path:     record.Values[3].(string),
			Distance: int(record.Values[4].(int64)),
		})
	}

	return dependencies, result.Err()
}

func (c *Neo4jClient) AnalyzeImpact(ctx context.Context, fileIDs []string) ([]ImpactResult, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (f:File)
		WHERE f.id IN $fileIDs
		MATCH (f)-[:DEFINES]->(s:Symbol)
		MATCH (dependent:Symbol)-[:CALLS|REFERENCES*1..10]->(s)
		MATCH (depFile:File)-[:DEFINES]->(dependent)
		RETURN DISTINCT dependent.id, dependent.name, dependent.kind, depFile.path, 1 as distance
		ORDER BY depFile.path, dependent.name
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"fileIDs": fileIDs,
	})
	if err != nil {
		return nil, err
	}

	var impacted []ImpactResult
	for result.Next(ctx) {
		record := result.Record()
		impacted = append(impacted, ImpactResult{
			ID:       record.Values[0].(string),
			Name:     record.Values[1].(string),
			Kind:     record.Values[2].(string),
			Path:     record.Values[3].(string),
			Distance: int(record.Values[4].(int64)),
		})
	}

	return impacted, result.Err()
}

func (c *Neo4jClient) FindCircularDependencies(ctx context.Context, repoID string) ([][]string, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (r:Repository {id: $repoID})-[:CONTAINS]->(f:File)
		MATCH (f1:File)-[:DEPENDS_ON]->(f2:File),
		      (f2)-[:DEPENDS_ON*]->(f1)
		RETURN f1.path, f2.path
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"repoID": repoID,
	})
	if err != nil {
		return nil, err
	}

	var cycles [][]string
	for result.Next(ctx) {
		record := result.Record()
		cycles = append(cycles, []string{
			record.Values[0].(string),
			record.Values[1].(string),
		})
	}

	return cycles, result.Err()
}

func (c *Neo4jClient) GetCallGraph(ctx context.Context, symbolID string, depth int) (map[string]interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH path = (s:Symbol {id: $symbolID})-[:CALLS*0..` + fmt.Sprintf("%d", depth) + `]-(other:Symbol)
		RETURN s, other, relationships(path)
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"symbolID": symbolID,
	})
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]interface{})
	edges := make([]map[string]interface{}, 0)

	for result.Next(ctx) {
		record := result.Record()
		nodes[record.Values[0].(neo4j.Node).ElementId] = record.Values[0]
		nodes[record.Values[1].(neo4j.Node).ElementId] = record.Values[1]
		edges = append(edges, map[string]interface{}{
			"from": record.Values[0].(neo4j.Node).ElementId,
			"to":   record.Values[1].(neo4j.Node).ElementId,
		})
	}

	return map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}, result.Err()
}
