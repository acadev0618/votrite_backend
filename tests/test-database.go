package tests

import (
	"fmt"
	"net/http"

	"../managers/database"
	"log"
)

func GetElections(w http.ResponseWriter) (result string, err error) {
	log.Println("get all columns from elections")
	rm, err := database.Read("election", nil, nil)

	if err != nil {
		return "", err
	}

	jr, err := database.ResultToJSON(rm)

	fmt.Fprintf(w, "%s", jr)

	log.Println("get some columns from elections")
	columns := []string{"title", "description", "start_date"}
	rm, err = database.Read("election", columns, nil)

	if err != nil {
		return "", err
	}

	jr, err = database.ResultToJSON(rm)

	fmt.Fprintf(w, "%s", jr)

	log.Println("get some columns from elections with condition")
	columns = []string{"ballot_location", "start_date"}
	args := make(map[string]interface{})
	args["election_id"] = 1
	args["ballot_id"] = 3
	rm, err = database.Read("ballot", columns, args)

	if err != nil {
		return "", err
	}

	jr, err = database.ResultToJSON(rm)

	fmt.Fprintf(w, "%s", jr)

	return jr, err
}
