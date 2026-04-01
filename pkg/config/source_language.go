package config

import (
	"fmt"
)

// SourceLanguage represents the source language type
type SourceLanguage string

const (
	SourceLanguageJava SourceLanguage = "JAVA"
	SourceLanguageGo   SourceLanguage = "GO"
)

// SourceLanguageFromString converts a string to SourceLanguage
func SourceLanguageFromString(s string) (SourceLanguage, error) {
	switch s {
	case "JAVA":
		return SourceLanguageJava, nil
	case "GO":
		return SourceLanguageGo, nil
	default:
		return "", fmt.Errorf("invalid source language: %s", s)
	}
}

// String returns the string representation
func (s SourceLanguage) String() string {
	return string(s)
}

// MarshalYAML implements yaml.Marshaler
func (s SourceLanguage) MarshalYAML() ([]byte, error) {
	return []byte(string(s)), nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (s *SourceLanguage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	
	lang, err := SourceLanguageFromString(str)
	if err != nil {
		return err
	}
	
	*s = lang
	return nil
}
