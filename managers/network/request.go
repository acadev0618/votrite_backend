package network

import (
	"../../common"
	"../../handlers"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

var (
	// allowedClients = []string{"192.168.0.101", "192.168.0.103", "10.0.1.114", "127.0.0.1", "::1"}
	routesFunc = func() (routes interface{}) {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data := make(map[string]interface{})
		js := handlers.ReadRoutes(fmt.Sprintf("%s/routes.json", pwd))
		err = json.Unmarshal([]byte(js), &data)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}

		for k, v := range data {
			if k == "routes" {
				routes = v
			}
		}

		return
	}
	routes = routesFunc()
	apiErr = common.ApiError{}
)

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Requested endpoint: %s", r.URL.Path)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		log.Printf("userip: %q is not IP:port", r.RemoteAddr)
	}

	remAddr := net.ParseIP(ip).String()
	log.Printf("Client address: %v", remAddr)
	/*remAddrFound := false

	for i := range allowedClients {
		if allowedClients[i] == remAddr {
			remAddrFound = true
			break
		}
	}

	if !remAddrFound {
		fmt.Fprintf(w, common.ErrorToJSON(403))
		return
	}*/

	route := r.URL.Path[1:]

	routeMap := make(map[string]interface{})
	for _, v := range routes.([]interface{}) {
		m := v.(map[string]interface{})
		if m[route] != nil {
			routeMap = m[route].(map[string]interface{})
			routeMap["endpoint"] = route
			break
		}
	}

	if len(routeMap) == 0 {
		if _, err := fmt.Fprintf(w, apiErr.ErrorToJSON(404)); err != nil {
			log.Fatal(err)
		}
		return
	}

	args := make(map[string]interface{})

	methodFound := false
	for _, m := range strings.Split(routeMap["method"].(string), ",") {
		if m == r.Method {
			methodFound = true
			break
		}
	}

	if !methodFound {
		if _, err := fmt.Fprintf(w, apiErr.ErrorToJSON(405)); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch r.Method {
	case "GET":
		if args, err = handleGet(r, routeMap); err != nil {
			if _, err := fmt.Fprintf(w, apiErr.ErrorFromString(err.Error())); err != nil {
				log.Fatal(err)
			}
			return
		}
	case "POST":
		if args, err = handlePost(r); err != nil {
			if _, err := fmt.Fprintf(w, apiErr.ErrorFromString(err.Error())); err != nil {
				log.Fatal(err)
			}
			return
		}
	default:
		if _, err := fmt.Fprintf(w, apiErr.ErrorToJSON(405)); err != nil {
			log.Fatal(err)
		}
		return
	}

	if _, ok := routeMap["data-source"]; ok {
		args["data-source"] = routeMap["data-source"]
	}

	_, ok := routeMap["args"]
	if ok {
		ok, err := checkArgs(routeMap["args"].([]interface{}), args)

		if err != nil {
			if _, err := fmt.Fprintf(w, apiErr.ErrorFromString(err.Error())); err != nil {
				log.Fatal(err)
			}
			return
		}

		if !ok {
			if _, err := fmt.Fprintf(w, apiErr.ErrorToJSON(500)); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	t := &handlers.MethodInterface{}
	_, err = callHandler(t, w, routeMap, args)

	if err != nil {
		log.Printf("An error has occured: %v", err)
	}

	return
}

func checkArgs(routeMapArgs []interface{}, args map[string]interface{}) (bool, error) {
	for _, v := range routeMapArgs {
		for key := range v.(map[string]interface{}) {
			if key == "default" || key == "required" || key == "encrypted" {
				continue
			}

			if req := v.(map[string]interface{})["required"]; req == true {
				_, ok := args[key]
				if !ok {
					return false, errors.New(fmt.Sprintf("missing required argument %s", key))
				} else {
					a := args[key]
					nt, err := common.InterfaceToType(a, v.(map[string]interface{})[key].(string))
					if err != nil {
						return false, errors.New(fmt.Sprintf("can't cast argument %s", key))
					}

					if req := v.(map[string]interface{})["encrypted"]; req == true {
						_, ok := args[key]
						if ok {
							args[key] = fmt.Sprintf("encrypted:%s", nt)
						} else {
							args[key] = nt
						}
					}
				}
			}
		}
	}

	return true, nil
}

func handleGet(r *http.Request, route map[string]interface{}) (args map[string]interface{}, err error) {
	args = make(map[string]interface{})
	qs := r.URL.RawQuery

	if qs != "" {
		if _, ok := route["args"]; !ok {
			return
		}

		if u, err := url.ParseQuery(qs); err != nil {
			if err := r.Body.Close(); err != nil {
				log.Fatal(err)
			}
			log.Fatal(err)
		} else {
			for k, v := range u {
				for _, vv := range v {

					if strings.Index(vv, "?") != -1 {
						uq, err := url.Parse(vv)
						if err != nil {
							log.Fatal(err)
						}
						vq, err := url.ParseQuery(uq.RawQuery)
						if err != nil {
							log.Fatal(err)
						}

						for qk, qq := range vq {
							for _, qv := range qq {
								val, err := url.QueryUnescape(qv)
								if err != nil {
									log.Fatal(err)
								}
								args[qk] = val
							}
						}

						u := fmt.Sprintf("%s://%s", uq.Scheme, uq.Host)
						val, err := url.QueryUnescape(u)
						if err != nil {
							log.Fatal(err)
						}
						args[k] = val
					} else {
						val, err := url.QueryUnescape(vv)
						if err != nil {
							log.Fatal(err)
						}
						args[k] = val
					}
				}
			}
		}

	} else {
		if r.ContentLength == 0 {
			if err := r.Body.Close(); err != nil {
				log.Fatal(err)
			}
			return
		}

		decoder := json.NewDecoder(r.Body)
		rm := make(map[string]interface{})

		if err = decoder.Decode(&rm); err != nil {
			if err := r.Body.Close(); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func handlePost(r *http.Request) (args map[string]interface{}, err error) {
	args = make(map[string]interface{})
	decoder := json.NewDecoder(r.Body)
	rm := make(map[string]interface{})

	if err = decoder.Decode(&rm); err != nil {
		if err := r.Body.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}

	qs := r.URL.RawQuery

	if qs != "" {
		q := strings.Split(qs, "&")
		qsArgs := make(map[string]interface{})
		for _, v := range q {
			if v != "" {
				va := strings.Split(v, "=")

				if len(va) == 2 {
					qsArgs[va[0]] = va[1]
				}
			}
		}
		args["keys"] = qsArgs
	}

	for k, v := range rm {
		args[k] = v
	}

	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func callHandler(t interface{}, w http.ResponseWriter, m map[string]interface{},
	args map[string]interface{}) (result []reflect.Value, err error) {
	mt := reflect.ValueOf(t)
	mn := mt.MethodByName(m["handler"].(string))

	in := make([]reflect.Value, mn.Type().NumIn())

	for i := 0; i < mn.Type().NumIn(); i++ {
		if i == 0 {
			in[i] = reflect.ValueOf(w)
		} else {
			in[i] = reflect.ValueOf(args)
		}
	}

	result = mn.Call(in)
	return nil, err
}
