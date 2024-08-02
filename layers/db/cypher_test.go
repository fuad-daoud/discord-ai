package db

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"reflect"
	"strings"
	"testing"
)

func TestToMap(t *testing.T) {
	mp, err := toMap(struct {
		Id   string
		some string
	}{
		Id:   "someId",
		some: "someValue",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(mp) != 1 {
		t.Fatal("expected 1 element")
	}
	id, ok := mp["Id"]
	if !ok {
		t.Fatal("id not found")
	}
	if id != "someId" {
		t.Fatalf("got %s, want someId", id)
	}
	some, ok := mp["some"]
	if ok {
		t.Fatalf("got %s, want none", some)
	}
}
func TestFailedToMap(t *testing.T) {
	t.Run("Testing giving invalid object to be marshalled", func(t *testing.T) {
		_, err := toMap(make(chan int))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestToProperties(t *testing.T) {
	cypher, err := ToProperties(struct {
		Id     string
		some   string
		Arr    []int
		StrArr []string
	}{
		Id:     "someId",
		some:   "someValue",
		Arr:    []int{1, 2, 3},
		StrArr: []string{"1", "2", "3"},
	})
	if err != nil {
		t.Fatal(err)
	}

	id := `Id: "someId"`
	arr := `Arr: [1,2,3]`
	strArr := `StrArr: ["1","2","3"]`

	if !strings.Contains(cypher, id) {
		t.Fatalf("cypher does not contain id %s, cypher: %s", id, cypher)
	}
	if !strings.Contains(cypher, arr) {
		t.Fatalf("cypher does not contain arr %s, cypher: %s", arr, cypher)
	}
	if !strings.Contains(cypher, strArr) {
		t.Fatalf("cypher does not contain strArr %s, cypher: %s", strArr, cypher)
	}
	if strings.Count(cypher, ",") != 6 {
		t.Fatalf("cypher does not contain all elements %s", cypher)
	}
}
func TestFailedProperties(t *testing.T) {
	t.Run("Test giving nil to ToProperties", func(t *testing.T) {
		_, err := ToProperties(nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Testing toMap returning err", func(t *testing.T) {
		_, err := ToProperties(make(chan int))
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("Testing passing empty struct should return empty string", func(t *testing.T) {
		cypher, _ := ToProperties(struct{}{})
		if cypher != "" {
			t.Fatalf("got %s, want empty string", cypher)
		}
	})
	t.Run("Testing passing struct with field string being empty should return empty string", func(t *testing.T) {
		cypher, _ := ToProperties(struct {
			Name string `json:"name"`
		}{Name: ""})
		if cypher != "" {
			t.Fatalf("got %s, want empty string", cypher)
		}
	})
	t.Run("Testing passing nested structs should return empty string", func(t *testing.T) {
		cypher, _ := ToProperties(struct{ Name struct{} }{})
		if cypher != "" {
			t.Fatalf("got %s, want empty string", cypher)
		}
	})
}

func TestCypher(t *testing.T) {
	type Testing struct{}
	cypher := Cypher("T", struct {
		Testing
		Id     string
		some   string
		Arr    []int
		StrArr []string
	}{
		Testing: Testing{},
		Id:      "someId",
		some:    "someValue",
		Arr:     []int{1, 2, 3},
		StrArr:  []string{"1", "2", "3"},
	})

	if !strings.HasPrefix(cypher, "(T") {
		t.Fatalf("cypher does not start with (T, cypher: %s", cypher)
	}

	id := `Id: "someId"`
	arr := `Arr: [1,2,3]`
	strArr := `StrArr: ["1","2","3"]`

	if !strings.Contains(cypher, id) {
		t.Fatalf("cypher does not contain id %s, cypher: %s", id, cypher)
	}
	if !strings.Contains(cypher, arr) {
		t.Fatalf("cypher does not contain arr %s, cypher: %s", arr, cypher)
	}
	if !strings.Contains(cypher, strArr) {
		t.Fatalf("cypher does not contain strArr %s, cypher: %s", strArr, cypher)
	}
	if strings.Count(cypher, ",") != 6 {
		t.Fatalf("cypher does not contain all elements %s", cypher)
	}
}
func TestFailedCypher(t *testing.T) {
	t.Run("Testing making ToProperties return err should return empty string", func(t *testing.T) {
		cypher := Cypher("t", make(chan int))
		if cypher != "" {
			t.Fatalf("got %s, want empty string", cypher)
		}
	})
}

func TestParseKey(t *testing.T) {
	type S struct {
		Id     string
		some   string
		Arr    []int
		StrArr []string
	}

	parsed, ok := ParseKey[S]("t", []*neo4j.Record{
		{
			Keys: []string{"t"},
			Values: []any{
				neo4j.Node{
					Id:        0,
					ElementId: "",
					Labels:    []string{"Test"},
					Props: map[string]interface{}{
						"Id":     "someId",
						"Arr":    []int{1, 2, 3},
						"StrArr": []string{"1", "2", "3"},
					},
				},
			},
		},
	})
	if !ok {
		t.Fatal("parsed key not found")
	}
	if parsed.some != "" {
		t.Fatalf("got %s, want %s", parsed.some, "someValue")
	}
	if parsed.Id != "someId" {
		t.Fatalf("got %s, want %s", parsed.Id, "someId")
	}
	if !reflect.DeepEqual(parsed.Arr, []int{1, 2, 3}) {
		t.Fatalf("got %+v, want %+v", parsed.Arr, []int{1, 2, 3})
	}
	if !reflect.DeepEqual(parsed.StrArr, []string{"1", "2", "3"}) {
		t.Fatalf("got %s, want none", parsed.StrArr)
	}
}
func TestFailedParseKey(t *testing.T) {
	t.Run("Testing giving zero records", func(t *testing.T) {
		_, ok := ParseKey[any]("s", make([]*neo4j.Record, 0))
		if ok {
			t.Fatalf("expected failure")
		}
	})
	t.Run("Testing giving key not in records", func(t *testing.T) {
		_, ok := ParseKey[any]("s", []*neo4j.Record{
			{
				Keys: []string{"t"},
				Values: []any{
					neo4j.Node{},
				},
			},
		})
		if ok {
			t.Fatalf("expected failure")
		}
	})
}

func TestParseAll(t *testing.T) {

	type S struct {
		Id     string
		some   string
		Arr    []int
		StrArr []string
	}
	parsed, ok := ParseAll[S]("t", []*neo4j.Record{
		{
			Keys: []string{"t"},
			Values: []any{
				neo4j.Node{
					Id:        0,
					ElementId: "",
					Labels:    []string{"Test"},
					Props: map[string]interface{}{
						"Id":     "someId",
						"Arr":    []int{1, 2, 3},
						"StrArr": []string{"1", "2", "3"},
					},
				},
			},
		},
	})
	if !ok {
		t.Fatal("parsed key not found")
	}
	if len(parsed) != 1 {
		t.Fatalf("got %d, want %d", len(parsed), 1)
	}
	s := parsed[0]
	if s.some != "" {
		t.Fatalf("got %s, want %s", s.some, "someValue")
	}
	if s.Id != "someId" {
		t.Fatalf("got %s, want %s", s.Id, "someId")
	}
	if !reflect.DeepEqual(s.Arr, []int{1, 2, 3}) {
		t.Fatalf("got %+v, want %+v", s.Arr, []int{1, 2, 3})
	}
	if !reflect.DeepEqual(s.StrArr, []string{"1", "2", "3"}) {
		t.Fatalf("got %s, want none", s.StrArr)
	}
}
func TestFailedParseAll(t *testing.T) {
	t.Run("Testing giving zero records", func(t *testing.T) {
		_, ok := ParseAll[any]("s", make([]*neo4j.Record, 0))
		if ok {
			t.Fatalf("expected failure")
		}
	})

	t.Run("Testing giving key not in records", func(t *testing.T) {
		_, ok := ParseAll[any]("s", []*neo4j.Record{
			{
				Keys: []string{"t"},
				Values: []any{
					neo4j.Node{},
				},
			},
		})
		if ok {
			t.Fatalf("expected failure")
		}
	})
}
