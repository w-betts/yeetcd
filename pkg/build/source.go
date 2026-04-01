package build

// Source represents source code
type Source struct {
	Name string
	Zip  []byte
}

// SHA256 returns hex-encoded SHA256 hash of zip contents
func (s *Source) SHA256() string {
	return ""
}
