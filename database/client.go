package database

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/net/context"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

type Client interface {
	Connect() error
	Close()
	Queryf(stmt string, params map[string]any) (*neo4j.EagerResult, error)
	Query(stmt string) (*neo4j.EagerResult, error)
	Write(stmt string, params map[string]any) (neo4j.ResultWithContext, error)
	Merge(v any) (neo4j.ResultWithContext, error)
}

var client Client

func GetClient() Client {
	if client != nil {
		return client
	}
	client = &defaultClient{}
	err := client.Connect()
	if err != nil {
		panic(err)
	}
	return client
}

type defaultClient struct {
	driver neo4j.DriverWithContext
}

func (dc *defaultClient) Merge(v any) (neo4j.ResultWithContext, error) {
	session := dc.driver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	cypherProperties := toCypher(v)
	split := strings.Split(reflect.TypeOf(v).String(), ".")
	cypher := fmt.Sprintf("MERGE (v:%s %s)", split[len(split)-1], cypherProperties)
	slog.Info("Executing", "Cypher", cypher)
	return session.Run(context.TODO(), cypher, make(map[string]any))
}

func (dc *defaultClient) Write(stmt string, params map[string]any) (neo4j.ResultWithContext, error) {
	session := dc.driver.NewSession(context.TODO(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	return session.Run(context.TODO(), stmt, params)
}

func (dc *defaultClient) Query(stmt string) (*neo4j.EagerResult, error) {
	//neo4j.
	result, err := neo4j.ExecuteQuery(context.Background(), dc.driver, stmt, make(map[string]any), neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (dc *defaultClient) Queryf(stmt string, params map[string]any) (*neo4j.EagerResult, error) {
	result, err := neo4j.ExecuteQuery(context.Background(), dc.driver, stmt, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (dc *defaultClient) Connect() error {
	dbUri := os.Getenv("NEO4J_DATABASE_URL")
	dbUser := os.Getenv("NEO4J_DATABASE_USER")
	dbPassword := os.Getenv("NEO4J_DATABASE_PASSWORD")
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		return err
	}
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		return err
	}
	dc.driver = driver
	slog.Info("Connection established.", "URI", dbUri)
	return nil
}

func (dc *defaultClient) Close() {
	err := dc.driver.Close(context.Background())
	if err != nil {
		panic(err)
	}
	slog.Info("Connection closed.")
}

func toCypher(v any) string {
	m := toMap(v)

	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("{")

	for key, value := range m {
		switch value.(type) {
		case string:
			stringBuilder.WriteString(fmt.Sprintf(`%s: "%v",`, key, value))
		default:
			stringBuilder.WriteString(fmt.Sprintf(`%s: %v,`, key, value))
		}
	}
	cypher := stringBuilder.String()
	cypher = cypher[:len(cypher)-1]
	return cypher + "}"
}

func toMap(in any) map[string]interface{} {

	inrec, _ := json.Marshal(in)
	var mp map[string]interface{}
	err := json.Unmarshal(inrec, &mp)
	if err != nil {
		panic(err)
	}
	return mp
}
