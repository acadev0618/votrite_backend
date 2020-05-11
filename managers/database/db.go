package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type DataSource struct {
	ColumnName string
	DataType   string
}

const (
	driver = "postgres"
	// schema = "public"
)

var connStr string

func SetDbConnString(s string) {
	connStr = s
}

func ResultToJSON(result map[string][]map[string]interface{}) (newResult string, err error) {
	res, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	newResult = string(res)
	return
}

func Read(schema string, dataSource string, columns []string, order string, args ...interface{}) (result map[string][]map[string]interface{}, err error) {
	var target string
	if len(columns) > 0 {
		for _, v := range columns {
			if target == "" {
				target = v
			} else {
				target += fmt.Sprintf(", %s", v)
			}
		}
	} else {
		target = "*"
	}

	query := fmt.Sprintf("SELECT %s FROM %s.%s", target, schema, dataSource)

	var whereClause string
	var qArgs []interface{}
	//var tableName string
	//var joinClause string
	//aliases := []string{"a", "b", "c", "d", "e", "j"}

	for _, v := range args {
		switch v.(type) {
		case map[string]interface{}:
			qArgs = make([]interface{}, len(v.(map[string]interface{})))
			var i = 1
			//var a = 0
			for key, val := range v.(map[string]interface{}) {
				if key == "data-source" {
					_, ok := v.(map[string]interface{})["dependencies"]
					if ok {
						for k, v := range val.(map[string]interface{}) {
							if k == "name" {
								log.Println(k, v)
							}
						}
					}
					/*for k, v := range val.(map[string]interface{}) {
						if k == "name" {
							tableName = v.(string)
						}
					}

					if tableName != "" {
						if joinClause == "" {
							joinClause = fmt.Sprintf("JOIN %s %s", tableName, aliases[a])
						} else {
							joinClause += fmt.Sprintf("JOIN %s %s", tableName, aliases[a])
						}
						a++
					}*/
				} else {
					if whereClause == "" {
						switch val.(type) {
						case string:
							if strings.Contains(val.(string), "encrypted:") {
								whereClause = fmt.Sprintf("WHERE %s=crypt($%d, %s)", key, i, key)
								val = val.(string)[10:]
							} else {
								whereClause = fmt.Sprintf("WHERE %s=$%d", key, i)
							}
						default:
							whereClause = fmt.Sprintf("WHERE %s=$%d", key, i)
						}
					} else {
						switch val.(type) {
						case string:
							if strings.Contains(val.(string), "encrypted:") {
								whereClause += fmt.Sprintf(" AND %s=crypt($%d, %s)", key, i, key)
								val = val.(string)[10:]
							} else {
								whereClause += fmt.Sprintf(" AND %s=$%d", key, i)
							}
						default:
							whereClause += fmt.Sprintf(" AND %s=$%d", key, i)
						}
					}
				}
				qArgs[i-1] = val
				i++
			}
			break
		default:
			break
		}
	}

	if query == "" {
		return
	}

	if whereClause != "" {
		query = fmt.Sprintf("%s %s", query, whereClause)
	}
	if order != "" {
		query = fmt.Sprintf("%s ORDER BY %s", query, order)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	var rows *sql.Rows
	log.Println(query)
	log.Println(qArgs)
	rows, err = db.Query(query, qArgs...)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return nil, err
	}

	i := 0
	cols, _ := rows.Columns()
	count := len(cols)
	vals := make([]interface{}, count)
	valuesPtr := make([]interface{}, count)
	result = make(map[string][]map[string]interface{})

	for c := range cols {
		valuesPtr[c] = &vals[c]
	}

	for rows.Next() {
		err := rows.Scan(valuesPtr...)

		if err != nil {
			if err := rows.Close(); err != nil {
				log.Fatal(err)
			}
			if err := db.Close(); err != nil {
				log.Fatal(err)
			}
			return nil, err
		}

		rm := make(map[string]interface{})

		for j, col := range cols {

			var v interface{}

			val := vals[j]

			b, ok := val.([]byte)

			if ok {
				v = string(b)
			} else {
				v = val
			}

			rm[col] = v
		}

		result["data"] = append(result["data"], rm)
		i++
	}
	if err := rows.Close(); err != nil {
		log.Fatal(err)
	}

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func Create(schema string, dataSource string, args ...interface{}) (lastInsertId int64, err error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	cols, err := getDataSourceInfo(schema, dataSource)
	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Println(cols)

	var rowId = cols[0].ColumnName
	var columns string
	var qArgs []interface{}
	var vArgs string
	for _, v := range args {
		switch v.(type) {
		case map[string]interface{}:
			qArgs = make([]interface{}, len(v.(map[string]interface{})))
			var i = 1
			for key, val := range v.(map[string]interface{}) {
				if columns == "" {
					columns = fmt.Sprintf("%s", key)
					vArgs = fmt.Sprintf("$%d", i)
					if key == "user_password" {
						vArgs = fmt.Sprintf("crypt($%d, gen_salt('bf'))", i)
					}
				} else {
					columns += fmt.Sprintf(" ,%s", key)
					switch key {
					case "user_password":
						vArgs += fmt.Sprintf(", crypt($%d, gen_salt('bf'))", i)
						break
					default:
						vArgs += fmt.Sprintf(", $%d", i)
					}
				}
				qArgs[i-1] = val
				i++
			}
			break
		default:
			break
		}
	}

	query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING %s",
		schema, dataSource, columns, vArgs, rowId)
	log.Println(query, qArgs)

	if err = db.QueryRow(query, qArgs...).Scan(&lastInsertId); err != nil {
		if er := db.Close(); er != nil {
			log.Fatal(er)
		}
		return
	}

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func Update(schema string, dataSource string, args ...interface{}) (rowsAffected int64, err error) {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	var values string
	var whereClause string
	var qArgs []interface{}

	for _, v := range args {
		switch v.(type) {
		case map[string]interface{}:
			if _, ok := v.(map[string]interface{})["keys"]; ok {
				qArgs = make([]interface{}, (len(v.(map[string]interface{}))-1)+
					len(v.(map[string]interface{})["keys"].(map[string]interface{})))
			} else {
				qArgs = make([]interface{}, len(v.(map[string]interface{})))
			}

			var i = 1
			for key, val := range v.(map[string]interface{}) {
				if key == "keys" {
					for ka, kv := range val.(map[string]interface{}) {
						if whereClause == "" {
							whereClause = fmt.Sprintf("WHERE %s=$%d", ka, i)
						} else {
							whereClause += fmt.Sprintf(" AND %s=$%d", ka, i)
						}
						qArgs[i-1] = kv
						i++
					}
				} else {
					if values == "" {
						values = fmt.Sprintf("%s=$%d", key, i)
					} else {
						values += fmt.Sprintf(", %s=$%d", key, i)
					}
					qArgs[i-1] = val
					i++
				}
			}
			break
		default:
			break
		}
	}

	query := fmt.Sprintf("UPDATE %s.%s SET %s %s",
		schema, dataSource, values, whereClause)

	stmt, err := db.Prepare(query)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Println(query, qArgs)

	res, err := stmt.Exec(qArgs...)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	rowsAffected, err = res.RowsAffected()

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func Delete(schema string, dataSource string, args ...interface{}) (rowsAffected int64, err error) {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	var whereClause string
	var qArgs []interface{}

	for _, v := range args {
		switch v.(type) {
		case map[string]interface{}:
			qArgs = make([]interface{}, len(v.(map[string]interface{})))
			var i = 1
			for key, val := range v.(map[string]interface{}) {
				if whereClause == "" {
					whereClause = fmt.Sprintf("WHERE %s=$%d", key, i)
				} else {
					whereClause += fmt.Sprintf(" AND %s=$%d", key, i)
				}
				qArgs[i-1] = val
				i++
			}
			break
		default:
			break
		}
	}

	if whereClause == "" {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return -1, errors.New("missing conditional arguments for DELETE statement")
	}

	query := fmt.Sprintf("DELETE FROM %s.%s %s", schema, dataSource, whereClause)

	stmt, err := db.Prepare(query)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	res, err := stmt.Exec(qArgs...)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	rowsAffected, err = res.RowsAffected()

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func Mixin(query string, args ...interface{}) (result map[string][]map[string]interface{}, err error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	var qArgs []interface{}

	for _, v := range args {
		switch v.(type) {
		case map[string]interface{}:
			qArgs = make([]interface{}, len(v.(map[string]interface{})))
			var i = 0
			for _, val := range v.(map[string]interface{}) {
				qArgs[i] = val
				i++
			}
			break
		default:
			break
		}
	}

	var rows *sql.Rows
	log.Println(query)
	rows, err = db.Query(query, qArgs...)

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return nil, err
	}

	i := 0
	cols, _ := rows.Columns()
	count := len(cols)
	vals := make([]interface{}, count)
	valuesPtr := make([]interface{}, count)
	result = make(map[string][]map[string]interface{})

	for c := range cols {
		valuesPtr[c] = &vals[c]
	}

	for rows.Next() {
		err := rows.Scan(valuesPtr...)

		if err != nil {
			if err := rows.Close(); err != nil {
				log.Fatal(err)
			}
			if err := db.Close(); err != nil {
				log.Fatal(err)
			}
			return nil, err
		}

		rm := make(map[string]interface{})

		for j, col := range cols {

			var v interface{}

			val := vals[j]

			b, ok := val.([]byte)

			if ok {
				v = string(b)
			} else {
				v = val
			}

			rm[col] = v
		}

		result["data"] = append(result["data"], rm)
		i++
	}
	if err := rows.Close(); err != nil {
		log.Fatal(err)
	}
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func GetStatus() (status string, err error) {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func getDataSourceInfo(schema string, dataSource string) ([]DataSource, error) {
	db, err := sql.Open(driver, connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	var ds []DataSource
	rows, err := db.Query("SELECT * "+
		"FROM information_schema.columns "+
		"WHERE table_schema=$1 "+
		"AND table_name=$2", schema, dataSource)

	// defer rows.Close()

	if err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return nil, err
	}

	i := 0
	cols, _ := rows.Columns()
	count := len(cols)
	vals := make([]interface{}, count)
	valuesPtr := make([]interface{}, count)
	result := make(map[string][]map[string]interface{})

	for c := range cols {
		valuesPtr[c] = &vals[c]
	}

	for rows.Next() {
		err := rows.Scan(valuesPtr...)

		if err != nil {
			if err := db.Close(); err != nil {
				log.Fatal(err)
			}
			return nil, err
		}

		rm := make(map[string]interface{})

		for j, col := range cols {

			var v interface{}

			val := vals[j]

			b, ok := val.([]byte)

			if ok {
				v = string(b)
			} else {
				v = val
			}

			rm[col] = v
		}

		result["data"] = append(result["data"], rm)
		i++
	}

	if _, ok := result["data"]; !ok {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		return nil, errors.New("missing data in the result")
	} else {
		for _, row := range result["data"] {
			d := DataSource{}
			for k, v := range row {
				switch k {
				case "column_name":
					d.ColumnName = v.(string)
					break
				case "data_type":
					d.DataType = v.(string)
					break
				}
			}
			ds = append(ds, d)
		}
	}

	if err := rows.Close(); err != nil {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
	return ds, nil
}
