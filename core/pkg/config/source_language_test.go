package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test: SourceLanguage enum has correct values
// Given: SourceLanguage type with JAVA and GO constants
// When: Checking the string values
// Then: JAVA returns "JAVA" and GO returns "GO"
func TestSourceLanguage_String(t *testing.T) {
	assert.Equal(t, "JAVA", string(SourceLanguageJava))
	assert.Equal(t, "GO", string(SourceLanguageGo))
}

// Test: SourceLanguageFromString() returns correct enum for valid strings
// Given: String values "JAVA" and "GO"
// When: Calling SourceLanguageFromString()
// Then: Returns SourceLanguageJava for "JAVA" and SourceLanguageGo for "GO"
func TestSourceLanguageFromString_Valid(t *testing.T) {
	lang, err := SourceLanguageFromString("JAVA")
	assert.NoError(t, err)
	assert.Equal(t, SourceLanguageJava, lang)

	lang, err = SourceLanguageFromString("GO")
	assert.NoError(t, err)
	assert.Equal(t, SourceLanguageGo, lang)
}

// Test: SourceLanguageFromString() returns error for invalid strings
// Given: An invalid string like "PYTHON" or ""
// When: Calling SourceLanguageFromString()
// Then: Returns error indicating invalid source language
func TestSourceLanguageFromString_Invalid(t *testing.T) {
	_, err := SourceLanguageFromString("PYTHON")
	assert.Error(t, err)

	_, err = SourceLanguageFromString("")
	assert.Error(t, err)

	_, err = SourceLanguageFromString("java") // lowercase should fail
	assert.Error(t, err)
}

// Test: SourceLanguageFromString() is case-sensitive
// Given: String "java" (lowercase)
// When: Calling SourceLanguageFromString()
// Then: Returns error (only "JAVA" uppercase is valid)
func TestSourceLanguageFromString_CaseSensitive(t *testing.T) {
	_, err := SourceLanguageFromString("java")
	assert.Error(t, err)

	_, err = SourceLanguageFromString("go")
	assert.Error(t, err)

	// Uppercase should work
	lang, err := SourceLanguageFromString("JAVA")
	assert.NoError(t, err)
	assert.Equal(t, SourceLanguageJava, lang)
}

// Test: SourceLanguage.MarshalYAML() serializes correctly
// Given: SourceLanguageJava and SourceLanguageGo values
// When: Marshaling to YAML
// Then: Returns "JAVA" and "GO" strings respectively
func TestSourceLanguage_MarshalYAML(t *testing.T) {
	javaBytes, err := SourceLanguageJava.MarshalYAML()
	assert.NoError(t, err)
	assert.Equal(t, "JAVA", string(javaBytes))

	goBytes, err := SourceLanguageGo.MarshalYAML()
	assert.NoError(t, err)
	assert.Equal(t, "GO", string(goBytes))
}

// Test: SourceLanguage.UnmarshalYAML() deserializes correctly
// Given: YAML strings "JAVA" and "GO"
// When: Unmarshaling into SourceLanguage
// Then: Sets value to SourceLanguageJava and SourceLanguageGo respectively
func TestSourceLanguage_UnmarshalYAML(t *testing.T) {
	var lang SourceLanguage
	
	err := lang.UnmarshalYAML(func(v interface{}) error {
		*(v.(*string)) = "JAVA"
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, SourceLanguageJava, lang)

	err = lang.UnmarshalYAML(func(v interface{}) error {
		*(v.(*string)) = "GO"
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, SourceLanguageGo, lang)
}

// Test: SourceLanguage.UnmarshalYAML() returns error for invalid values
// Given: Invalid YAML string "INVALID"
// When: Unmarshaling into SourceLanguage
// Then: Returns error
func TestSourceLanguage_UnmarshalYAML_Invalid(t *testing.T) {
	var lang SourceLanguage
	
	err := lang.UnmarshalYAML(func(v interface{}) error {
		*(v.(*string)) = "INVALID"
		return nil
	})
	assert.Error(t, err)
}
