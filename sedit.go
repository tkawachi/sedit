package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

// mutable に original 内の secret を base64 decode する
func decodeSecret(original interface{}) {
	switch v := original.(type) {
	case map[interface{}]interface{}:
		if v["kind"] == "Secret" {
			switch data := v["data"].(type) {
			case map[interface{}]interface{}:
				for dk := range data {
					switch dv := data[dk].(type) {
					case string:
						decoded, err := base64.StdEncoding.DecodeString(dv)
						if err != nil {
							log.Fatal("Failed to decode base64. ", err)
						}
						data[dk] = string(decoded)
					default:
						log.Fatalf("Unexpected secret value %T.", dv)
					}
				}
			default:
				log.Fatalf("Unexpected data type %T.", data)
			}
		} else {
			for k := range v {
				decodeSecret(v[k])
			}
		}
	default:
		log.Fatalf("Unknown type %T.", v)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Must call with one argument. ", os.Args)
	}

	original, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var parsedYaml interface{}

	err = yaml.Unmarshal(original, &parsedYaml)
	if err != nil {
		log.Fatal(err)
	}

	decodeSecret(parsedYaml)
	c, err := yaml.Marshal(parsedYaml)
	if err != nil {
		log.Fatal(err)
	}

	file, err := ioutil.TempFile("", "sedit_")
	defer func() {
		err := os.Remove(file.Name())
		if err != nil {
			log.Fatal("Failed to remove temporary file. ", err)
		}
	}()
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(file.Name(), c, 0)
	if err != nil {
		log.Fatal(err)
	}

	if editor, defined := os.LookupEnv("EDITOR"); defined {
		cmd := exec.Command(editor, file.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println(file.Name())
}
