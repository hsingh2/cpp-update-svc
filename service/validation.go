package app

import (
	"fmt"
)

/*
 * this is header for the file. if a file doesn't matches this format,
 * we discard the file and reject it saying a bad request
 */
var (
	//constHeader ...
	constHeader = []string{"itrc_db_id", "location_md5", "UpdatedBy", "PiTag", "NewPiTag", "NewComment"}
)

const (
	//UpdateCommentRequest ....
	UpdateCommentRequest = "update"

	//VerifyCommentRequest ....
	VerifyCommentRequest = "verify"

	//AddCommentRequest ....
	AddCommentRequest = "add"
)

//isValidFile ; checks for the required fields in the file. If not present we will discard the file immediately
func isValidFile(header []string) error {
	headerMap := make(map[string]bool, len(header))

	//adding to map
	for _, val := range header {
		headerMap[val] = true
	}

	//check if expected header is there in the file header

	for _, val := range constHeader {
		if ok := headerMap[val]; !ok {
			return fmt.Errorf("invalid file, header doesn't include required field %s", val)
		}
	}

	return nil
}
