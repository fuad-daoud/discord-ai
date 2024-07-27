package cypher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fuad-daoud/discord-ai/logger/dlog"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
func Set(key string, val any) (string, error) {
	properties, err := ToProperties(val)
	if err != nil {
		return "", err
	}
	return "SET " + key + "=" + properties, nil
}

func Return(keys ...string) string {
	return "RETURN " + strings.Join(keys, ",")

}
func Delete(keys ...string) string {
	return "DELETE " + strings.Join(keys, ",")

}

func Cypher(key string, val any) string {
	cypherProperties, err := ToProperties(val)

	if err != nil {
		dlog.Log.Error("Cypher: "+err.Error(), "err", err)
		return ""
	}

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

func ToProperties(val any) (string, error) {
	if val == nil {
		dlog.Log.Error("Val in ToProperties can't be nil")
		return "", errors.New("val in ToProperties can't be nil")
	}

	m, err := toMap(val)
	if err != nil {
		return "", err
	}

	stringBuilder := strings.Builder{}
	stringBuilder.WriteString("{")

	for key, value := range m {
		switch value.(type) {
		case string:
			if value == "" {
				continue
			}
			value = strings.Replace(value.(string), "\"", "\\\"", -1)
			stringBuilder.WriteString(fmt.Sprintf(`%s: "%v",`, key, value))
			break
		case []interface{}:
			stringBuilder.WriteString(fmt.Sprintf(`%s: `, key))
			stringBuilder.WriteString("[")
			elements := value.([]interface{})
			if len(elements) != 0 {
				stringBuilder.WriteString(fmt.Sprintf(`"%v"`, elements[0]))
			}
			for i, element := range elements {
				if i == 0 {
					continue
				}
				stringBuilder.WriteString(", ")
				stringBuilder.WriteString(fmt.Sprintf(`"%v"`, element))
			}
			stringBuilder.WriteString("],")
			break
		default:
			stringBuilder.WriteString(fmt.Sprintf(`%s: %v,`, key, value))
		}
	}
	cypherProperties := stringBuilder.String()
	if len(cypherProperties) == 1 {
		return "", nil
	}
	cypherProperties = cypherProperties[:len(cypherProperties)-1]
	cypherProperties = cypherProperties + "}"
	return cypherProperties, nil
}

func toMap(in any) (map[string]interface{}, error) {
	inrec, _ := json.Marshal(in)
	var mp map[string]interface{}
	err := json.Unmarshal(inrec, &mp)
	if err != nil {
		dlog.Log.Error("failed to convert to map", "in", in, "err", err)
		return nil, err
	}
	return mp, nil
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
			dlog.Log.Error("Invalid key", "key", key)
			return nil, false
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
		dlog.Log.Error("Invalid key", "key", key)
		// can only panic
		panic("Invalid key")
	}
	node := get.(neo4j.Node)

	result = parse[KeyValue](node.Props)
	return result, true
}

func parse[RESULT any](props map[string]any) RESULT {
	var result RESULT
	err := mapstructure.Decode(props, &result)
	if err != nil {
		dlog.Log.Error("Failed to decode result", "err", err)
		// can only panic
		panic(err)
	}
	return result
}
