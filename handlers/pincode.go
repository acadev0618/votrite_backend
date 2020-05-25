package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"../common"
	"../managers/database"
)

var alphaNum = []rune("0123456789")

func randSeq(n uint) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = alphaNum[r.Intn(len(alphaNum))]
	}
	return string(b)
}

func (t *MethodInterface) CreatePinCode(w http.ResponseWriter, args map[string]interface{}) (pin string, err error) {
	var pinLen uint

	if _, ok := args["pincode_len"]; !ok {
		pinLen = 10
	} else {
		ret, e := common.InterfaceToType(args["pincode_len"], "uint")
		delete(args, "pincode_len")
		if e != nil {
			if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", e); err != nil {
				log.Fatal(err)
			}
			return
		}

		pinLen = ret.(uint)
	}

	if pinLen > 64 {
		if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", "pincode length exceeded max size (64)"); err != nil {
			log.Fatal(err)
		}
		return
	}

	if _, ok := args["pin"]; !ok {
		args["pin"] = randSeq(pinLen)
	}

	var schema string
	var dataSource string

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
		dataSource = ds["name"].(string)
		delete(args, "data-source")
	}

	if len(args) > 0 {
		args["created_at"] = time.Now().Format(time.RFC3339)
		if _, ok := args["created_by"]; !ok {
			args["created_by"] = "admin"
		}
	}
	lastInsertedId, err := database.Create(schema, dataSource, args)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	result := fmt.Sprintf("{\"inserted\": \"%d\"}", lastInsertedId)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}

	pin = strconv.FormatUint(uint64(lastInsertedId), 10)

	return
}

func (t *MethodInterface) CreatePinCodeMulti(w http.ResponseWriter, args map[string]interface{}) (pin string, err error) {
	var pinLen uint
	var pinCount uint

	if _, ok := args["pincode_len"]; !ok {
		pinLen = 10
	} else {
		ret, e := common.InterfaceToType(args["pincode_len"], "uint")
		delete(args, "pincode_len")
		if e != nil {
			if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", e); err != nil {
				log.Fatal(err)
			}
			return
		}

		pinLen = ret.(uint)
	}

	if pinLen > 64 {
		if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", "pincode length exceeded max size (64)"); err != nil {
			log.Fatal(err)
		}
		return
	}
	if _, ok := args["pincode_count"]; !ok {
		return
	} else {
		ret, e := common.InterfaceToType(args["pincode_count"], "uint")
		delete(args, "pincode_count")
		if e != nil {
			if _, err := fmt.Fprintf(w, "{\"error\": \"%v\"}", e); err != nil {
				log.Fatal(err)
			}
			return
		}

		pinCount = ret.(uint)
	}
	if pinCount <= 0 {
		return
	}

	var schema string
	var dataSource string

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
		dataSource = ds["name"].(string)
		delete(args, "data-source")
	}

	if len(args) > 0 {
		args["created_at"] = time.Now().Format(time.RFC3339)
		if _, ok := args["created_by"]; !ok {
			args["created_by"] = "admin"
		}
	}

	log.Println(pinCount)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	for pinCount > 0 {
		args["pin"] = randSeq(pinLen)
		pinCount--
		lastInsertedId, err := database.Create(schema, dataSource, args)
		log.Println(args)

		if err != nil {
			if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
				log.Fatal(e)
			}
			continue
		}

		result := fmt.Sprintf("{\"inserted\": \"%d\"}", lastInsertedId)
		if _, err := fmt.Fprintf(w, result); err != nil {
			log.Fatal(err)
		}
	}

	return
}
