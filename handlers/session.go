package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (t *MethodInterface) NewSession(w http.ResponseWriter, args map[string]interface{}) (result string, err error) {
	id := uuid.New()
	if _, e := fmt.Fprintf(w, "{\"session\": {\"id\": \"%v\"}}", id); e != nil {
		log.Fatal(e)
	}
	return
}
