package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

func (t *MethodInterface) NewSession(w http.ResponseWriter, args map[string]interface{}) (result string, err error) {
	id := uuid.New()
	if _, e := fmt.Fprintf(w, "{\"session\": {\"id\": \"%v\"}}", id); e != nil {
		log.Fatal(e)
	}
	return
}


