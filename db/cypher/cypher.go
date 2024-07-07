package cypher

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log/slog"
	"reflect"
	"strings"
)

func MatchN(key string, val any) string {
	return "MATCH " + Cypher(key, val)
}
func Match(stmt string) string {
	return "MATCH " + stmt
}
func Merge(stmt string) string {
	return "MERGE " + stmt
}
func MergeN(key string, val any) string {
	return "MERGE " + Cypher(key, val)
}
func Create(stmt string) string {
	return "CREATE " + stmt
}
func CreateN(key string, val any) string {
	return "CREATE " + Cypher(key, val)
}
func Set(key string, val any) string {
	return "SET " + key + "=" + ToProperties(val)
}

func Return(keys ...string) string {
	return "RETURN " + strings.Join(keys, ",")

}

func Cypher(key string, val any) string {
	cypherProperties := ToProperties(val)

	labels := strings.Builder{}
	ifv := reflect.ValueOf(val)
	ift := reflect.TypeOf(val)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)

		switch v.Kind() {
		case reflect.Struct:
			labels.WriteString(":" + v.Type().Name())
		default:

		}
	}
	labels.WriteString(":" + reflect.TypeOf(val).Name())
	return fmt.Sprintf("(%s%s %s)", key, labels.String(), cypherProperties)
}

func ToProperties(val any) string {
	if val == nil {
		panic("Val in ToProperties can't be nil")
	}

	m := toMap(val)

	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("{")

	for key, value := range m {
		switch value.(type) {
		case string:
			if value == "" {
				continue
			}
			stringBuilder.WriteString(fmt.Sprintf(`%s: "%v",`, key, value))
		default:
			stringBuilder.WriteString(fmt.Sprintf(`%s: %v,`, key, value))
		}
	}
	cypherProperties := stringBuilder.String()
	cypherProperties = cypherProperties[:len(cypherProperties)-1]
	cypherProperties = cypherProperties + "}"
	return cypherProperties
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

func ParseAll[KeyValue any](key string, eagerResult *neo4j.EagerResult) ([]KeyValue, bool) {
	var (
		results = make([]KeyValue, 0)
	)
	if len(eagerResult.Records) == 0 {
		return results, false
	}
	for _, record := range eagerResult.Records {
		get, b := record.Get(key)
		if !b {
			slog.Error("Invalid key", "key", key)
			panic("Invalid key")
		}
		node := get.(neo4j.Node)

		result := parse[KeyValue](node.Props)
		results = append(results, result)
	}
	return results, true
}

func ParseKey[KeyValue any](key string, eagerResult *neo4j.EagerResult) (KeyValue, bool) {
	var (
		result KeyValue
	)
	if len(eagerResult.Records) == 0 {
		return result, false
	}

	get, b := eagerResult.Records[0].Get(key)
	if !b {
		slog.Error("Invalid key", "key", key)
		panic("Invalid key")
	}
	node := get.(neo4j.Node)

	result = parse[KeyValue](node.Props)
	return result, true
}

func Parse2Key[FirstKeyValue any, SecondKeyValue any](firstKey, secondKey string, eagerResult *neo4j.EagerResult) (FirstKeyValue, SecondKeyValue, bool) {
	var (
		firstResult  FirstKeyValue
		secondResult SecondKeyValue
	)
	if len(eagerResult.Records) == 0 {
		return firstResult, secondResult, false
	}

	get, b := eagerResult.Records[0].Get(firstKey)
	if !b {
		slog.Error("Invalid key", "key", &firstKey)
		panic("Invalid key")
	}
	firstResult = parse[FirstKeyValue](get.(neo4j.Node).Props)
	if len(eagerResult.Records) == 1 {
		return firstResult, secondResult, false
	}
	get, b = eagerResult.Records[1].Get(secondKey)
	if !b {
		slog.Error("Invalid key", "key", secondKey)
		panic("Invalid key")
	}
	secondResult = parse[SecondKeyValue](get.(neo4j.Node).Props)
	return firstResult, secondResult, true
}

func parse[RESULT any](props map[string]any) RESULT {
	var result RESULT
	err := mapstructure.Decode(props, &result)
	if err != nil {
		slog.Error("Failed to decode result", "err", err)
		panic(err)
	}
	return result
}
