package engine

import (
	"reflect"
	"testing"
)

func TestImageBase_BaseImage_JAVA(t *testing.T) {
	got := JAVA.BaseImage()
	want := "maven:3.9.9-eclipse-temurin-17"
	if got != want {
		t.Errorf("JAVA.BaseImage() = %q, want %q", got, want)
	}
}

func TestImageBase_EntryPoint_JAVA(t *testing.T) {
	tests := []struct {
		name                    string
		artifactParentDir       string
		artifactDefinitionNames []string
		want                    []string
	}{
		{
			name:                    "single artifact",
			artifactParentDir:       "/artifacts",
			artifactDefinitionNames: []string{"classes"},
			want:                    []string{"java", "-cp", "/artifacts/classes:/artifacts/classes/*"},
		},
		{
			name:                    "multiple artifacts",
			artifactParentDir:       "/artifacts",
			artifactDefinitionNames: []string{"classes", "dependencies"},
			want:                    []string{"java", "-cp", "/artifacts/classes:/artifacts/classes/*:/artifacts/dependencies:/artifacts/dependencies/*"},
		},
		{
			name:                    "empty artifact names",
			artifactParentDir:       "/artifacts",
			artifactDefinitionNames: []string{},
			want:                    []string{"java", "-cp", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JAVA.EntryPoint(tt.artifactParentDir, tt.artifactDefinitionNames)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JAVA.EntryPoint(%q, %v) = %v, want %v", tt.artifactParentDir, tt.artifactDefinitionNames, got, tt.want)
			}
		})
	}
}
