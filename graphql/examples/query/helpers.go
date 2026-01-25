package query

import (
	"encoding/json"
	"fmt"
	"log"
)

// printJSON prints a value as pretty-printed JSON.
func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("JSON marshal error for %s: %v", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(data))
}
