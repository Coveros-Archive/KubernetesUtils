package stub

import (
	"encoding/base64"
	"fmt"
)

func errorCheck(err error, block func()) error {
	if err != nil {
		block()
		return err
	}
	return nil
}

func encodeDecode(str, action string) string {
	var finalString string
	switch action {
	case "encode":
		finalString = base64.StdEncoding.EncodeToString([]byte(str))
		break
	case "decode":
		deocdedByte, _ := base64.StdEncoding.DecodeString(str)
		finalString = fmt.Sprintf("%s", deocdedByte)
	}

	return finalString
}
