package db

import "C"
import (
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/net/context"
	"strings"
)

type Connection struct {
	driver neo4j.DriverWithContext
}

func (conn *Connection) Transaction(execute TransactionExecute) error {
	ctx := context.Background()
	session := conn.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
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
		err2 := transaction.Rollback(ctx)
		// not testable
		if err2 != nil {
			dlog.Log.Error("Rollback failed", "err", err2)
			return err2
		}
		dlog.Log.Error("Transaction failed", "err", err)
		return err
	}
	return nil
}

type Write func(stmts ...string) error
type TransactionExecute func(write Write) error

func getTxWrite(transaction neo4j.ExplicitTransaction, ctx context.Context) Write {
	return func(stmts ...string) error {
		stmt := strings.Join(stmts, " ")
		dlog.Log.Debug("Writing ", "stmt", stmt)
		_, err := transaction.Run(ctx, stmt, make(map[string]any))
		if err != nil {
			dlog.Log.Error("Transaction run failed", "err", err)
			return err
		}
		return nil
	}
}

func (conn *Connection) Query(stmts ...string) (*neo4j.EagerResult, error) {
	stmt := strings.Join(stmts, " ")
	dlog.Log.Debug("Querying ", "stmt", stmt)
	result, err := neo4j.ExecuteQuery(context.Background(), conn.driver, stmt, make(map[string]any), neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		dlog.Log.Error("Error executing query", "err", err)
		return nil, err
	}
	return result, nil
}

func (conn *Connection) Connect(dbUri, dbUser, dbPassword string) {
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		dlog.Log.Error("Error connecting to Neo4j", "err", err)
		panic(err)
	}
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		dlog.Log.Error("Error connecting to Neo4j", "err", err)
		panic(err)
	}
	conn.driver = driver
	dlog.Log.Info("Connection established.", "URI", dbUri)
}

func (conn *Connection) Close() {
	dlog.Log.Info("Closing Neo4j session")
	_ = conn.driver.Close(context.Background())
	dlog.Log.Info("db Connection closed.")
}

func NewConnection(dbUri, dbUser, dbPassword string) Connection {
	conn := Connection{}
	conn.Connect(dbUri, dbUser, dbPassword)
	return conn
}
