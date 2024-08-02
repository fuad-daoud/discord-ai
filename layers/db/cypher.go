package db

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
		if v.Kind() == reflect.Struct {
			labels.WriteString(":" + v.Type().Name())
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
		property, ok := ToProperty(value)
		if !ok {
			continue
		}
		stringBuilder.WriteString(fmt.Sprintf(`%s: `, key))
		stringBuilder.WriteString(property)
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
	inrec, err := json.Marshal(in)
	if err != nil {
		dlog.Log.Error("Error marshalling in", "in", in)
		return nil, err
	}
	var mp map[string]interface{}
	_ = json.Unmarshal(inrec, &mp)
	return mp, nil
}

func ParseAll[KeyValue any](key string, records []*neo4j.Record) ([]KeyValue, bool) {
	var (
		results = make([]KeyValue, 0)
	)
	if len(records) == 0 {
		return results, false
	}
	for _, record := range records {
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

func ParseKey[KeyValue any](key string, records []*neo4j.Record) (KeyValue, bool) {
	var (
		result KeyValue
	)
	if len(records) == 0 {
		return result, false
	}

	get, b := records[0].Get(key)
	if !b {
		dlog.Log.Error("Invalid key", "key", key)
		var zeroValue KeyValue
		return zeroValue, false
	}
	node := get.(neo4j.Node)

	result = parse[KeyValue](node.Props)
	return result, true
}

func parse[RESULT any](props map[string]any) RESULT {
	var result RESULT
	_ = mapstructure.Decode(props, &result)
	// not testable at all
	//if err != nil {
	//	dlog.Log.Error("Failed to decode result", "err", err)
	//	var zeroValue RESULT
	//	return zeroValue, err
	//}
	return result
}

func ToProperty(value any) (string, bool) {
	switch value.(type) {
	case string:
		{
			if value.(string) == "" {
				return "", false
			}
			value = strings.Replace(value.(string), "\"", "\\\"", -1)
			return fmt.Sprintf(`"%v",`, value), true
		}
	case []interface{}:
		{
			builder := strings.Builder{}
			builder.WriteString("[")
			elements := value.([]interface{})
			for _, element := range elements {
				property, b := ToProperty(element)
				if b {
					builder.WriteString(property)
				}
			}
			property := builder.String()
			property = property[:len(property)-1]
			property += "],"
			return property, true
		}
	case map[string]interface{}:
		{
			return "", false
		}
	default:
		return fmt.Sprintf(`%v,`, value), true
	}
}
