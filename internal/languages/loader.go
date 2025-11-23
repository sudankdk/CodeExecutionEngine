package languages

import (
	"encoding/json"
	"errors"
	"os"
)

type Language struct {
	Image string `json:"image"`
	Entry string `json:"entry"`
	Ext   string `json:"extension"`
}

type LanguageMap map[string]Language

var cached LanguageMap

func Load(path string) (LanguageMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m LanguageMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	cached = m
	return m, nil
}

func Get(name string) (Language, error) {
	if cached == nil {
		return Language{}, errors.New("languages not loaded")
	}
	l, ok := cached[name]
	if !ok {
		return Language{}, errors.New("language not found")
	}
	return l, nil
}
