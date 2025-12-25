package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
)

// Analyzer extracts Custom Element Manifest information from TypeScript source files
type Analyzer struct {
	// Map of class name to class info for inheritance resolution
	classes map[string]*ClassInfo

	// Map of module path to module
	modules map[string]*Module

	// Base directory for relative paths
	baseDir string

	// Current source file being analyzed (for JSDoc access)
	currentSourceFile *ast.SourceFile
}

// ClassInfo stores class information for inheritance resolution
type ClassInfo struct {
	Name        string
	SuperClass  string
	Members     []ClassMember
	ModulePath  string
	Description string
	TagName     string
	IsElement   bool
	Attributes  []*Attribute
	Events      []*Event
	Slots       []*Slot
	CSSProps    []*CSSCustomProperty
	CSSParts    []*CSSPart
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(baseDir string) *Analyzer {
	return &Analyzer{
		classes: make(map[string]*ClassInfo),
		modules: make(map[string]*Module),
		baseDir: baseDir,
	}
}

// AnalyzeFile parses and analyzes a single TypeScript file
func (a *Analyzer) AnalyzeFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(a.baseDir, filePath)
	if err != nil {
		relPath = filePath
	}

	opts := ast.SourceFileParseOptions{
		FileName:         filePath,
		JSDocParsingMode: ast.JSDocParsingModeParseAll, // Parse all JSDoc comments
	}

	// Determine script kind based on file extension
	scriptKind := core.ScriptKindTS
	if strings.HasSuffix(filePath, ".tsx") {
		scriptKind = core.ScriptKindTSX
	} else if strings.HasSuffix(filePath, ".js") {
		scriptKind = core.ScriptKindJS
	} else if strings.HasSuffix(filePath, ".jsx") {
		scriptKind = core.ScriptKindJSX
	}

	sourceFile := parser.ParseSourceFile(opts, string(content), scriptKind)
	a.currentSourceFile = sourceFile

	module := NewModule(relPath)
	a.modules[relPath] = module

	a.visitSourceFile(sourceFile, module)

	return nil
}

// visitSourceFile visits all statements in a source file
func (a *Analyzer) visitSourceFile(sf *ast.SourceFile, module *Module) {
	if sf.Statements == nil {
		return
	}

	for _, stmt := range sf.Statements.Nodes {
		a.visitStatement(stmt, module)
	}
}

// visitStatement processes a single statement
func (a *Analyzer) visitStatement(node *ast.Node, module *Module) {
	switch node.Kind {
	case ast.KindClassDeclaration:
		a.visitClassDeclaration(node, module)
	case ast.KindFunctionDeclaration:
		a.visitFunctionDeclaration(node, module)
	case ast.KindVariableStatement:
		a.visitVariableStatement(node, module)
	case ast.KindExportDeclaration:
		a.visitExportDeclaration(node, module)
	case ast.KindExportAssignment:
		a.visitExportAssignment(node, module)
	}
}

// visitClassDeclaration extracts class information
func (a *Analyzer) visitClassDeclaration(node *ast.Node, module *Module) {
	classDecl := node.AsClassDeclaration()
	if classDecl == nil {
		return
	}

	name := ""
	if classDecl.Name() != nil {
		name = classDecl.Name().Text()
	}

	if name == "" {
		return
	}

	classInfo := &ClassInfo{
		Name:       name,
		ModulePath: module.Path,
		Members:    make([]ClassMember, 0),
	}

	// Extract JSDoc description using AST
	classInfo.Description = a.extractJSDocDescription(node)

	// Extract tag name from JSDoc @customElement or decorator
	classInfo.TagName = a.extractTagName(node, name)
	classInfo.IsElement = classInfo.TagName != "" || a.extendsHTMLElement(classDecl)

	// Extract superclass
	if classDecl.HeritageClauses != nil {
		for _, clause := range classDecl.HeritageClauses.Nodes {
			hc := clause.AsHeritageClause()
			if hc != nil && hc.Token == ast.KindExtendsKeyword && hc.Types != nil {
				for _, typeExpr := range hc.Types.Nodes {
					classInfo.SuperClass = a.getExpressionText(typeExpr)
					break
				}
			}
		}
	}

	// Extract JSDoc tags for slots, events, CSS properties (using AST)
	a.extractJSDocTagsAST(node, classInfo)

	// Extract members
	if classDecl.Members != nil {
		for _, member := range classDecl.Members.Nodes {
			a.visitClassMember(member, classInfo)
		}
	}

	a.classes[name] = classInfo

	// Create declaration
	if classInfo.IsElement {
		decl := a.createCustomElementDeclaration(classInfo)
		module.Declarations = append(module.Declarations, decl)

		// Add export if it's exported
		if a.isExported(node) && classInfo.TagName != "" {
			module.Exports = append(module.Exports, &Export{
				Kind: "custom-element-definition",
				Name: classInfo.TagName,
				Declaration: &Reference{
					Name:   name,
					Module: module.Path,
				},
			})
		}
	} else {
		decl := a.createClassDeclaration(classInfo)
		module.Declarations = append(module.Declarations, decl)
	}

	// Add class export if exported
	if a.isExported(node) {
		module.Exports = append(module.Exports, &Export{
			Kind: "js",
			Name: name,
			Declaration: &Reference{
				Name:   name,
				Module: module.Path,
			},
		})
	}
}

// visitClassMember extracts class member information
func (a *Analyzer) visitClassMember(node *ast.Node, classInfo *ClassInfo) {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		a.visitPropertyDeclaration(node, classInfo)
	case ast.KindMethodDeclaration:
		a.visitMethodDeclaration(node, classInfo)
	case ast.KindGetAccessor:
		a.visitGetAccessor(node, classInfo)
	case ast.KindSetAccessor:
		a.visitSetAccessor(node, classInfo)
	case ast.KindConstructor:
		a.visitConstructor(node, classInfo)
	}
}

// visitPropertyDeclaration extracts property information
func (a *Analyzer) visitPropertyDeclaration(node *ast.Node, classInfo *ClassInfo) {
	propDecl := node.AsPropertyDeclaration()
	if propDecl == nil {
		return
	}

	name := a.getPropertyName(propDecl.Name())
	if name == "" {
		return
	}

	field := &ClassField{
		Kind:        "field",
		Name:        name,
		Description: a.extractJSDocDescription(node),
		Privacy:     a.getPrivacy(node),
		Static:      a.hasModifier(node, ast.KindStaticKeyword),
		Readonly:    a.hasModifier(node, ast.KindReadonlyKeyword),
	}

	// Check for deprecated via JSDoc
	if a.hasJSDocTag(node, "deprecated") {
		field.Deprecated = true
	}

	// Extract type
	if propDecl.Type != nil {
		field.Type = &Type{Text: a.getTypeText(propDecl.Type)}
	}

	// Extract default value
	if propDecl.Initializer != nil {
		field.Default = a.getExpressionText(propDecl.Initializer)
	}

	// Check for attribute reflection (from decorators or JSDoc @attr)
	field.Attribute = a.extractAttributeName(node, name)
	field.Reflects = field.Attribute != ""

	classInfo.Members = append(classInfo.Members, field)

	// Add to attributes if it has an attribute
	if field.Attribute != "" {
		classInfo.Attributes = append(classInfo.Attributes, &Attribute{
			Name:        field.Attribute,
			Description: field.Description,
			Type:        field.Type,
			Default:     field.Default,
			FieldName:   name,
		})
	}
}

// visitMethodDeclaration extracts method information
func (a *Analyzer) visitMethodDeclaration(node *ast.Node, classInfo *ClassInfo) {
	methodDecl := node.AsMethodDeclaration()
	if methodDecl == nil {
		return
	}

	name := a.getPropertyName(methodDecl.Name())
	if name == "" {
		return
	}

	method := &ClassMethod{
		Kind:        "method",
		Name:        name,
		Description: a.extractJSDocDescription(node),
		Privacy:     a.getPrivacy(node),
		Static:      a.hasModifier(node, ast.KindStaticKeyword),
		Parameters:  a.extractParameters(node, methodDecl.Parameters),
	}

	// Check for deprecated via JSDoc
	if a.hasJSDocTag(node, "deprecated") {
		method.Deprecated = true
	}

	// Extract return type
	if methodDecl.ReturnType() != nil {
		method.Return = &ReturnType{
			Type: &Type{Text: a.getTypeText(methodDecl.ReturnType())},
		}
	}

	// Extract return description from JSDoc @returns tag
	if returnDesc := a.getJSDocTagComment(node, "returns", "return"); returnDesc != "" {
		if method.Return == nil {
			method.Return = &ReturnType{}
		}
		method.Return.Description = returnDesc
	}

	// Extract type parameters
	if methodDecl.TypeParameters != nil {
		method.TypeParameters = a.extractTypeParameters(methodDecl.TypeParameters)
	}

	classInfo.Members = append(classInfo.Members, method)
}

// visitGetAccessor extracts getter information
func (a *Analyzer) visitGetAccessor(node *ast.Node, classInfo *ClassInfo) {
	getAccessor := node.AsGetAccessorDeclaration()
	if getAccessor == nil {
		return
	}

	name := a.getPropertyName(getAccessor.Name())
	if name == "" {
		return
	}

	// Check if we already have a field with this name (from setter)
	for _, member := range classInfo.Members {
		if field, ok := member.(*ClassField); ok && field.Name == name {
			// Update existing field
			if getAccessor.ReturnType() != nil && field.Type == nil {
				field.Type = &Type{Text: a.getTypeText(getAccessor.ReturnType())}
			}
			return
		}
	}

	field := &ClassField{
		Kind:        "field",
		Name:        name,
		Description: a.extractJSDocDescription(node),
		Privacy:     a.getPrivacy(node),
		Static:      a.hasModifier(node, ast.KindStaticKeyword),
		Readonly:    true, // getters without setters are effectively readonly
	}

	if getAccessor.ReturnType() != nil {
		field.Type = &Type{Text: a.getTypeText(getAccessor.ReturnType())}
	}

	// Check for attribute from JSDoc @attr tag
	field.Attribute = a.extractAttributeName(node, name)
	field.Reflects = field.Attribute != ""

	classInfo.Members = append(classInfo.Members, field)

	if field.Attribute != "" {
		classInfo.Attributes = append(classInfo.Attributes, &Attribute{
			Name:        field.Attribute,
			Description: field.Description,
			Type:        field.Type,
			FieldName:   name,
		})
	}
}

// visitSetAccessor extracts setter information
func (a *Analyzer) visitSetAccessor(node *ast.Node, classInfo *ClassInfo) {
	setAccessor := node.AsSetAccessorDeclaration()
	if setAccessor == nil {
		return
	}

	name := a.getPropertyName(setAccessor.Name())
	if name == "" {
		return
	}

	// Check if we already have a field with this name (from getter)
	for _, member := range classInfo.Members {
		if field, ok := member.(*ClassField); ok && field.Name == name {
			field.Readonly = false // has setter, so not readonly
			return
		}
	}

	field := &ClassField{
		Kind:        "field",
		Name:        name,
		Description: a.extractJSDocDescription(node),
		Privacy:     a.getPrivacy(node),
		Static:      a.hasModifier(node, ast.KindStaticKeyword),
	}

	// Get type from parameter
	if setAccessor.Parameters != nil && len(setAccessor.Parameters.Nodes) > 0 {
		param := setAccessor.Parameters.Nodes[0].AsParameterDeclaration()
		if param != nil && param.Type != nil {
			field.Type = &Type{Text: a.getTypeText(param.Type)}
		}
	}

	classInfo.Members = append(classInfo.Members, field)
}

// visitConstructor extracts constructor parameter properties
func (a *Analyzer) visitConstructor(node *ast.Node, classInfo *ClassInfo) {
	ctor := node.AsConstructorDeclaration()
	if ctor == nil || ctor.Parameters == nil {
		return
	}

	for _, paramNode := range ctor.Parameters.Nodes {
		param := paramNode.AsParameterDeclaration()
		if param == nil {
			continue
		}

		// Check if parameter is a parameter property (has public/private/protected/readonly modifier)
		if !a.isParameterProperty(paramNode) {
			continue
		}

		name := ""
		if param.Name() != nil {
			name = a.getBindingName(param.Name())
		}

		if name == "" {
			continue
		}

		field := &ClassField{
			Kind:        "field",
			Name:        name,
			Description: a.extractJSDocDescription(paramNode),
			Privacy:     a.getPrivacy(paramNode),
			Readonly:    a.hasModifier(paramNode, ast.KindReadonlyKeyword),
		}

		if param.Type != nil {
			field.Type = &Type{Text: a.getTypeText(param.Type)}
		}

		if param.Initializer != nil {
			field.Default = a.getExpressionText(param.Initializer)
		}

		classInfo.Members = append(classInfo.Members, field)
	}
}

// visitFunctionDeclaration extracts function information
func (a *Analyzer) visitFunctionDeclaration(node *ast.Node, module *Module) {
	funcDecl := node.AsFunctionDeclaration()
	if funcDecl == nil {
		return
	}

	name := ""
	if funcDecl.Name() != nil {
		name = funcDecl.Name().Text()
	}

	if name == "" {
		return
	}

	decl := &FunctionDecl{
		Kind:        "function",
		Name:        name,
		Description: a.extractJSDocDescription(node),
		Parameters:  a.extractParameters(node, funcDecl.Parameters),
	}

	if a.hasJSDocTag(node, "deprecated") {
		decl.Deprecated = true
	}

	if funcDecl.ReturnType() != nil {
		decl.Return = &ReturnType{
			Type: &Type{Text: a.getTypeText(funcDecl.ReturnType())},
		}
	}

	if returnDesc := a.getJSDocTagComment(node, "returns", "return"); returnDesc != "" {
		if decl.Return == nil {
			decl.Return = &ReturnType{}
		}
		decl.Return.Description = returnDesc
	}

	if funcDecl.TypeParameters != nil {
		decl.TypeParameters = a.extractTypeParameters(funcDecl.TypeParameters)
	}

	module.Declarations = append(module.Declarations, decl)

	if a.isExported(node) {
		module.Exports = append(module.Exports, &Export{
			Kind: "js",
			Name: name,
			Declaration: &Reference{
				Name:   name,
				Module: module.Path,
			},
		})
	}
}

// visitVariableStatement extracts variable declarations
func (a *Analyzer) visitVariableStatement(node *ast.Node, module *Module) {
	varStmt := node.AsVariableStatement()
	if varStmt == nil || varStmt.DeclarationList == nil {
		return
	}

	declList := varStmt.DeclarationList.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}

	for _, declNode := range declList.Declarations.Nodes {
		varDecl := declNode.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}

		name := a.getBindingName(varDecl.Name())
		if name == "" {
			continue
		}

		decl := &VariableDecl{
			Kind:        "variable",
			Name:        name,
			Description: a.extractJSDocDescription(declNode),
			Readonly:    declList.Flags&ast.NodeFlagsConst != 0,
		}

		if a.hasJSDocTag(declNode, "deprecated") {
			decl.Deprecated = true
		}

		if varDecl.Type != nil {
			decl.Type = &Type{Text: a.getTypeText(varDecl.Type)}
		}

		if varDecl.Initializer != nil {
			decl.Default = a.getExpressionText(varDecl.Initializer)
		}

		module.Declarations = append(module.Declarations, decl)

		if a.isExported(node) {
			module.Exports = append(module.Exports, &Export{
				Kind: "js",
				Name: name,
				Declaration: &Reference{
					Name:   name,
					Module: module.Path,
				},
			})
		}
	}
}

// visitExportDeclaration handles export declarations
func (a *Analyzer) visitExportDeclaration(node *ast.Node, module *Module) {
	exportDecl := node.AsExportDeclaration()
	if exportDecl == nil {
		return
	}

	// Handle named exports
	if exportDecl.ExportClause != nil {
		switch exportDecl.ExportClause.Kind {
		case ast.KindNamedExports:
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports != nil && namedExports.Elements != nil {
				for _, elem := range namedExports.Elements.Nodes {
					specifier := elem.AsExportSpecifier()
					if specifier == nil {
						continue
					}

					exportName := ""
					localName := ""

					if specifier.Name() != nil {
						exportName = specifier.Name().Text()
					}
					if specifier.PropertyName != nil {
						localName = specifier.PropertyName.Text()
					} else {
						localName = exportName
					}

					module.Exports = append(module.Exports, &Export{
						Kind: "js",
						Name: exportName,
						Declaration: &Reference{
							Name:   localName,
							Module: module.Path,
						},
					})
				}
			}
		}
	}
}

// visitExportAssignment handles default exports
func (a *Analyzer) visitExportAssignment(node *ast.Node, module *Module) {
	exportAssign := node.AsExportAssignment()
	if exportAssign == nil || exportAssign.Expression == nil {
		return
	}

	name := a.getExpressionText(exportAssign.Expression)
	if name == "" {
		name = "default"
	}

	module.Exports = append(module.Exports, &Export{
		Kind: "js",
		Name: "default",
		Declaration: &Reference{
			Name:   name,
			Module: module.Path,
		},
	})
}

// extractParameters extracts parameter information from a parameter list
// Uses JSDoc @param tags for descriptions
func (a *Analyzer) extractParameters(node *ast.Node, params *ast.NodeList) []*Parameter {
	if params == nil {
		return nil
	}

	// Build a map of parameter descriptions from JSDoc @param tags
	paramDescs := make(map[string]string)
	jsdocs := node.JSDoc(a.currentSourceFile)
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil || jsDocNode.Tags == nil {
			continue
		}
		for _, tag := range jsDocNode.Tags.Nodes {
			if tag.Kind == ast.KindJSDocParameterTag {
				paramTag := tag.AsJSDocParameterOrPropertyTag()
				if paramTag != nil && paramTag.Name() != nil {
					paramName := a.getEntityName(paramTag.Name())
					paramDescs[paramName] = a.getJSDocCommentText(paramTag.Comment)
				}
			}
		}
	}

	result := make([]*Parameter, 0, len(params.Nodes))
	for _, paramNode := range params.Nodes {
		param := paramNode.AsParameterDeclaration()
		if param == nil {
			continue
		}

		name := a.getBindingName(param.Name())
		if name == "" {
			continue
		}

		p := &Parameter{
			Name:        name,
			Description: paramDescs[name],
			Optional:    param.PostfixToken != nil && param.PostfixToken.Kind == ast.KindQuestionToken,
			Rest:        param.DotDotDotToken != nil,
		}

		if param.Type != nil {
			p.Type = &Type{Text: a.getTypeText(param.Type)}
		}

		if param.Initializer != nil {
			p.Default = a.getExpressionText(param.Initializer)
			p.Optional = true
		}

		result = append(result, p)
	}

	return result
}

// extractTypeParameters extracts type parameter information
func (a *Analyzer) extractTypeParameters(typeParams *ast.NodeList) []*TypeParameter {
	if typeParams == nil {
		return nil
	}

	result := make([]*TypeParameter, 0, len(typeParams.Nodes))
	for _, tpNode := range typeParams.Nodes {
		tp := tpNode.AsTypeParameter()
		if tp == nil {
			continue
		}

		param := &TypeParameter{
			Name: tp.Name().Text(),
		}

		if tp.Constraint != nil {
			param.Extends = a.getTypeText(tp.Constraint)
		}

		if tp.Default != nil {
			param.Default = a.getTypeText(tp.Default)
		}

		result = append(result, param)
	}

	return result
}

// extractJSDocDescription extracts the main description from JSDoc comments
func (a *Analyzer) extractJSDocDescription(node *ast.Node) string {
	jsdocs := node.JSDoc(a.currentSourceFile)
	if len(jsdocs) == 0 {
		return ""
	}

	var parts []string
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil {
			continue
		}

		// Get the comment (main description)
		if jsDocNode.Comment != nil {
			text := a.getJSDocCommentText(jsDocNode.Comment)
			if text != "" {
				parts = append(parts, text)
			}
		}
	}

	return strings.Join(parts, "\n")
}

// getJSDocCommentText extracts text from a JSDoc comment node list
func (a *Analyzer) getJSDocCommentText(comment *ast.NodeList) string {
	if comment == nil {
		return ""
	}

	var parts []string
	for _, node := range comment.Nodes {
		switch node.Kind {
		case ast.KindJSDocText:
			textNode := node.AsJSDocText()
			if textNode != nil {
				parts = append(parts, node.Text())
			}
		case ast.KindJSDocLink, ast.KindJSDocLinkCode, ast.KindJSDocLinkPlain:
			parts = append(parts, node.Text())
		}
	}

	return strings.TrimSpace(strings.Join(parts, ""))
}

// hasJSDocTag checks if a node has a specific JSDoc tag
func (a *Analyzer) hasJSDocTag(node *ast.Node, tagNames ...string) bool {
	jsdocs := node.JSDoc(a.currentSourceFile)
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil || jsDocNode.Tags == nil {
			continue
		}
		for _, tag := range jsDocNode.Tags.Nodes {
			tagName := tag.TagName()
			if tagName != nil {
				for _, name := range tagNames {
					if tagName.Text == name {
						return true
					}
				}
			}
		}
	}
	return false
}

// getJSDocTagComment gets the comment text from a specific JSDoc tag
func (a *Analyzer) getJSDocTagComment(node *ast.Node, tagNames ...string) string {
	jsdocs := node.JSDoc(a.currentSourceFile)
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil || jsDocNode.Tags == nil {
			continue
		}
		for _, tag := range jsDocNode.Tags.Nodes {
			tagName := tag.TagName()
			if tagName != nil {
				for _, name := range tagNames {
					if tagName.Text == name {
						return a.getJSDocCommentText(tag.Comment())
					}
				}
			}
		}
	}
	return ""
}

// extractJSDocTagsAST extracts special JSDoc tags for web components using AST
func (a *Analyzer) extractJSDocTagsAST(node *ast.Node, classInfo *ClassInfo) {
	jsdocs := node.JSDoc(a.currentSourceFile)
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil || jsDocNode.Tags == nil {
			continue
		}

		for _, tag := range jsDocNode.Tags.Nodes {
			tagName := tag.TagName()
			if tagName == nil {
				continue
			}

			comment := a.getJSDocCommentText(tag.Comment())

			switch tagName.Text {
			case "fires", "event":
				// @fires event-name - description
				if comment != "" {
					parts := strings.SplitN(comment, " ", 2)
					eventName := parts[0]
					eventDesc := ""
					if len(parts) > 1 {
						eventDesc = strings.TrimPrefix(parts[1], "- ")
					}
					classInfo.Events = append(classInfo.Events, &Event{
						Name:        eventName,
						Description: eventDesc,
					})
				}

			case "slot":
				// @slot name - description or @slot - description (default slot)
				if comment != "" {
					parts := strings.SplitN(comment, " ", 2)
					slotName := ""
					slotDesc := comment
					if len(parts) > 1 && parts[0] != "-" {
						slotName = parts[0]
						slotDesc = strings.TrimPrefix(parts[1], "- ")
					}
					classInfo.Slots = append(classInfo.Slots, &Slot{
						Name:        slotName,
						Description: slotDesc,
					})
				}

			case "cssProperty", "cssproperty", "cssProperty":
				// @cssProperty --prop-name - description
				if comment != "" {
					parts := strings.SplitN(comment, " ", 2)
					propName := parts[0]
					propDesc := ""
					if len(parts) > 1 {
						propDesc = strings.TrimPrefix(parts[1], "- ")
					}
					classInfo.CSSProps = append(classInfo.CSSProps, &CSSCustomProperty{
						Name:        propName,
						Description: propDesc,
					})
				}

			case "cssPart", "csspart":
				// @cssPart part-name - description
				if comment != "" {
					parts := strings.SplitN(comment, " ", 2)
					partName := parts[0]
					partDesc := ""
					if len(parts) > 1 {
						partDesc = strings.TrimPrefix(parts[1], "- ")
					}
					classInfo.CSSParts = append(classInfo.CSSParts, &CSSPart{
						Name:        partName,
						Description: partDesc,
					})
				}

			case "customElement":
				// @customElement tag-name
				if comment != "" {
					classInfo.TagName = strings.TrimSpace(comment)
				}
			}
		}
	}
}

// extractTagName extracts custom element tag name from decorators or JSDoc
func (a *Analyzer) extractTagName(node *ast.Node, className string) string {
	// Check decorators for @customElement('tag-name')
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindDecorator {
				decorator := mod.AsDecorator()
				if decorator.Expression != nil && decorator.Expression.Kind == ast.KindCallExpression {
					call := decorator.Expression.AsCallExpression()
					exprName := a.getExpressionText(call.Expression)
					if exprName == "customElement" || exprName == "define" {
						if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
							arg := call.Arguments.Nodes[0]
							if arg.Kind == ast.KindStringLiteral {
								return arg.AsStringLiteral().Text
							}
						}
					}
				}
			}
		}
	}

	// Check JSDoc @customElement tag (already extracted in extractJSDocTagsAST)
	// Return empty to let the caller check classInfo.TagName
	return ""
}

// extractAttributeName extracts attribute name from decorators or JSDoc @attr
func (a *Analyzer) extractAttributeName(node *ast.Node, propName string) string {
	// Check for @attr or @attribute JSDoc tag using AST
	jsdocs := node.JSDoc(a.currentSourceFile)
	for _, jsdoc := range jsdocs {
		jsDocNode := jsdoc.AsJSDoc()
		if jsDocNode == nil || jsDocNode.Tags == nil {
			continue
		}
		for _, tag := range jsDocNode.Tags.Nodes {
			tagName := tag.TagName()
			if tagName != nil && (tagName.Text == "attr" || tagName.Text == "attribute") {
				comment := a.getJSDocCommentText(tag.Comment())
				if comment != "" {
					return strings.TrimSpace(comment)
				}
				return toKebabCase(propName)
			}
		}
	}

	// Check @property decorator with attribute option
	modifiers := node.Modifiers()
	if modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindDecorator {
				decorator := mod.AsDecorator()
				if decorator.Expression != nil && decorator.Expression.Kind == ast.KindCallExpression {
					call := decorator.Expression.AsCallExpression()
					exprName := a.getExpressionText(call.Expression)
					if exprName == "property" {
						// Has @property decorator, check for attribute: false
						if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
							arg := call.Arguments.Nodes[0]
							if arg.Kind == ast.KindObjectLiteralExpression {
								objLit := arg.AsObjectLiteralExpression()
								if objLit.Properties != nil {
									for _, prop := range objLit.Properties.Nodes {
										if prop.Kind == ast.KindPropertyAssignment {
											propAssign := prop.AsPropertyAssignment()
											propKey := a.getPropertyName(propAssign.Name())
											if propKey == "attribute" {
												if propAssign.Initializer != nil &&
													propAssign.Initializer.Kind == ast.KindFalseKeyword {
													return "" // attribute: false
												}
											}
										}
									}
								}
							}
						}
						return toKebabCase(propName)
					}
				}
			}
		}
	}

	return ""
}

// createClassDeclaration creates a ClassDecl from ClassInfo
func (a *Analyzer) createClassDeclaration(info *ClassInfo) *ClassDecl {
	decl := &ClassDecl{
		Kind:        "class",
		Name:        info.Name,
		Description: info.Description,
		Members:     info.Members,
	}

	if info.SuperClass != "" {
		decl.Superclass = &Reference{Name: info.SuperClass}
	}

	return decl
}

// createCustomElementDeclaration creates a CustomElementDecl from ClassInfo
func (a *Analyzer) createCustomElementDeclaration(info *ClassInfo) *CustomElementDecl {
	decl := &CustomElementDecl{
		Kind:          "class",
		Name:          info.Name,
		TagName:       info.TagName,
		Description:   info.Description,
		Members:       info.Members,
		Attributes:    info.Attributes,
		Events:        info.Events,
		Slots:         info.Slots,
		CSSProperties: info.CSSProps,
		CSSParts:      info.CSSParts,
		CustomElement: true,
	}

	if info.SuperClass != "" {
		decl.Superclass = &Reference{Name: info.SuperClass}
	}

	return decl
}

// ResolveInheritance resolves class inheritance and copies inherited members
func (a *Analyzer) ResolveInheritance() {
	// Build inheritance chain for each class
	for _, classInfo := range a.classes {
		a.resolveClassInheritance(classInfo, make(map[string]bool))
	}

	// Update declarations in modules with resolved inheritance
	for _, module := range a.modules {
		for i, decl := range module.Declarations {
			switch d := decl.(type) {
			case *ClassDecl:
				if info, ok := a.classes[d.Name]; ok {
					module.Declarations[i] = a.createClassDeclaration(info)
				}
			case *CustomElementDecl:
				if info, ok := a.classes[d.Name]; ok {
					module.Declarations[i] = a.createCustomElementDeclaration(info)
				}
			}
		}
	}
}

// resolveClassInheritance recursively resolves inheritance for a class
func (a *Analyzer) resolveClassInheritance(classInfo *ClassInfo, visited map[string]bool) {
	if visited[classInfo.Name] {
		return // Prevent circular inheritance
	}
	visited[classInfo.Name] = true

	if classInfo.SuperClass == "" {
		return
	}

	superClass, ok := a.classes[classInfo.SuperClass]
	if !ok {
		return // Superclass not found in analyzed files
	}

	// Resolve superclass first
	a.resolveClassInheritance(superClass, visited)

	// Copy inherited members
	existingMembers := make(map[string]bool)
	for _, member := range classInfo.Members {
		existingMembers[member.GetMemberName()] = true
	}

	for _, member := range superClass.Members {
		if existingMembers[member.GetMemberName()] {
			continue // Don't copy overridden members
		}

		// Clone and mark as inherited
		inheritedMember := a.cloneMemberWithInheritance(member, superClass.Name, superClass.ModulePath)
		classInfo.Members = append(classInfo.Members, inheritedMember)
	}

	// Copy inherited attributes
	existingAttrs := make(map[string]bool)
	for _, attr := range classInfo.Attributes {
		existingAttrs[attr.Name] = true
	}

	for _, attr := range superClass.Attributes {
		if existingAttrs[attr.Name] {
			continue
		}
		inheritedAttr := *attr
		inheritedAttr.InheritedFrom = &Reference{
			Name:   superClass.Name,
			Module: superClass.ModulePath,
		}
		classInfo.Attributes = append(classInfo.Attributes, &inheritedAttr)
	}

	// Copy inherited events
	existingEvents := make(map[string]bool)
	for _, event := range classInfo.Events {
		existingEvents[event.Name] = true
	}

	for _, event := range superClass.Events {
		if existingEvents[event.Name] {
			continue
		}
		inheritedEvent := *event
		inheritedEvent.InheritedFrom = &Reference{
			Name:   superClass.Name,
			Module: superClass.ModulePath,
		}
		classInfo.Events = append(classInfo.Events, &inheritedEvent)
	}

	// Copy inherited slots
	existingSlots := make(map[string]bool)
	for _, slot := range classInfo.Slots {
		existingSlots[slot.Name] = true
	}

	for _, slot := range superClass.Slots {
		if existingSlots[slot.Name] {
			continue
		}
		inheritedSlot := *slot
		inheritedSlot.InheritedFrom = &Reference{
			Name:   superClass.Name,
			Module: superClass.ModulePath,
		}
		classInfo.Slots = append(classInfo.Slots, &inheritedSlot)
	}

	// Copy inherited CSS properties
	existingCSSProps := make(map[string]bool)
	for _, prop := range classInfo.CSSProps {
		existingCSSProps[prop.Name] = true
	}

	for _, prop := range superClass.CSSProps {
		if existingCSSProps[prop.Name] {
			continue
		}
		inheritedProp := *prop
		inheritedProp.InheritedFrom = &Reference{
			Name:   superClass.Name,
			Module: superClass.ModulePath,
		}
		classInfo.CSSProps = append(classInfo.CSSProps, &inheritedProp)
	}

	// Copy inherited CSS parts
	existingCSSParts := make(map[string]bool)
	for _, part := range classInfo.CSSParts {
		existingCSSParts[part.Name] = true
	}

	for _, part := range superClass.CSSParts {
		if existingCSSParts[part.Name] {
			continue
		}
		inheritedPart := *part
		inheritedPart.InheritedFrom = &Reference{
			Name:   superClass.Name,
			Module: superClass.ModulePath,
		}
		classInfo.CSSParts = append(classInfo.CSSParts, &inheritedPart)
	}
}

// cloneMemberWithInheritance clones a member and marks it as inherited
func (a *Analyzer) cloneMemberWithInheritance(member ClassMember, className, modulePath string) ClassMember {
	ref := &Reference{
		Name:   className,
		Module: modulePath,
	}

	switch m := member.(type) {
	case *ClassField:
		cloned := *m
		cloned.InheritedFrom = ref
		return &cloned
	case *ClassMethod:
		cloned := *m
		cloned.InheritedFrom = ref
		return &cloned
	}

	return member
}

// GetManifest returns the complete manifest with all analyzed modules
func (a *Analyzer) GetManifest() *Manifest {
	manifest := NewManifest()

	for _, module := range a.modules {
		manifest.Modules = append(manifest.Modules, module)
	}

	return manifest
}

// Helper functions

func (a *Analyzer) getPropertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().Text
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text
	}

	return ""
}

func (a *Analyzer) getBindingName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		return "" // Complex binding patterns not supported
	}

	return ""
}

func (a *Analyzer) getExpressionText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return "\"" + node.AsStringLiteral().Text + "\""
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().Text
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		return a.getExpressionText(propAccess.Expression) + "." + propAccess.Name().Text()
	case ast.KindExpressionWithTypeArguments:
		exprWithType := node.AsExpressionWithTypeArguments()
		return a.getExpressionText(exprWithType.Expression)
	}

	return ""
}

func (a *Analyzer) getTypeText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindStringKeyword:
		return "string"
	case ast.KindNumberKeyword:
		return "number"
	case ast.KindBooleanKeyword:
		return "boolean"
	case ast.KindAnyKeyword:
		return "any"
	case ast.KindVoidKeyword:
		return "void"
	case ast.KindNeverKeyword:
		return "never"
	case ast.KindUnknownKeyword:
		return "unknown"
	case ast.KindUndefinedKeyword:
		return "undefined"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindObjectKeyword:
		return "object"
	case ast.KindSymbolKeyword:
		return "symbol"
	case ast.KindBigIntKeyword:
		return "bigint"
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		typeName := a.getEntityName(typeRef.TypeName)
		if typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0 {
			args := make([]string, 0, len(typeRef.TypeArguments.Nodes))
			for _, arg := range typeRef.TypeArguments.Nodes {
				args = append(args, a.getTypeText(arg))
			}
			typeName += "<" + strings.Join(args, ", ") + ">"
		}
		return typeName
	case ast.KindArrayType:
		arrayType := node.AsArrayTypeNode()
		return a.getTypeText(arrayType.ElementType) + "[]"
	case ast.KindUnionType:
		unionType := node.AsUnionTypeNode()
		types := make([]string, 0, len(unionType.Types.Nodes))
		for _, t := range unionType.Types.Nodes {
			types = append(types, a.getTypeText(t))
		}
		return strings.Join(types, " | ")
	case ast.KindIntersectionType:
		interType := node.AsIntersectionTypeNode()
		types := make([]string, 0, len(interType.Types.Nodes))
		for _, t := range interType.Types.Nodes {
			types = append(types, a.getTypeText(t))
		}
		return strings.Join(types, " & ")
	case ast.KindLiteralType:
		literalType := node.AsLiteralTypeNode()
		return a.getExpressionText(literalType.Literal)
	case ast.KindFunctionType:
		return "Function"
	case ast.KindTypeLiteral:
		return "object"
	case ast.KindTupleType:
		tupleType := node.AsTupleTypeNode()
		types := make([]string, 0)
		if tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				types = append(types, a.getTypeText(elem))
			}
		}
		return "[" + strings.Join(types, ", ") + "]"
	case ast.KindParenthesizedType:
		parenType := node.AsParenthesizedTypeNode()
		return "(" + a.getTypeText(parenType.Type) + ")"
	}

	return "any"
}

func (a *Analyzer) getEntityName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindQualifiedName:
		qn := node.AsQualifiedName()
		return a.getEntityName(qn.Left) + "." + qn.Right.Text
	}

	return ""
}

func (a *Analyzer) hasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}

	for _, mod := range modifiers.Nodes {
		if mod.Kind == kind {
			return true
		}
	}

	return false
}

func (a *Analyzer) getPrivacy(node *ast.Node) string {
	if a.hasModifier(node, ast.KindPrivateKeyword) {
		return "private"
	}
	if a.hasModifier(node, ast.KindProtectedKeyword) {
		return "protected"
	}

	// Check for private identifier (#)
	name := node.Name()
	if name != nil && name.Kind == ast.KindPrivateIdentifier {
		return "private"
	}

	return "public"
}

func (a *Analyzer) isExported(node *ast.Node) bool {
	return a.hasModifier(node, ast.KindExportKeyword)
}

func (a *Analyzer) isParameterProperty(node *ast.Node) bool {
	return a.hasModifier(node, ast.KindPublicKeyword) ||
		a.hasModifier(node, ast.KindPrivateKeyword) ||
		a.hasModifier(node, ast.KindProtectedKeyword) ||
		a.hasModifier(node, ast.KindReadonlyKeyword)
}

func (a *Analyzer) extendsHTMLElement(classDecl *ast.ClassDeclaration) bool {
	if classDecl.HeritageClauses == nil {
		return false
	}

	for _, clause := range classDecl.HeritageClauses.Nodes {
		hc := clause.AsHeritageClause()
		if hc != nil && hc.Token == ast.KindExtendsKeyword && hc.Types != nil {
			for _, typeExpr := range hc.Types.Nodes {
				name := a.getExpressionText(typeExpr)
				if strings.HasSuffix(name, "Element") || name == "HTMLElement" || name == "LitElement" {
					return true
				}
			}
		}
	}

	return false
}

// toKebabCase converts PascalCase or camelCase to kebab-case
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r + 32) // Convert to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
