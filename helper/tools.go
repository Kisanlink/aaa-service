package helper

import (
	"encoding/json"
	"fmt"
)

func PrettyJSON(body interface{}) {

	marshaled, err := json.MarshalIndent(body, "", "   ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", string(marshaled))
}
