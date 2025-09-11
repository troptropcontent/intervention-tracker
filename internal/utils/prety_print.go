package utils

import (
	"encoding/json"
	"fmt"
)

func PP(s any) {
	jsonData, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(jsonData))
}
