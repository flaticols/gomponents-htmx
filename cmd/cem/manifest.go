// Package main provides a Custom Elements Manifest generator using the TypeScript Go compiler.
// This tool parses TypeScript files and generates a custom-elements.json manifest
// following the Custom Elements Manifest specification.
//
// It supports:
// - Class inheritance resolution
// - Properties, methods, attributes, events, slots, and CSS custom properties
// - JSDoc comment parsing
// - Multiple file analysis
package main

// Manifest represents the root of a Custom Elements Manifest.
// See: https://github.com/webcomponents/custom-elements-manifest
type Manifest struct {
	// Schema version of the manifest
	SchemaVersion string `json:"schemaVersion"`

	// README content for the package
	Readme string `json:"readme,omitempty"`

	// Modules in the package
	Modules []*Module `json:"modules"`
}

// Module represents a JavaScript module in the manifest.
type Module struct {
	// Kind is always "javascript-module"
	Kind string `json:"kind"`

	// Path to the module file
	Path string `json:"path"`

	// Summary of the module
	Summary string `json:"summary,omitempty"`

	// Description of the module
	Description string `json:"description,omitempty"`

	// Declarations in this module
	Declarations []Declaration `json:"declarations,omitempty"`

	// Exports from this module
	Exports []*Export `json:"exports,omitempty"`

	// Whether this module is deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// Declaration is an interface for all declaration types
type Declaration interface {
	isDeclaration()
	GetName() string
}

// ClassDecl represents a class declaration
type ClassDecl struct {
	// Kind is "class"
	Kind string `json:"kind"`

	// Name of the class
	Name string `json:"name"`

	// Summary of the class
	Summary string `json:"summary,omitempty"`

	// Description of the class (usually from JSDoc)
	Description string `json:"description,omitempty"`

	// Superclass information
	Superclass *Reference `json:"superclass,omitempty"`

	// Mixins applied to this class
	Mixins []*Reference `json:"mixins,omitempty"`

	// Members of this class (fields, methods)
	Members []ClassMember `json:"members,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Whether this class is deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`

	// Type parameters
	TypeParameters []*TypeParameter `json:"typeParameters,omitempty"`
}

func (c *ClassDecl) isDeclaration() {}
func (c *ClassDecl) GetName() string { return c.Name }

// CustomElementDecl extends ClassDecl for custom elements
type CustomElementDecl struct {
	// Kind is "class"
	Kind string `json:"kind"`

	// Name of the class
	Name string `json:"name"`

	// Custom element tag name
	TagName string `json:"tagName,omitempty"`

	// Summary of the class
	Summary string `json:"summary,omitempty"`

	// Description of the class (usually from JSDoc)
	Description string `json:"description,omitempty"`

	// Superclass information
	Superclass *Reference `json:"superclass,omitempty"`

	// Mixins applied to this class
	Mixins []*Reference `json:"mixins,omitempty"`

	// Members of this class (fields, methods)
	Members []ClassMember `json:"members,omitempty"`

	// Attributes on this element
	Attributes []*Attribute `json:"attributes,omitempty"`

	// CSS custom properties
	CSSProperties []*CSSCustomProperty `json:"cssProperties,omitempty"`

	// CSS parts
	CSSParts []*CSSPart `json:"cssParts,omitempty"`

	// Slots
	Slots []*Slot `json:"slots,omitempty"`

	// Events
	Events []*Event `json:"events,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Whether this class is deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`

	// Type parameters
	TypeParameters []*TypeParameter `json:"typeParameters,omitempty"`

	// Customized built-in element
	CustomElement bool `json:"customElement"`
}

func (c *CustomElementDecl) isDeclaration() {}
func (c *CustomElementDecl) GetName() string { return c.Name }

// FunctionDecl represents a function declaration
type FunctionDecl struct {
	// Kind is "function"
	Kind string `json:"kind"`

	// Name of the function
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Parameters
	Parameters []*Parameter `json:"parameters,omitempty"`

	// Return type
	Return *ReturnType `json:"return,omitempty"`

	// Type parameters
	TypeParameters []*TypeParameter `json:"typeParameters,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

func (f *FunctionDecl) isDeclaration() {}
func (f *FunctionDecl) GetName() string { return f.Name }

// VariableDecl represents a variable declaration
type VariableDecl struct {
	// Kind is "variable"
	Kind string `json:"kind"`

	// Name of the variable
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Type information
	Type *Type `json:"type,omitempty"`

	// Default value
	Default string `json:"default,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`

	// Whether it's readonly
	Readonly bool `json:"readonly,omitempty"`
}

func (v *VariableDecl) isDeclaration() {}
func (v *VariableDecl) GetName() string { return v.Name }

// MixinDecl represents a mixin
type MixinDecl struct {
	// Kind is "mixin"
	Kind string `json:"kind"`

	// Name of the mixin
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Mixins applied to this mixin
	Mixins []*Reference `json:"mixins,omitempty"`

	// Members
	Members []ClassMember `json:"members,omitempty"`

	// Parameters
	Parameters []*Parameter `json:"parameters,omitempty"`

	// Type parameters
	TypeParameters []*TypeParameter `json:"typeParameters,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

func (m *MixinDecl) isDeclaration() {}
func (m *MixinDecl) GetName() string { return m.Name }

// ClassMember is an interface for class members
type ClassMember interface {
	isClassMember()
	GetMemberName() string
}

// ClassField represents a field/property in a class
type ClassField struct {
	// Kind is "field"
	Kind string `json:"kind"`

	// Name of the field
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Privacy level
	Privacy string `json:"privacy,omitempty"`

	// Type information
	Type *Type `json:"type,omitempty"`

	// Default value
	Default string `json:"default,omitempty"`

	// Whether static
	Static bool `json:"static,omitempty"`

	// Whether readonly
	Readonly bool `json:"readonly,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`

	// Associated attribute name (for custom elements)
	Attribute string `json:"attribute,omitempty"`

	// Whether it reflects to attribute
	Reflects bool `json:"reflects,omitempty"`
}

func (f *ClassField) isClassMember() {}
func (f *ClassField) GetMemberName() string { return f.Name }

// ClassMethod represents a method in a class
type ClassMethod struct {
	// Kind is "method"
	Kind string `json:"kind"`

	// Name of the method
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Privacy level
	Privacy string `json:"privacy,omitempty"`

	// Parameters
	Parameters []*Parameter `json:"parameters,omitempty"`

	// Return type
	Return *ReturnType `json:"return,omitempty"`

	// Type parameters
	TypeParameters []*TypeParameter `json:"typeParameters,omitempty"`

	// Whether static
	Static bool `json:"static,omitempty"`

	// Source location
	Source *SourceReference `json:"source,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

func (m *ClassMethod) isClassMember() {}
func (m *ClassMethod) GetMemberName() string { return m.Name }

// Reference represents a reference to another declaration
type Reference struct {
	// Name of the referenced declaration
	Name string `json:"name"`

	// Package containing the declaration
	Package string `json:"package,omitempty"`

	// Module containing the declaration
	Module string `json:"module,omitempty"`
}

// Type represents type information
type Type struct {
	// Text representation of the type
	Text string `json:"text"`

	// References in the type
	References []*TypeReference `json:"references,omitempty"`
}

// TypeReference represents a reference within a type
type TypeReference struct {
	// Name of the referenced type
	Name string `json:"name,omitempty"`

	// Package containing the type
	Package string `json:"package,omitempty"`

	// Module containing the type
	Module string `json:"module,omitempty"`

	// Start position in text
	Start int `json:"start,omitempty"`

	// End position in text
	End int `json:"end,omitempty"`
}

// TypeParameter represents a type parameter
type TypeParameter struct {
	// Name of the type parameter
	Name string `json:"name"`

	// Default value
	Default string `json:"default,omitempty"`

	// Extends constraint
	Extends string `json:"extends,omitempty"`

	// Description
	Description string `json:"description,omitempty"`
}

// Parameter represents a function/method parameter
type Parameter struct {
	// Name of the parameter
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Type information
	Type *Type `json:"type,omitempty"`

	// Default value
	Default string `json:"default,omitempty"`

	// Whether optional
	Optional bool `json:"optional,omitempty"`

	// Whether rest parameter
	Rest bool `json:"rest,omitempty"`
}

// ReturnType represents return type information
type ReturnType struct {
	// Type information
	Type *Type `json:"type,omitempty"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`
}

// Attribute represents a custom element attribute
type Attribute struct {
	// Name of the attribute
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Type information
	Type *Type `json:"type,omitempty"`

	// Default value
	Default string `json:"default,omitempty"`

	// Field name this attribute is associated with
	FieldName string `json:"fieldName,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// CSSCustomProperty represents a CSS custom property
type CSSCustomProperty struct {
	// Name of the property (including --)
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Default value
	Default string `json:"default,omitempty"`

	// Syntax (e.g., "<color>")
	Syntax string `json:"syntax,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// CSSPart represents a CSS part
type CSSPart struct {
	// Name of the part
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// Slot represents a slot
type Slot struct {
	// Name of the slot (empty for default)
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// Event represents a custom element event
type Event struct {
	// Name of the event
	Name string `json:"name"`

	// Summary
	Summary string `json:"summary,omitempty"`

	// Description
	Description string `json:"description,omitempty"`

	// Type of the event
	Type *Type `json:"type,omitempty"`

	// Inherited from
	InheritedFrom *Reference `json:"inheritedFrom,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// Export represents an export from a module
type Export struct {
	// Kind of export
	Kind string `json:"kind"`

	// Name of the export
	Name string `json:"name"`

	// Declaration reference
	Declaration *Reference `json:"declaration,omitempty"`

	// Whether deprecated
	Deprecated interface{} `json:"deprecated,omitempty"`
}

// SourceReference represents a source code location
type SourceReference struct {
	// Href to the source
	Href string `json:"href,omitempty"`
}

// NewManifest creates a new empty manifest
func NewManifest() *Manifest {
	return &Manifest{
		SchemaVersion: "2.1.0",
		Modules:       make([]*Module, 0),
	}
}

// NewModule creates a new module
func NewModule(path string) *Module {
	return &Module{
		Kind:         "javascript-module",
		Path:         path,
		Declarations: make([]Declaration, 0),
		Exports:      make([]*Export, 0),
	}
}
