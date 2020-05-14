package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"../managers/database"
	"github.com/google/uuid"
)

func (t *MethodInterface) CreateVoter(w http.ResponseWriter, args map[string]interface{}) (lastInsertedId int64, err error) {
	var schema string
	var dataSource string
	ts := time.Now().Format(time.RFC3339)

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
	}

	pArgs := make(map[string]interface{})
	pArgs["created_at"] = ts

	if len(args) > 0 {
		if _, ok := args["created_by"]; !ok {
			pArgs["created_by"] = "admin"
		} else {
			pArgs["created_by"] = args["created_by"]
		}
	}

	pArgs["voter_email"] = args["voter_email"]
	pArgs["voter_phone"] = args["voter_phone"]
	pArgs["registration_token"] = uuid.New().String()

	dataSource = "voter"

	lastInsertedId, err = database.Create(schema, dataSource, pArgs)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
	}

	// Create pin code
	pArgs = make(map[string]interface{})
	pArgs["ballot_id"] = args["ballot_id"]
	pArgs["is_active"] = true
	pArgs["expiration_time"] = "2020-12-31T23:59:59"
	pArgs["created_at"] = ts
	pArgs["pincode_len"] = uint(10)
	pArgs["voter_id"] = lastInsertedId
	ds := make(map[string]interface{})
	ds["schema"] = schema
	ds["name"] = "pincode"
	pArgs["data-source"] = ds

	pin, err := t.CreatePinCode(w, pArgs)
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	message := fmt.Sprintf("Welcome to online voting by Votrite\n\nThank you for your registration!\nYou successfully registered for %s. Your PIN code for the election is %s", args["ballot_name"].(string), pin)

	if args["pin_delivery"].(string) == "email" {
		err = t.SendMail(args["voter_email"].(string),
			fmt.Sprintf("Registration for %s", args["ballot_name"].(string)),
			message)
	}
	if args["pin_delivery"].(string) == "text" {
		err = t.SendSms(
			fmt.Sprintf("%s:demo-election", os.Getenv("AWS_SNS_ARN")),
			fmt.Sprintf("+1001%s", args["voter_phone"].(string)),
			message)
	}

	return lastInsertedId, err
}
