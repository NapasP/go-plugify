package manifest

// Manifest represents the structure of the .pplugin file
type Manifest struct {
	Schema       string       `json:"$schema"`
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Description  string       `json:"description,omitempty"`
	Author       string       `json:"author,omitempty"`
	Website      string       `json:"website,omitempty"`
	License      string       `json:"license,omitempty"`
	Platforms    []string     `json:"platforms,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
	Conflicts    []Conflict   `json:"conflicts,omitempty"`
	Entry        string       `json:"entry"`
	Language     string       `json:"language"`
	Methods      []Method     `json:"methods"`
	Prototypes   []*Method    `json:"prototypes,omitempty"`
	Enums        []*Enum      `json:"enums,omitempty"`
}

// Method represents a single exported method
type Method struct {
	Name        string     `json:"name"`
	FuncName    string     `json:"funcName"`
	Description string     `json:"description,omitempty"`
	ParamTypes  []Property `json:"paramTypes"`
	RetType     Property   `json:"retType"`
	Group       string     `json:"group,omitempty"`
}

// Value represents a single enumeration value
type Value struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Value       int64  `json:"value"`
}

// Enum represents an enumeration
type Enum struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Values      []Value `json:"values"`
}

// Alias represents an alias definition
type Alias struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Property represents a parameter type
type Property struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Ref         bool   `json:"ref,omitempty"`
	Prototype   string `json:"prototype,omitempty"`
	Enum        string `json:"enum,omitempty"`
	Alias       *Alias `json:"alias,omitempty"`
}

// Dependency represents a plugin's dependency
type Dependency struct {
	Name        string `json:"name"`
	Constraints string `json:"constraints,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
}

// Conflict represents a plugin's conflict
type Conflict struct {
	Name        string `json:"name"`
	Constraints string `json:"constraints,omitempty"`
	Reason      string `json:"reason,omitempty"`
}
