package providers

import (
	"fmt"
	"net/http"
)

func mustWriteResponse(w http.ResponseWriter, payload string) {
	if _, err := w.Write([]byte(payload)); err != nil {
		panic(fmt.Sprintf("failed to write response: %v", err))
	}
}
