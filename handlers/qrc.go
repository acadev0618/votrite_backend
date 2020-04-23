package handlers

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"rsc.io/qr"
)

func (t *MethodInterface) GetQrc(w http.ResponseWriter, args map[string]interface{}) image.Image {
	var url string
	/*
		if _, ok := args["url"]; ok {
			url = args["url"].(string)
		} else {
			return nil
		}
	*/

	var qs string
	for k, v := range args {
		switch k {
		case "url":
			url = v.(string)
		default:
			if qs == "" {
				qs = fmt.Sprintf("%s=%s", k, v.(string))
			} else {
				qs = fmt.Sprintf("%s&%s=%s", qs, k, v.(string))
			}
		}
	}

	if qs != "" {
		url = fmt.Sprintf("%s?%s", url, qs)
	}
	log.Println(url)
	qrc, err := qr.Encode(url, 0)
	if err != nil {
		log.Fatal(err)
	}

	img := qrc.PNG()
	w.Header().Set("Content-Type", "image/png")
	if _, err := w.Write(img); err != nil {
		log.Fatal(err)
	}

	return qrc.Image()
}
