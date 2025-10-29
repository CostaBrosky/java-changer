package java

// Version represents a Java installation
type Version struct {
	Version  string // Version string (e.g., "17.0.1", "1.8.0_322")
	Path     string // Full path to Java installation
	IsCustom bool   // Whether this is from custom paths or auto-detected
}
