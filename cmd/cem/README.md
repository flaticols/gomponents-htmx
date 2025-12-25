# Custom Elements Manifest Generator

A Go tool that generates Custom Elements Manifest (CEM) JSON files from TypeScript source code using Microsoft's [TypeScript-Go](https://github.com/microsoft/typescript-go) compiler.

## Features

- **Full AST Parsing**: Uses the TypeScript-Go compiler's AST parser (no regex)
- **Class Inheritance Resolution**: Properly resolves and includes inherited members
- **JSDoc Support**: Extracts descriptions, `@param`, `@returns`, `@deprecated` tags
- **Web Component Annotations**: Supports `@customElement`, `@attr`, `@fires`, `@slot`, `@cssProperty`, `@cssPart`
- **Decorator Support**: Recognizes `@customElement()`, `@property()` decorators

## Requirements

- Go 1.25+ (required by typescript-go)
- TypeScript-Go repository

## Installation

Since this tool uses internal packages from typescript-go, it must be built from within the typescript-go repository:

```bash
# Clone typescript-go
git clone https://github.com/microsoft/typescript-go.git
cd typescript-go

# Copy the cem tool
cp -r /path/to/this/cmd/cem cmd/cem

# Build
go build -o cem ./cmd/cem
```

## Usage

```bash
# Basic usage
./cem src/*.ts

# With options
./cem -o custom-elements.json -p src/**/*.ts

# Output to stdout
./cem -o - src/*.ts
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o`, `--output` | Output file path | `custom-elements.json` |
| `-d`, `--dir` | Base directory for relative paths | `.` |
| `-p`, `--pretty` | Pretty-print JSON output | `false` |
| `-h`, `--help` | Show help message | |

## Manifest Schema

The generated manifest follows the [Custom Elements Manifest](https://github.com/webcomponents/custom-elements-manifest) specification v2.1.0.

## JSDoc Tags

The analyzer recognizes the following JSDoc tags:

| Tag | Description |
|-----|-------------|
| `@customElement tag-name` | Defines the custom element tag name |
| `@attr [name]` | Marks a property as an attribute |
| `@fires event-name - description` | Documents a fired event |
| `@slot [name] - description` | Documents a slot |
| `@cssProperty --name - description` | Documents a CSS custom property |
| `@cssPart name - description` | Documents a CSS part |
| `@deprecated [message]` | Marks as deprecated |
| `@param name description` | Documents a parameter |
| `@returns description` | Documents return value |

## Example

Input TypeScript:

```typescript
/**
 * A button component
 * @customElement my-button
 * @fires click - When clicked
 * @slot icon - Icon slot
 * @cssProperty --button-bg - Background color
 */
export class MyButton extends HTMLElement {
  /** @attr */
  disabled: boolean = false;

  /**
   * Click the button
   * @returns The click event
   */
  click(): MouseEvent { ... }
}
```

Output manifest (partial):

```json
{
  "schemaVersion": "2.1.0",
  "modules": [{
    "kind": "javascript-module",
    "path": "button.ts",
    "declarations": [{
      "kind": "class",
      "name": "MyButton",
      "tagName": "my-button",
      "customElement": true,
      "attributes": [{"name": "disabled", "type": {"text": "boolean"}}],
      "events": [{"name": "click", "description": "When clicked"}],
      "slots": [{"name": "icon", "description": "Icon slot"}],
      "cssProperties": [{"name": "--button-bg", "description": "Background color"}]
    }]
  }]
}
```

## Class Inheritance

The tool properly resolves class inheritance:

```typescript
export class BaseElement extends HTMLElement {
  theme: string = 'light';
}

export class MyButton extends BaseElement {
  // Inherits 'theme' property, marked with inheritedFrom
}
```

The manifest will include inherited members with proper attribution.

## Architecture

- `manifest.go` - Data structures for Custom Elements Manifest
- `analyzer.go` - TypeScript AST analysis and extraction
- `main.go` - CLI entry point

## License

Apache-2.0 (same as typescript-go)
