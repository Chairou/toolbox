package routinepool

import (
	"fmt"
	"github.com/google/uuid"
)

// GetUUIDString get uuid string
func GetUUIDString() string {
	id, err := uuid.NewUUID()
	if err != nil {
		_ = fmt.Errorf("GetUUIDString error: %v", err)
		return ""
	}

	return id.String()
}
