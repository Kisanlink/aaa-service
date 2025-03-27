package helper

import (
	"encoding/json"
	"fmt"
	"strings"
)

func PrettyJSON(body interface{}) {

	marshaled, err := json.MarshalIndent(body, "", "   ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", string(marshaled))
}

func LowerCaseSlice(input []string) []string {
    for i, val := range input {
        input[i] = strings.ToLower(val)
    }
    return input
}