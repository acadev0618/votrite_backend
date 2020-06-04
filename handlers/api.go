package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"../managers/database"
)

func (t *MethodInterface) Read(w http.ResponseWriter, args map[string]interface{}) (result string, err error) {
	var schema string
	var dataSource string
	var order string
	var fields []string

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
		dataSource = ds["name"].(string)
		if _, ok := ds["order"]; ok {
			for _, v := range ds["order"].([]interface{}) {
				for kk, vv := range v.(map[string]interface{}) {
					if order == "" {
						order = fmt.Sprintf("%s %s", kk, vv)
					} else {
						order += fmt.Sprintf(", %s %s", kk, vv)
					}
				}
			}
		}
		if _, ok := ds["fields"]; ok {
			for _, v := range ds["fields"].([]interface{}) {
				fields = append(fields, v.(string))
			}
		}

		delete(args, "data-source")
	}

	rm, err := database.Read(schema, dataSource, fields, order, args)

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

	result, err = database.ResultToJSON(rm)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}
	return
}

func (t *MethodInterface) Create(w http.ResponseWriter, args map[string]interface{}) (lastInsertedId int64, err error) {
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
	lastInsertedId, err = database.Create(schema, dataSource, args)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"state\": \"error\", \"message\": \"%s\"}", strings.ReplaceAll(err.Error(), "\"", "")); e != nil {
			log.Fatal(e)
		}
		return
	}
	//result := fmt.Sprintf("{\"inserted\": \"%d\"}", lastInsertedId)
	result := fmt.Sprintf("{\"state\": \"success\",  \"message\": \"%d\"}", lastInsertedId)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}

	return
}

func (t *MethodInterface) Update(w http.ResponseWriter, args map[string]interface{}) (rowsAffected int64, err error) {
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
		args["modified_at"] = time.Now().Format(time.RFC3339)
		if _, ok := args["modified_by"]; !ok {
			args["modified_by"] = "admin"
		}
	}
	rowsAffected, err = database.Update(schema, dataSource, args)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"state\": \"error\", \"message\": \"%s\"}", strings.ReplaceAll(err.Error(), "\"", "")); e != nil {
			log.Fatal(e)
		}
		return
	}

	result := fmt.Sprintf("{\"state\": \"success\",  \"message\": \"%d\"}", rowsAffected)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}

	return
}

func (t *MethodInterface) Delete(w http.ResponseWriter, args map[string]interface{}) (rowsAffected int64, err error) {
	var schema string
	var dataSource string

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
		dataSource = ds["name"].(string)
		delete(args, "data-source")
	}

	rowsAffected, err = database.Delete(schema, dataSource, args)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"state\": \"error\", \"message\": \"%s\"}", strings.ReplaceAll(err.Error(), "\"", "")); e != nil {
			log.Fatal(e)
		}
		return
	}

	result := fmt.Sprintf("{\"state\": \"success\",  \"message\": \"%d\"}", rowsAffected)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}

	return
}

func (t *MethodInterface) Mixin(w http.ResponseWriter, args map[string]interface{}) (result string, err error) {
	ds := args["data-source"].(map[string]interface{})
	query := ds["query"].(string)
	delete(args, "data-source")

	rm, err := database.Mixin(query, args)
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

	result, err = database.ResultToJSON(rm)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}
	return
}

func (t *MethodInterface) GetApiStatus(w http.ResponseWriter, args map[string]interface{}) (rowsAffected int64, err error) {
	status, err := database.GetStatus()
	if err != nil {
		status = "500"
	} else {
		status = "200"
	}
	if _, err := fmt.Fprintf(w, "%s", status); err != nil {
		log.Fatal(err)
	}
	return
}

func (t *MethodInterface) GetApiList(w http.ResponseWriter, args map[string]interface{}) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	js, err := ioutil.ReadFile(fmt.Sprintf("%s/routes.json", pwd))
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	if _, err := fmt.Fprintf(w, "%s", js); err != nil {
		log.Fatal(err)
	}
	return
}
