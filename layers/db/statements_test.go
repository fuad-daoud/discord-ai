package db

import "testing"

func TestMatches(t *testing.T) {
	t.Run("Testing Match", func(t *testing.T) {
		cypher := Match("something")
		if cypher != "MATCH something" {
			t.Fatalf("expected 'MATCH something', got '%s'", cypher)
		}
	})
	t.Run("Testing MatchN", func(t *testing.T) {
		cypher := MatchN("t", make(chan int))
		if cypher != "MATCH " {
			t.Fatalf("expected 'MATCH ', got '%s'", cypher)
		}
	})
}

func TestMerges(t *testing.T) {
	t.Run("Testing Merge", func(t *testing.T) {
		cypher := Merge("something")
		if cypher != "MERGE something" {
			t.Fatalf("expected 'MERGE something', got '%s'", cypher)
		}
	})
	t.Run("Testing MergeN", func(t *testing.T) {
		cypher := MergeN("t", make(chan int))
		if cypher != "MERGE " {
			t.Fatalf("expected 'MERGE ', got '%s'", cypher)
		}
	})
}
func TestCreates(t *testing.T) {
	t.Run("Testing Create", func(t *testing.T) {
		cypher := Create("something")
		if cypher != "CREATE something" {
			t.Fatalf("expected 'CREATE something', got '%s'", cypher)
		}
	})
	t.Run("Testing CreateN", func(t *testing.T) {
		cypher := CreateN("t", make(chan int))
		if cypher != "CREATE " {
			t.Fatalf("expected 'CREATE ', got '%s'", cypher)
		}
	})
}

func TestSet(t *testing.T) {
	t.Run("Testing failed Set should return err", func(t *testing.T) {
		_, err := Set("t", make(chan int))
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("Testing Set", func(t *testing.T) {
		cypher, _ := Set("t", struct{}{})
		if cypher != "SET t=" {
			t.Fatalf("expected 'SET', got '%s'", cypher)
		}
	})
}

func TestReturnDelete(t *testing.T) {
	t.Run("Testing Delete", func(t *testing.T) {
		cypher := Delete("t")
		if cypher != "DELETE t" {
			t.Fatalf("expected 'DELETE t', got '%s'", cypher)
		}
	})
	t.Run("Testing Return", func(t *testing.T) {
		cypher := Return("t")
		if cypher != "RETURN t" {
			t.Fatalf("expected 'RETURN t', got '%s'", cypher)
		}
	})

}
