package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

var bodySep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

func parseBody(requestBody io.ReadCloser, target interface{}) (string, error) {
	body, err := ioutil.ReadAll(requestBody)

	if err != nil {
		return "", fmt.Errorf("failed to read body, %v", err)
	}

	docs := bodySep.Split(string(body), 2)

	if len(docs) != 2 {
		return "", fmt.Errorf("cannot split body, got %d elements after split, but 2 elements required", len(docs))
	}

	params, data := []byte(docs[0]), docs[1]

	if err := yaml.Unmarshal(params, target); err != nil {
		return "", fmt.Errorf("cannot unmarshal params, %v", err)
	}

	return data, nil
}
