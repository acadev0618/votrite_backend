package handlers

import (
	"cloud.google.com/go/translate"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

func (t *MethodInterface) TranslateText(w http.ResponseWriter, args map[string]interface{}) (string, error) {
	ctx := context.Background()
	opt := option.WithAPIKey(os.Getenv("GCP_API_KEY"))
	client, err := translate.NewClient(ctx, opt)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return "", err
	}

	target, err := language.Parse(args["targetLang"].(string))
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return "", err
	}

	translations, err := client.Translate(ctx, []string{args["text"].(string)}, target, nil)
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return "", err
	}

	translation := translations[0].Text

	result := fmt.Sprintf("{\"translation\": {\"target\": \"%s\"}}", translation)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}
	return translation, nil
}

func (t *MethodInterface) TranslateScreen(w http.ResponseWriter, args map[string]interface{}) (map[string]interface{}, error) {
	ctx := context.Background()
	opt := option.WithAPIKey(os.Getenv("GCP_API_KEY"))
	client, err := translate.NewClient(ctx, opt)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return nil, err
	}

	if _, ok := args["screen"]; ok {
		for sk, sv := range args["screen"].(map[string]interface{}) {
			if sk == "element" {
				for _, ev := range sv.([]interface{}) {
					if _, ok := ev.(map[string]interface{})["translation"]; ok {
						if ev.(map[string]interface{})["translation"] != "en" {
							target, err := language.Parse(ev.(map[string]interface{})["translation"].(string))
							if err != nil {
								if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
									log.Fatal(e)
								}
							}
							if _, ok := ev.(map[string]interface{})["text"]; ok {
								translations, err := client.Translate(ctx, []string{ev.(map[string]interface{})["text"].(string)}, target, nil)
								if err != nil {
									if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
										log.Fatal(e)
									}
									return args["screen"].(map[string]interface{}), err
								}

								ev.(map[string]interface{})["text_translated"] = translations[0].Text
							}
						}
					}
				}
			}
		}
	} else {
		if _, e := fmt.Fprintf(w, "{\"error\": \"Missing screen in arguments\"}"); e != nil {
			log.Fatal(e)
		}
	}
	b, err := json.Marshal(args["screen"])
	if err != nil {
		fmt.Println("error:", err)
	}
	result := fmt.Sprintf("{\"screen\": %v}", string(b))
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}
	return args["screen"].(map[string]interface{}), nil
}

func (t *MethodInterface) TranslationSupportedLanguages(w http.ResponseWriter) ([]translate.Language, error) {
	ctx := context.Background()
	opt := option.WithAPIKey(os.Getenv("GCP_API_KEY"))
	client, err := translate.NewClient(ctx, opt)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return nil, err
	}
	target, err := language.Parse("en")
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return nil, err
	}
	languages, err := client.SupportedLanguages(ctx, target)
	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return nil, err
	}
	result := fmt.Sprintf("{\"translation\": {\"languages\": \"%v\"}}", languages)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}
	return languages, nil
}
