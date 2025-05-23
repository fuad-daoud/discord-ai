package db

import "C"
import (
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/net/context"
	"os"
	"strings"
)

type dbConnection struct {
	driver neo4j.DriverWithContext
}

var connection *dbConnection

func InTransaction(execute TransactionExecute) error {
	ctx := context.Background()
	session := connection.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	transaction, err := session.BeginTransaction(ctx)
	if err != nil {
		dlog.Log.Error("Transaction failed", "err", err)
		return err
	}

	txWrite := getTxWrite(transaction, ctx)

	err = execute(txWrite)
	if err != nil {
		return err
	}
	err = transaction.Commit(ctx)
	if err != nil {
		err := transaction.Rollback(ctx)
		if err != nil {
			dlog.Log.Error("Rollback failed", "err", err)
			return err
		}
		dlog.Log.Error("Transaction failed", "err", err)
		return err
	}
	return nil
}

type Write func(stmts ...string) (neo4j.ResultWithContext, error)
type TransactionExecute func(write Write) error

func getTxWrite(transaction neo4j.ExplicitTransaction, ctx context.Context) Write {
	return func(stmts ...string) (neo4j.ResultWithContext, error) {
		stmt := strings.Join(stmts, " ")
		dlog.Log.Debug("Writing ", "stmt", stmt)
		run, err := transaction.Run(ctx, stmt, make(map[string]any))
		if err != nil {
			dlog.Log.Error("Transaction run failed", "err", err)
			return nil, err
		}
		return run, nil
	}
}

func Query(stmts ...string) (*neo4j.EagerResult, error) {
	stmt := strings.Join(stmts, " ")
	dlog.Log.Debug("Querying ", "stmt", stmt)
	result, err := neo4j.ExecuteQuery(context.Background(), connection.driver, stmt, make(map[string]any), neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		dlog.Log.Error("Error executing query", "err", err)
		return nil, err
	}
	return result, nil
}

func Connect() {
	dbUri := os.Getenv("NEO4J_DATABASE_URL")
	dbUser := os.Getenv("NEO4J_DATABASE_USER")
	dbPassword := os.Getenv("NEO4J_DATABASE_PASSWORD")
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		dlog.Log.Error("Error connecting to Neo4j", "err", err)
		// normal panic
		panic(err)
	}
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		dlog.Log.Error("Error connecting to Neo4j", "err", err)
		// normal panic
		panic(err)
	}
	connection = &dbConnection{driver: driver}
	dlog.Log.Info("Connection established.", "URI", dbUri)

	//InTransaction(func(write Write) {
	//	hostname, err := os.Hostname()
	//	if err != nil {
	//		dlog.Log.Error("Error getting hostname", "err", err)
	//		panic(err)
	//	}
	//	write(cypher.MergeN("s", Server{
	//		HostName:    hostname,
	//		CreatedDate: time.Now().String(),
	//	}))
	//})
}

func Close() {
	err := connection.driver.Close(context.Background())
	if err != nil {
		// normal panic
		panic(err)
	}
	dlog.Log.Info("db Connection closed.")
}
