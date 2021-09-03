package main

import (
	"errors"
	"os"
	"strings"
)

func processGivenPrompt() ([]string, error) {
	if len(os.Args) != 2 {
		return nil, errors.New("please provide a prompt file")
	}
	filePath := os.Args[1]

	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	var ps []string

	for _, s := range strings.Split(string(data), "\n") {
		if s != "" {
			ps = append(ps, s)
		}
	}
	return ps, nil
}
