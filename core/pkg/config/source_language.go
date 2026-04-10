package config

import (
	"fmt"

	"github.com/yeetcd/yeetcd/pkg/engine"
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

// GetGeneratePipelineDefinitionsCmd returns the command to generate pipeline definitions
func (s SourceLanguage) GetGeneratePipelineDefinitionsCmd() []string {
	switch s {
	case SourceLanguageJava:
		return []string{"yeetcd.sdk.GeneratedPipelineDefinitions"}
	default:
		return nil
	}
}

// GetImageBase returns the ImageBase for this language
func (s SourceLanguage) GetImageBase() engine.ImageBase {
	switch s {
	case SourceLanguageJava:
		return engine.JAVA
	default:
		return -1
	}
}

// GetCustomTaskRunnerCmd returns the command to run custom work for this language
func (s SourceLanguage) GetCustomTaskRunnerCmd(pipelineName, executionID string) []string {
	switch s {
	case SourceLanguageJava:
		// The ENTRYPOINT in Dockerfile sets up java -cp <classpath>
		// We just need to provide the main class and arguments
		return []string{"yeetcd.sdk.GeneratedCustomWorkRunner", pipelineName, executionID}
	default:
		return nil
	}
}
