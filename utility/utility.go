package utility

import "github.com/satori/go.uuid"

func GenerateUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}
