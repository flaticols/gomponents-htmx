#!/usr/bin/env bun
/**
 * Custom Elements Manifest Generator
 *
 * Uses TypeScript's compiler API to parse source files and generate
 * a custom-elements.json manifest following the CEM specification.
 *
 * Usage:
 *   bun run index.ts [options] <files...>
 *
 * Options:
 *   -o, --output <file>   Output file (default: custom-elements.json)
 *   -p, --pretty          Pretty print JSON
 *   -h, --help            Show help
 */

import ts from "typescript";
import { parseArgs } from "util";
import { writeFileSync } from "fs";
import { resolve, relative } from "path";

// ============================================================================
// Manifest Types (Custom Elements Manifest v2.1.0)
// ============================================================================

interface Manifest {
  schemaVersion: string;
  readme?: string;
  modules: Module[];
}

interface Module {
  kind: "javascript-module";
  path: string;
  summary?: string;
  description?: string;
  declarations?: Declaration[];
  exports?: Export[];
  deprecated?: string | boolean;
}

type Declaration =
  | ClassDeclaration
  | CustomElementDeclaration
  | FunctionDeclaration
  | VariableDeclaration
  | MixinDeclaration;

interface ClassDeclaration {
  kind: "class";
  name: string;
  summary?: string;
  description?: string;
  superclass?: Reference;
  mixins?: Reference[];
  members?: ClassMember[];
  source?: SourceReference;
  deprecated?: string | boolean;
  typeParameters?: TypeParameter[];
}

interface CustomElementDeclaration extends ClassDeclaration {
  tagName?: string;
  customElement: true;
  attributes?: Attribute[];
  cssProperties?: CSSCustomProperty[];
  cssParts?: CSSPart[];
  slots?: Slot[];
  events?: Event[];
}

interface FunctionDeclaration {
  kind: "function";
  name: string;
  summary?: string;
  description?: string;
  parameters?: Parameter[];
  return?: ReturnType;
  typeParameters?: TypeParameter[];
  source?: SourceReference;
  deprecated?: string | boolean;
}

interface VariableDeclaration {
  kind: "variable";
  name: string;
  summary?: string;
  description?: string;
  type?: Type;
  default?: string;
  source?: SourceReference;
  deprecated?: string | boolean;
  readonly?: boolean;
}

interface MixinDeclaration {
  kind: "mixin";
  name: string;
  summary?: string;
  description?: string;
  mixins?: Reference[];
  members?: ClassMember[];
  parameters?: Parameter[];
  typeParameters?: TypeParameter[];
  source?: SourceReference;
  deprecated?: string | boolean;
}

type ClassMember = ClassField | ClassMethod;

interface ClassField {
  kind: "field";
  name: string;
  summary?: string;
  description?: string;
  privacy?: "public" | "protected" | "private";
  type?: Type;
  default?: string;
  static?: boolean;
  readonly?: boolean;
  source?: SourceReference;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
  attribute?: string;
  reflects?: boolean;
}

interface ClassMethod {
  kind: "method";
  name: string;
  summary?: string;
  description?: string;
  privacy?: "public" | "protected" | "private";
  parameters?: Parameter[];
  return?: ReturnType;
  typeParameters?: TypeParameter[];
  static?: boolean;
  source?: SourceReference;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface Reference {
  name: string;
  package?: string;
  module?: string;
}

interface Type {
  text: string;
  references?: TypeReference[];
}

interface TypeReference {
  name?: string;
  package?: string;
  module?: string;
  start?: number;
  end?: number;
}

interface TypeParameter {
  name: string;
  default?: string;
  extends?: string;
  description?: string;
}

interface Parameter {
  name: string;
  summary?: string;
  description?: string;
  type?: Type;
  default?: string;
  optional?: boolean;
  rest?: boolean;
}

interface ReturnType {
  type?: Type;
  summary?: string;
  description?: string;
}

interface Attribute {
  name: string;
  summary?: string;
  description?: string;
  type?: Type;
  default?: string;
  fieldName?: string;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface CSSCustomProperty {
  name: string;
  summary?: string;
  description?: string;
  default?: string;
  syntax?: string;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface CSSPart {
  name: string;
  summary?: string;
  description?: string;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface Slot {
  name: string;
  summary?: string;
  description?: string;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface Event {
  name: string;
  summary?: string;
  description?: string;
  type?: Type;
  inheritedFrom?: Reference;
  deprecated?: string | boolean;
}

interface Export {
  kind: "js" | "custom-element-definition";
  name: string;
  declaration?: Reference;
  deprecated?: string | boolean;
}

interface SourceReference {
  href?: string;
}

// ============================================================================
// Analyzer
// ============================================================================

interface ClassInfo {
  name: string;
  superClass?: string;
  modulePath: string;
  description?: string;
  tagName?: string;
  isElement: boolean;
  members: ClassMember[];
  attributes: Attribute[];
  events: Event[];
  slots: Slot[];
  cssProperties: CSSCustomProperty[];
  cssParts: CSSPart[];
  typeParameters?: TypeParameter[];
}

class Analyzer {
  private program: ts.Program;
  private checker: ts.TypeChecker;
  private classes = new Map<string, ClassInfo>();
  private modules = new Map<string, Module>();
  private baseDir: string;

  constructor(files: string[], baseDir: string) {
    this.baseDir = baseDir;
    this.program = ts.createProgram(files, {
      target: ts.ScriptTarget.ESNext,
      module: ts.ModuleKind.ESNext,
      allowJs: true,
      checkJs: true,
    });
    this.checker = this.program.getTypeChecker();
  }

  analyze(): Manifest {
    // First pass: collect all classes
    for (const sourceFile of this.program.getSourceFiles()) {
      if (sourceFile.isDeclarationFile) continue;
      this.analyzeSourceFile(sourceFile);
    }

    // Second pass: resolve inheritance
    this.resolveInheritance();

    // Build manifest
    return {
      schemaVersion: "2.1.0",
      modules: Array.from(this.modules.values()),
    };
  }

  private analyzeSourceFile(sourceFile: ts.SourceFile): void {
    const relativePath = relative(this.baseDir, sourceFile.fileName);
    const module: Module = {
      kind: "javascript-module",
      path: relativePath,
      declarations: [],
      exports: [],
    };
    this.modules.set(relativePath, module);

    ts.forEachChild(sourceFile, (node) => {
      this.visitNode(node, module, sourceFile);
    });
  }

  private visitNode(
    node: ts.Node,
    module: Module,
    sourceFile: ts.SourceFile
  ): void {
    if (ts.isClassDeclaration(node)) {
      this.visitClassDeclaration(node, module, sourceFile);
    } else if (ts.isFunctionDeclaration(node)) {
      this.visitFunctionDeclaration(node, module, sourceFile);
    } else if (ts.isVariableStatement(node)) {
      this.visitVariableStatement(node, module, sourceFile);
    } else if (ts.isExportDeclaration(node)) {
      this.visitExportDeclaration(node, module);
    } else if (ts.isExportAssignment(node)) {
      this.visitExportAssignment(node, module);
    }
  }

  private visitClassDeclaration(
    node: ts.ClassDeclaration,
    module: Module,
    sourceFile: ts.SourceFile
  ): void {
    const name = node.name?.text;
    if (!name) return;

    const classInfo: ClassInfo = {
      name,
      modulePath: module.path,
      isElement: false,
      members: [],
      attributes: [],
      events: [],
      slots: [],
      cssProperties: [],
      cssParts: [],
    };

    // Get JSDoc
    const jsDoc = this.getJSDoc(node, sourceFile);
    classInfo.description = jsDoc.description;

    // Extract tag name from @customElement JSDoc or decorator
    classInfo.tagName = this.extractTagName(node, jsDoc);

    // Check if extends HTMLElement
    if (node.heritageClauses) {
      for (const clause of node.heritageClauses) {
        if (clause.token === ts.SyntaxKind.ExtendsKeyword) {
          for (const type of clause.types) {
            const typeName = type.expression.getText(sourceFile);
            classInfo.superClass = typeName;
            if (
              typeName.endsWith("Element") ||
              typeName === "HTMLElement" ||
              typeName === "LitElement"
            ) {
              classInfo.isElement = true;
            }
          }
        }
      }
    }

    classInfo.isElement = classInfo.isElement || !!classInfo.tagName;

    // Extract JSDoc tags
    this.extractJSDocTags(jsDoc, classInfo);

    // Extract type parameters
    if (node.typeParameters) {
      classInfo.typeParameters = node.typeParameters.map((tp) => ({
        name: tp.name.text,
        extends: tp.constraint ? tp.constraint.getText(sourceFile) : undefined,
        default: tp.default ? tp.default.getText(sourceFile) : undefined,
      }));
    }

    // Visit members
    for (const member of node.members) {
      this.visitClassMember(member, classInfo, sourceFile);
    }

    this.classes.set(name, classInfo);

    // Create declaration
    const decl = this.createDeclaration(classInfo);
    module.declarations!.push(decl);

    // Add exports
    if (this.isExported(node)) {
      if (classInfo.isElement && classInfo.tagName) {
        module.exports!.push({
          kind: "custom-element-definition",
          name: classInfo.tagName,
          declaration: { name, module: module.path },
        });
      }
      module.exports!.push({
        kind: "js",
        name,
        declaration: { name, module: module.path },
      });
    }
  }

  private visitClassMember(
    node: ts.ClassElement,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    if (ts.isPropertyDeclaration(node)) {
      this.visitPropertyDeclaration(node, classInfo, sourceFile);
    } else if (ts.isMethodDeclaration(node)) {
      this.visitMethodDeclaration(node, classInfo, sourceFile);
    } else if (ts.isGetAccessor(node)) {
      this.visitGetAccessor(node, classInfo, sourceFile);
    } else if (ts.isSetAccessor(node)) {
      this.visitSetAccessor(node, classInfo, sourceFile);
    } else if (ts.isConstructorDeclaration(node)) {
      this.visitConstructor(node, classInfo, sourceFile);
    }
  }

  private visitPropertyDeclaration(
    node: ts.PropertyDeclaration,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    const name = this.getPropertyName(node.name, sourceFile);
    if (!name) return;

    const jsDoc = this.getJSDoc(node, sourceFile);

    const field: ClassField = {
      kind: "field",
      name,
      description: jsDoc.description,
      privacy: this.getPrivacy(node),
      static: this.hasModifier(node, ts.SyntaxKind.StaticKeyword),
      readonly: this.hasModifier(node, ts.SyntaxKind.ReadonlyKeyword),
    };

    if (jsDoc.deprecated) {
      field.deprecated = jsDoc.deprecated;
    }

    if (node.type) {
      field.type = { text: node.type.getText(sourceFile) };
    }

    if (node.initializer) {
      field.default = node.initializer.getText(sourceFile);
    }

    // Check for @attr tag
    const attrName = this.extractAttributeName(node, jsDoc, name, sourceFile);
    if (attrName) {
      field.attribute = attrName;
      field.reflects = true;

      classInfo.attributes.push({
        name: attrName,
        description: field.description,
        type: field.type,
        default: field.default,
        fieldName: name,
      });
    }

    classInfo.members.push(field);
  }

  private visitMethodDeclaration(
    node: ts.MethodDeclaration,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    const name = this.getPropertyName(node.name, sourceFile);
    if (!name) return;

    const jsDoc = this.getJSDoc(node, sourceFile);

    const method: ClassMethod = {
      kind: "method",
      name,
      description: jsDoc.description,
      privacy: this.getPrivacy(node),
      static: this.hasModifier(node, ts.SyntaxKind.StaticKeyword),
      parameters: this.extractParameters(node.parameters, jsDoc, sourceFile),
    };

    if (jsDoc.deprecated) {
      method.deprecated = jsDoc.deprecated;
    }

    if (node.type) {
      method.return = {
        type: { text: node.type.getText(sourceFile) },
        description: jsDoc.returns,
      };
    } else if (jsDoc.returns) {
      method.return = { description: jsDoc.returns };
    }

    if (node.typeParameters) {
      method.typeParameters = node.typeParameters.map((tp) => ({
        name: tp.name.text,
        extends: tp.constraint ? tp.constraint.getText(sourceFile) : undefined,
        default: tp.default ? tp.default.getText(sourceFile) : undefined,
      }));
    }

    classInfo.members.push(method);
  }

  private visitGetAccessor(
    node: ts.GetAccessorDeclaration,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    const name = this.getPropertyName(node.name, sourceFile);
    if (!name) return;

    // Check if we already have this field from a setter
    const existing = classInfo.members.find(
      (m) => m.kind === "field" && m.name === name
    );
    if (existing && existing.kind === "field") {
      if (node.type && !existing.type) {
        existing.type = { text: node.type.getText(sourceFile) };
      }
      return;
    }

    const jsDoc = this.getJSDoc(node, sourceFile);

    const field: ClassField = {
      kind: "field",
      name,
      description: jsDoc.description,
      privacy: this.getPrivacy(node),
      static: this.hasModifier(node, ts.SyntaxKind.StaticKeyword),
      readonly: true,
    };

    if (node.type) {
      field.type = { text: node.type.getText(sourceFile) };
    }

    const attrName = this.extractAttributeName(node, jsDoc, name, sourceFile);
    if (attrName) {
      field.attribute = attrName;
      field.reflects = true;
      classInfo.attributes.push({
        name: attrName,
        description: field.description,
        type: field.type,
        fieldName: name,
      });
    }

    classInfo.members.push(field);
  }

  private visitSetAccessor(
    node: ts.SetAccessorDeclaration,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    const name = this.getPropertyName(node.name, sourceFile);
    if (!name) return;

    // Check if we already have this field from a getter
    const existing = classInfo.members.find(
      (m) => m.kind === "field" && m.name === name
    );
    if (existing && existing.kind === "field") {
      existing.readonly = false;
      return;
    }

    const jsDoc = this.getJSDoc(node, sourceFile);

    const field: ClassField = {
      kind: "field",
      name,
      description: jsDoc.description,
      privacy: this.getPrivacy(node),
      static: this.hasModifier(node, ts.SyntaxKind.StaticKeyword),
    };

    if (node.parameters.length > 0) {
      const param = node.parameters[0];
      if (param.type) {
        field.type = { text: param.type.getText(sourceFile) };
      }
    }

    classInfo.members.push(field);
  }

  private visitConstructor(
    node: ts.ConstructorDeclaration,
    classInfo: ClassInfo,
    sourceFile: ts.SourceFile
  ): void {
    for (const param of node.parameters) {
      if (!this.isParameterProperty(param)) continue;

      const name = ts.isIdentifier(param.name) ? param.name.text : undefined;
      if (!name) continue;

      const jsDoc = this.getJSDoc(param, sourceFile);

      const field: ClassField = {
        kind: "field",
        name,
        description: jsDoc.description,
        privacy: this.getPrivacy(param),
        readonly: this.hasModifier(param, ts.SyntaxKind.ReadonlyKeyword),
      };

      if (param.type) {
        field.type = { text: param.type.getText(sourceFile) };
      }

      if (param.initializer) {
        field.default = param.initializer.getText(sourceFile);
      }

      classInfo.members.push(field);
    }
  }

  private visitFunctionDeclaration(
    node: ts.FunctionDeclaration,
    module: Module,
    sourceFile: ts.SourceFile
  ): void {
    const name = node.name?.text;
    if (!name) return;

    const jsDoc = this.getJSDoc(node, sourceFile);

    const decl: FunctionDeclaration = {
      kind: "function",
      name,
      description: jsDoc.description,
      parameters: this.extractParameters(node.parameters, jsDoc, sourceFile),
    };

    if (jsDoc.deprecated) {
      decl.deprecated = jsDoc.deprecated;
    }

    if (node.type) {
      decl.return = {
        type: { text: node.type.getText(sourceFile) },
        description: jsDoc.returns,
      };
    }

    if (node.typeParameters) {
      decl.typeParameters = node.typeParameters.map((tp) => ({
        name: tp.name.text,
        extends: tp.constraint ? tp.constraint.getText(sourceFile) : undefined,
        default: tp.default ? tp.default.getText(sourceFile) : undefined,
      }));
    }

    module.declarations!.push(decl);

    if (this.isExported(node)) {
      module.exports!.push({
        kind: "js",
        name,
        declaration: { name, module: module.path },
      });
    }
  }

  private visitVariableStatement(
    node: ts.VariableStatement,
    module: Module,
    sourceFile: ts.SourceFile
  ): void {
    const isConst = (node.declarationList.flags & ts.NodeFlags.Const) !== 0;

    for (const decl of node.declarationList.declarations) {
      if (!ts.isIdentifier(decl.name)) continue;
      const name = decl.name.text;

      const jsDoc = this.getJSDoc(node, sourceFile);

      const varDecl: VariableDeclaration = {
        kind: "variable",
        name,
        description: jsDoc.description,
        readonly: isConst,
      };

      if (jsDoc.deprecated) {
        varDecl.deprecated = jsDoc.deprecated;
      }

      if (decl.type) {
        varDecl.type = { text: decl.type.getText(sourceFile) };
      }

      if (decl.initializer) {
        varDecl.default = decl.initializer.getText(sourceFile);
      }

      module.declarations!.push(varDecl);

      if (this.isExported(node)) {
        module.exports!.push({
          kind: "js",
          name,
          declaration: { name, module: module.path },
        });
      }
    }
  }

  private visitExportDeclaration(node: ts.ExportDeclaration, module: Module): void {
    if (!node.exportClause || !ts.isNamedExports(node.exportClause)) return;

    for (const element of node.exportClause.elements) {
      const exportName = element.name.text;
      const localName = element.propertyName?.text ?? exportName;

      module.exports!.push({
        kind: "js",
        name: exportName,
        declaration: { name: localName, module: module.path },
      });
    }
  }

  private visitExportAssignment(node: ts.ExportAssignment, module: Module): void {
    const name = ts.isIdentifier(node.expression)
      ? node.expression.text
      : "default";

    module.exports!.push({
      kind: "js",
      name: "default",
      declaration: { name, module: module.path },
    });
  }

  // ============================================================================
  // Inheritance Resolution
  // ============================================================================

  private resolveInheritance(): void {
    for (const classInfo of this.classes.values()) {
      this.resolveClassInheritance(classInfo, new Set());
    }

    // Update declarations
    for (const module of this.modules.values()) {
      for (let i = 0; i < module.declarations!.length; i++) {
        const decl = module.declarations![i];
        if (decl.kind === "class") {
          const classInfo = this.classes.get(decl.name);
          if (classInfo) {
            module.declarations![i] = this.createDeclaration(classInfo);
          }
        }
      }
    }
  }

  private resolveClassInheritance(
    classInfo: ClassInfo,
    visited: Set<string>
  ): void {
    if (visited.has(classInfo.name)) return;
    visited.add(classInfo.name);

    if (!classInfo.superClass) return;

    const superClass = this.classes.get(classInfo.superClass);
    if (!superClass) return;

    // Resolve superclass first
    this.resolveClassInheritance(superClass, visited);

    // Copy inherited members
    const existingMembers = new Set(classInfo.members.map((m) => m.name));

    for (const member of superClass.members) {
      if (existingMembers.has(member.name)) continue;

      const inheritedMember = {
        ...member,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      };
      classInfo.members.push(inheritedMember);
    }

    // Copy inherited attributes
    const existingAttrs = new Set(classInfo.attributes.map((a) => a.name));
    for (const attr of superClass.attributes) {
      if (existingAttrs.has(attr.name)) continue;
      classInfo.attributes.push({
        ...attr,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      });
    }

    // Copy inherited events
    const existingEvents = new Set(classInfo.events.map((e) => e.name));
    for (const event of superClass.events) {
      if (existingEvents.has(event.name)) continue;
      classInfo.events.push({
        ...event,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      });
    }

    // Copy inherited slots
    const existingSlots = new Set(classInfo.slots.map((s) => s.name));
    for (const slot of superClass.slots) {
      if (existingSlots.has(slot.name)) continue;
      classInfo.slots.push({
        ...slot,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      });
    }

    // Copy inherited CSS properties
    const existingCssProps = new Set(classInfo.cssProperties.map((c) => c.name));
    for (const prop of superClass.cssProperties) {
      if (existingCssProps.has(prop.name)) continue;
      classInfo.cssProperties.push({
        ...prop,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      });
    }

    // Copy inherited CSS parts
    const existingCssParts = new Set(classInfo.cssParts.map((c) => c.name));
    for (const part of superClass.cssParts) {
      if (existingCssParts.has(part.name)) continue;
      classInfo.cssParts.push({
        ...part,
        inheritedFrom: {
          name: superClass.name,
          module: superClass.modulePath,
        },
      });
    }
  }

  private createDeclaration(classInfo: ClassInfo): Declaration {
    if (classInfo.isElement) {
      const decl: CustomElementDeclaration = {
        kind: "class",
        name: classInfo.name,
        customElement: true,
        description: classInfo.description,
        members: classInfo.members.length > 0 ? classInfo.members : undefined,
        typeParameters: classInfo.typeParameters,
      };

      if (classInfo.tagName) decl.tagName = classInfo.tagName;
      if (classInfo.superClass) {
        decl.superclass = { name: classInfo.superClass };
      }
      if (classInfo.attributes.length > 0) {
        decl.attributes = classInfo.attributes;
      }
      if (classInfo.events.length > 0) {
        decl.events = classInfo.events;
      }
      if (classInfo.slots.length > 0) {
        decl.slots = classInfo.slots;
      }
      if (classInfo.cssProperties.length > 0) {
        decl.cssProperties = classInfo.cssProperties;
      }
      if (classInfo.cssParts.length > 0) {
        decl.cssParts = classInfo.cssParts;
      }

      return decl;
    }

    const decl: ClassDeclaration = {
      kind: "class",
      name: classInfo.name,
      description: classInfo.description,
      members: classInfo.members.length > 0 ? classInfo.members : undefined,
      typeParameters: classInfo.typeParameters,
    };

    if (classInfo.superClass) {
      decl.superclass = { name: classInfo.superClass };
    }

    return decl;
  }

  // ============================================================================
  // JSDoc Parsing
  // ============================================================================

  private getJSDoc(
    node: ts.Node,
    sourceFile: ts.SourceFile
  ): {
    description?: string;
    deprecated?: string | boolean;
    returns?: string;
    params: Map<string, string>;
    tags: Map<string, string[]>;
  } {
    const result = {
      description: undefined as string | undefined,
      deprecated: undefined as string | boolean | undefined,
      returns: undefined as string | undefined,
      params: new Map<string, string>(),
      tags: new Map<string, string[]>(),
    };

    const jsDocs = ts.getJSDocCommentsAndTags(node);

    for (const jsDoc of jsDocs) {
      if (ts.isJSDoc(jsDoc)) {
        if (jsDoc.comment) {
          result.description = ts.getTextOfJSDocComment(jsDoc.comment);
        }

        if (jsDoc.tags) {
          for (const tag of jsDoc.tags) {
            const tagName = tag.tagName.text;
            const comment = tag.comment
              ? ts.getTextOfJSDocComment(tag.comment)
              : "";

            if (tagName === "deprecated") {
              result.deprecated = comment || true;
            } else if (tagName === "returns" || tagName === "return") {
              result.returns = comment;
            } else if (tagName === "param" && ts.isJSDocParameterTag(tag)) {
              const paramName = tag.name.getText(sourceFile);
              result.params.set(paramName, comment ?? "");
            } else {
              const existing = result.tags.get(tagName) ?? [];
              existing.push(comment ?? "");
              result.tags.set(tagName, existing);
            }
          }
        }
      }
    }

    return result;
  }

  private extractTagName(
    node: ts.ClassDeclaration,
    jsDoc: ReturnType<typeof this.getJSDoc>
  ): string | undefined {
    // Check @customElement JSDoc tag
    const customElementTags = jsDoc.tags.get("customElement");
    if (customElementTags && customElementTags[0]) {
      return customElementTags[0].trim();
    }

    // Check decorators
    const decorators = ts.getDecorators(node);
    if (decorators) {
      for (const decorator of decorators) {
        if (ts.isCallExpression(decorator.expression)) {
          const expr = decorator.expression.expression;
          if (ts.isIdentifier(expr)) {
            const name = expr.text;
            if (name === "customElement" || name === "define") {
              const arg = decorator.expression.arguments[0];
              if (arg && ts.isStringLiteral(arg)) {
                return arg.text;
              }
            }
          }
        }
      }
    }

    return undefined;
  }

  private extractAttributeName(
    node: ts.Node,
    jsDoc: ReturnType<typeof this.getJSDoc>,
    propName: string,
    sourceFile: ts.SourceFile
  ): string | undefined {
    // Check @attr JSDoc tag
    const attrTags = jsDoc.tags.get("attr") ?? jsDoc.tags.get("attribute");
    if (attrTags) {
      const value = attrTags[0]?.trim();
      return value || this.toKebabCase(propName);
    }

    // Check @property decorator
    if (ts.isPropertyDeclaration(node) || ts.isGetAccessor(node)) {
      const decorators = ts.getDecorators(node);
      if (decorators) {
        for (const decorator of decorators) {
          if (ts.isCallExpression(decorator.expression)) {
            const expr = decorator.expression.expression;
            if (ts.isIdentifier(expr) && expr.text === "property") {
              // Check for attribute: false
              const arg = decorator.expression.arguments[0];
              if (arg && ts.isObjectLiteralExpression(arg)) {
                for (const prop of arg.properties) {
                  if (
                    ts.isPropertyAssignment(prop) &&
                    ts.isIdentifier(prop.name) &&
                    prop.name.text === "attribute" &&
                    prop.initializer.kind === ts.SyntaxKind.FalseKeyword
                  ) {
                    return undefined;
                  }
                }
              }
              return this.toKebabCase(propName);
            }
          }
        }
      }
    }

    return undefined;
  }

  private extractJSDocTags(
    jsDoc: ReturnType<typeof this.getJSDoc>,
    classInfo: ClassInfo
  ): void {
    // @fires / @event
    const fires = jsDoc.tags.get("fires") ?? jsDoc.tags.get("event") ?? [];
    for (const fire of fires) {
      const parts = fire.split(/\s+/);
      const eventName = parts[0];
      const description = parts.slice(1).join(" ").replace(/^-\s*/, "");
      classInfo.events.push({ name: eventName, description });
    }

    // @slot
    const slots = jsDoc.tags.get("slot") ?? [];
    for (const slot of slots) {
      const parts = slot.split(/\s+/);
      let slotName = "";
      let description = slot;
      if (parts[0] !== "-") {
        slotName = parts[0];
        description = parts.slice(1).join(" ").replace(/^-\s*/, "");
      }
      classInfo.slots.push({ name: slotName, description });
    }

    // @cssProperty
    const cssProps =
      jsDoc.tags.get("cssProperty") ?? jsDoc.tags.get("cssproperty") ?? [];
    for (const prop of cssProps) {
      const parts = prop.split(/\s+/);
      const propName = parts[0];
      const description = parts.slice(1).join(" ").replace(/^-\s*/, "");
      classInfo.cssProperties.push({ name: propName, description });
    }

    // @cssPart
    const cssParts =
      jsDoc.tags.get("cssPart") ?? jsDoc.tags.get("csspart") ?? [];
    for (const part of cssParts) {
      const parts = part.split(/\s+/);
      const partName = parts[0];
      const description = parts.slice(1).join(" ").replace(/^-\s*/, "");
      classInfo.cssParts.push({ name: partName, description });
    }
  }

  private extractParameters(
    params: ts.NodeArray<ts.ParameterDeclaration>,
    jsDoc: ReturnType<typeof this.getJSDoc>,
    sourceFile: ts.SourceFile
  ): Parameter[] {
    return params.map((param) => {
      const name = ts.isIdentifier(param.name) ? param.name.text : "unknown";

      const p: Parameter = {
        name,
        description: jsDoc.params.get(name),
        optional: !!param.questionToken,
        rest: !!param.dotDotDotToken,
      };

      if (param.type) {
        p.type = { text: param.type.getText(sourceFile) };
      }

      if (param.initializer) {
        p.default = param.initializer.getText(sourceFile);
        p.optional = true;
      }

      return p;
    });
  }

  // ============================================================================
  // Helpers
  // ============================================================================

  private getPropertyName(
    node: ts.PropertyName,
    sourceFile: ts.SourceFile
  ): string | undefined {
    if (ts.isIdentifier(node)) return node.text;
    if (ts.isStringLiteral(node)) return node.text;
    if (ts.isNumericLiteral(node)) return node.text;
    if (ts.isPrivateIdentifier(node)) return node.text;
    return undefined;
  }

  private hasModifier(node: ts.Node, kind: ts.SyntaxKind): boolean {
    const modifiers = ts.getModifiers(node);
    return modifiers?.some((m) => m.kind === kind) ?? false;
  }

  private getPrivacy(
    node: ts.Node
  ): "public" | "protected" | "private" | undefined {
    if (this.hasModifier(node, ts.SyntaxKind.PrivateKeyword)) return "private";
    if (this.hasModifier(node, ts.SyntaxKind.ProtectedKeyword))
      return "protected";

    // Check for private identifier (#name)
    if (
      (ts.isPropertyDeclaration(node) ||
        ts.isMethodDeclaration(node) ||
        ts.isGetAccessor(node) ||
        ts.isSetAccessor(node)) &&
      ts.isPrivateIdentifier(node.name)
    ) {
      return "private";
    }

    return undefined;
  }

  private isExported(node: ts.Node): boolean {
    return this.hasModifier(node, ts.SyntaxKind.ExportKeyword);
  }

  private isParameterProperty(node: ts.ParameterDeclaration): boolean {
    return (
      this.hasModifier(node, ts.SyntaxKind.PublicKeyword) ||
      this.hasModifier(node, ts.SyntaxKind.PrivateKeyword) ||
      this.hasModifier(node, ts.SyntaxKind.ProtectedKeyword) ||
      this.hasModifier(node, ts.SyntaxKind.ReadonlyKeyword)
    );
  }

  private toKebabCase(str: string): string {
    return str.replace(/([a-z])([A-Z])/g, "$1-$2").toLowerCase();
  }
}

// ============================================================================
// CLI
// ============================================================================

async function main() {
  const { values, positionals } = parseArgs({
    args: Bun.argv.slice(2),
    options: {
      output: { type: "string", short: "o", default: "custom-elements.json" },
      pretty: { type: "boolean", short: "p", default: false },
      help: { type: "boolean", short: "h", default: false },
    },
    allowPositionals: true,
  });

  if (values.help || positionals.length === 0) {
    console.log(`
Custom Elements Manifest Generator

Usage:
  bun run index.ts [options] <files...>

Options:
  -o, --output <file>   Output file (default: custom-elements.json)
  -p, --pretty          Pretty print JSON
  -h, --help            Show help

Examples:
  bun run index.ts src/*.ts
  bun run index.ts -o manifest.json -p src/**/*.ts
`);
    process.exit(values.help ? 0 : 1);
  }

  // Expand globs using Bun's built-in Glob
  const files: string[] = [];
  for (const pattern of positionals) {
    const g = new Bun.Glob(pattern);
    for await (const file of g.scan({ cwd: process.cwd(), absolute: true })) {
      files.push(file);
    }
  }

  if (files.length === 0) {
    console.error("No files found");
    process.exit(1);
  }

  const baseDir = process.cwd();

  console.error(`Analyzing ${files.length} file(s)...`);
  for (const file of files) {
    console.error(`  ${relative(baseDir, file)}`);
  }

  const analyzer = new Analyzer(files, baseDir);
  const manifest = analyzer.analyze();

  const json = values.pretty
    ? JSON.stringify(manifest, null, 2)
    : JSON.stringify(manifest);

  if (values.output === "-") {
    console.log(json);
  } else {
    writeFileSync(values.output!, json);
    console.error(`Manifest written to: ${values.output}`);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
