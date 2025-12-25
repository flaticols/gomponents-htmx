# CEM Generator

Custom Elements Manifest generator using TypeScript compiler API with full class inheritance support.

## Features

- **AST-based parsing** - Uses TypeScript's compiler API, not regex
- **Class inheritance resolution** - Tracks inherited members with `inheritedFrom` references
- **JSDoc support** - Parses `@attr`, `@fires`, `@slot`, `@cssProperty`, `@cssPart`, `@customElement`, `@deprecated`
- **Decorator support** - Recognizes `@customElement()` and `@property()` decorators
- **CEM v2.1.0 compliant** - Outputs valid Custom Elements Manifest

## Installation

```bash
# Bun
bun add @anthropic/cem-generator

# npm
npm install @anthropic/cem-generator

# JSR
deno add @anthropic/cem-generator
```

## CLI Usage

```bash
# Generate manifest for TypeScript files
bun run cem-generator src/**/*.ts

# Output to custom file with pretty printing
bun run cem-generator -o manifest.json -p src/**/*.ts

# Output to stdout
bun run cem-generator -o - src/**/*.ts
```

### Options

- `-o, --output <file>` - Output file (default: `custom-elements.json`)
- `-p, --pretty` - Pretty print JSON output
- `-h, --help` - Show help

## Programmatic Usage

```typescript
import { Analyzer, type Manifest } from "@anthropic/cem-generator";

const files = ["src/my-element.ts"];
const baseDir = process.cwd();

const analyzer = new Analyzer(files, baseDir);
const manifest: Manifest = analyzer.analyze();

console.log(JSON.stringify(manifest, null, 2));
```

## Supported JSDoc Tags

| Tag | Description |
|-----|-------------|
| `@customElement <tag-name>` | Defines custom element tag name |
| `@attr [name]` | Marks property as an attribute |
| `@fires <event-name>` | Documents fired events |
| `@slot [name]` | Documents slots |
| `@cssProperty <name>` | Documents CSS custom properties |
| `@cssPart <name>` | Documents CSS parts |
| `@deprecated [message]` | Marks as deprecated |

## Example

```typescript
/**
 * A button component
 * @customElement my-button
 * @fires click - Fired on click
 * @slot icon - Optional icon slot
 * @cssProperty --button-bg - Background color
 */
export class MyButton extends HTMLElement {
  /**
   * Button variant
   * @attr
   */
  variant: "primary" | "secondary" = "primary";
}
```

## License

MIT
