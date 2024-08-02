package db

import (
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"testing"
)

const (
	dbUri      = "neo4j://localhost:7687"
	dbUser     = "neo4j"
	dbPassword = "neo4j"
)

func TestConnectionClose(t *testing.T) {
	connection := NewConnection(dbUri, dbUser, dbPassword)
	connection.Close()
}

func TestCreateDeleteNode(t *testing.T) {
	connection := Connection{}
	t.Cleanup(connection.Close)
	connection.Connect(dbUri, dbUser, dbPassword)
	err := connection.Transaction(func(write Write) error {
		err := write(`CREATE (T:TEST {id: "test123"}) RETURN T`)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = connection.Transaction(func(write Write) error {
		err := write(`MATCH (T:TEST {id: "test123"}) DELETE T`)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateNodeQueryNode(t *testing.T) {
	connection := Connection{}
	t.Cleanup(connection.Close)
	t.Cleanup(removeAllNodes(&connection))
	connection.Connect(dbUri, dbUser, dbPassword)
	err := connection.Transaction(func(write Write) error {
		err := write(`CREATE (T:TEST {id: "test123"}) RETURN T`)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	query, err := connection.Query(`MATCH (T:TEST {id: "test123"}) RETURN T`)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.Keys) != 1 {
		t.Fatalf("Query returned wrong number of keys. Expected: 1, got: %d", len(query.Keys))
	}
	key := query.Keys[0]
	if key != "T" {
		t.Fatalf("Query returned wrong key. Expected: T, got: %s", key)
	}
	if len(query.Records) != 1 {
		t.Fatalf("Query returned wrong number of records. Expected: 1, got: %d", len(query.Records))
	}
	record := query.Records[0]
	get, b := record.Get(key)
	if !b {
		t.Fatalf("Query returned wrong key. Expected: value, got: %s", key)
	}
	node := get.(dbtype.Node)
	if len(node.Labels) != 1 {
		t.Fatalf("Query returned wrong number of labels. Expected: 1, got: %d", len(node.Labels))
	}
	if node.Labels[0] != "TEST" {
		t.Fatalf("Query returned wrong label. Expected: TEST, got: %s", node.Labels[0])
	}
	if len(node.Props) != 1 {
		t.Fatalf("Query returned wrong number of properties. Expected: 1, got: %d", len(node.Props))
	}
	id := node.Props["id"]
	if id != "test123" {
		t.Fatalf("Query returned wrong id. Expected: test123, got: %s", id)
	}
}

func removeAllNodes(connection *Connection) func() {
	return func() {
		fmt.Println("removing all nodes")
		err := connection.Transaction(func(write Write) error {
			err := write(`MATCH (T:TEST) DELETE T`)
			if err != nil {
				panic(err)
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

}

func TestFailedTransaction(t *testing.T) {
	t.Run("Testing Transaction on closed connection", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		connection.Close()
		err := connection.Transaction(func(write Write) error {
			return nil
		})
		if err == nil {
			t.Fatalf("did not get expected error")
		}
	})
	t.Run("Testing Transaction on failed transaction execute function", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		err := connection.Transaction(func(write Write) error {
			return errors.New("return error")
		})
		if err == nil {
			t.Fatalf("did not get expected error")
		}
		connection.Close()
	})
	t.Run("Testing Transaction on failed write function", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		err := connection.Transaction(func(write Write) error {
			return write("invalid statement")
		})
		if err == nil {
			t.Fatalf("did not get expected error")
		}
		connection.Close()
	})
	t.Run("Testing Transaction failing on commit", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		err := connection.Transaction(func(write Write) error {
			connection.Close()
			return nil
		})
		if err == nil {
			t.Fatalf("did not get expected error")
		}
		connection.Close()
	})
}

func TestFailedQuery(t *testing.T) {
	t.Run("Testing Query on closed connection", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		connection.Close()
		_, err := connection.Query(`MATCH (T:TEST {id: "test123"}) RETURN T`)
		if err == nil {
			t.Fatalf("did not get expected error")
		}
	})

	t.Run("Testing invalid Query", func(t *testing.T) {
		connection := NewConnection(dbUri, dbUser, dbPassword)
		_, err := connection.Query(`MATCH (T:TEST {id: "test123"}) RETURN `)
		if err == nil {
			t.Fatalf("did not get expected error")
		}
	})
}

func TestFailedConnection(t *testing.T) {
	t.Run("Testing Connection with invalid uri", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("did not get expected panic")
			}
		}()
		NewConnection("", dbUser, dbPassword)
	})

	t.Run("Testing Connection with invalid creds", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("did not get expected panic")
			}
		}()
		NewConnection("neo4j+s://8081da8a.databases.neo4j.io:7687", "user1231235", dbPassword)
	})
}
