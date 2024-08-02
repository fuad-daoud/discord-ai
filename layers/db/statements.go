package db

import (
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
