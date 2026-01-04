````markdown
---
name: ai-code-quality
description: >
  Code formatting and linting standards for all AI services.
  Covers .NET, Go, Python, TypeScript, and React projects.
  Use after building code and before running unit tests to ensure consistent code quality.
  Enforces consistent code style across 100+ developers.
---

# AI Code Quality - Formatting & Linting

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ⚠️  CODE QUALITY STANDARDS  ⚠️                                     ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   ALL code MUST pass formatting and linting checks before proceeding to tests!           ║
║                                                                                          ║
║   This phase runs AFTER successful build (Phase 2) and BEFORE unit tests (Phase 3)      ║
║                                                                                          ║
║   ❌ DO NOT skip linting - it catches bugs before runtime!                               ║
║   ❌ DO NOT ignore warnings - fix them or document exceptions!                           ║
║   ❌ DO NOT proceed to tests until code quality passes!                                  ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```

---

## Phase Position in Development Workflow

```
Phase 0:  Architecture Analysis
Phase 1:  Generate Code & Seed Files
Phase 2:  Build Locally & Fix Errors
          ↓
┌─────────────────────────────────────────────────────────────────────────┐
│  ▶ Phase 2.5: CODE QUALITY (Format & Lint)  ◀                          │
│     • Run formatters to auto-fix style issues                          │
│     • Run linters to catch potential bugs                              │
│     • Fix any linting errors before proceeding                         │
└─────────────────────────────────────────────────────────────────────────┘
          ↓
Phase 3:  Run Unit Tests
Phase 4:  Build Docker Image
...
```

---

## Quick Reference

| Language | Formatter | Linter | Config Files |
|----------|-----------|--------|--------------|
| **.NET** | `dotnet format` | Built-in analyzers + StyleCop | `.editorconfig`, `Directory.Build.props` |
| **Go** | `gofmt`, `goimports` | `golangci-lint` | `.golangci.yml` |
| **Python** | `black`, `isort` | `ruff` or `flake8` + `mypy` | `pyproject.toml`, `.flake8` |
| **TypeScript/React** | `prettier` | `eslint` | `.prettierrc`, `eslint.config.js` |

---

## .NET Code Quality

### Required Configuration Files

#### .editorconfig (Project Root)

```ini
# .editorconfig
root = true

[*]
indent_style = space
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.{cs,csx}]
# Naming conventions
dotnet_naming_rule.public_members_must_be_capitalized.symbols = public_symbols
dotnet_naming_rule.public_members_must_be_capitalized.style = first_word_upper_case_style
dotnet_naming_rule.public_members_must_be_capitalized.severity = warning

dotnet_naming_symbols.public_symbols.applicable_kinds = property, method, field, event, delegate
dotnet_naming_symbols.public_symbols.applicable_accessibilities = public

dotnet_naming_style.first_word_upper_case_style.capitalization = first_word_upper

# Private fields with underscore prefix
dotnet_naming_rule.private_fields_with_underscore.symbols = private_fields
dotnet_naming_rule.private_fields_with_underscore.style = prefix_underscore
dotnet_naming_rule.private_fields_with_underscore.severity = warning

dotnet_naming_symbols.private_fields.applicable_kinds = field
dotnet_naming_symbols.private_fields.applicable_accessibilities = private

dotnet_naming_style.prefix_underscore.capitalization = camel_case
dotnet_naming_style.prefix_underscore.required_prefix = _

# Code style
csharp_style_var_for_built_in_types = true:suggestion
csharp_style_var_when_type_is_apparent = true:suggestion
csharp_style_expression_bodied_methods = when_on_single_line:suggestion
csharp_style_expression_bodied_properties = true:suggestion
csharp_prefer_braces = true:warning
csharp_using_directive_placement = outside_namespace:warning

# Formatting
csharp_new_line_before_open_brace = all
csharp_new_line_before_else = true
csharp_new_line_before_catch = true
csharp_new_line_before_finally = true

# Analyzer severities
dotnet_diagnostic.CA1062.severity = warning  # Validate arguments of public methods
dotnet_diagnostic.CA1303.severity = none     # Do not pass literals as localized parameters
dotnet_diagnostic.CA1816.severity = warning  # Dispose methods should call SuppressFinalize
dotnet_diagnostic.CA2007.severity = none     # ConfigureAwait
dotnet_diagnostic.IDE0058.severity = none    # Expression value is never used

[*.{json,yaml,yml}]
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

#### Directory.Build.props (Solution Root)

```xml
<Project>
  <PropertyGroup>
    <!-- Enable analyzers -->
    <EnableNETAnalyzers>true</EnableNETAnalyzers>
    <AnalysisLevel>latest</AnalysisLevel>
    <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
    
    <!-- Treat warnings as errors in CI -->
    <TreatWarningsAsErrors Condition="'$(CI)' == 'true'">true</TreatWarningsAsErrors>
    
    <!-- Enable nullable reference types -->
    <Nullable>enable</Nullable>
    
    <!-- Documentation -->
    <GenerateDocumentationFile>true</GenerateDocumentationFile>
    <NoWarn>$(NoWarn);1591</NoWarn> <!-- Missing XML comment -->
  </PropertyGroup>
  
  <ItemGroup>
    <!-- StyleCop Analyzers -->
    <PackageReference Include="StyleCop.Analyzers" Version="1.2.0-beta.556">
      <PrivateAssets>all</PrivateAssets>
      <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
    </PackageReference>
  </ItemGroup>
</Project>
```

### .NET Commands

```powershell
# Format code (auto-fix style issues)
dotnet format --verbosity normal

# Format with verification only (CI mode - no changes, just report)
dotnet format --verify-no-changes --verbosity diagnostic

# Build with all analyzers
dotnet build /p:EnforceCodeStyleInBuild=true

# Run analyzers without build
dotnet format analyzers --severity info
```

### .NET Taskfile Tasks

```yaml
# Add to Taskfile.yml
tasks:
  format:
    desc: Format .NET code
    cmds:
      - dotnet format --verbosity normal
    
  lint:
    desc: Run .NET analyzers
    cmds:
      - dotnet build /p:EnforceCodeStyleInBuild=true /p:TreatWarningsAsErrors=true

  quality:
    desc: Run all code quality checks
    cmds:
      - task: format
      - task: lint
```

---

## Go Code Quality

### Required Configuration Files

#### .golangci.yml (Project Root)

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    # Default linters
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Report suspicious constructs
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # Static analysis
    - unused        # Check for unused code
    
    # Additional recommended linters
    - bodyclose     # Check HTTP response body is closed
    - contextcheck  # Check context.Context is passed correctly
    - dupl          # Code duplication
    - errname       # Check error naming conventions
    - errorlint     # Error wrapping issues
    - exhaustive    # Check switch exhaustiveness
    - gocognit      # Cognitive complexity
    - goconst       # Find repeated strings that could be constants
    - gocritic      # Opinionated linter
    - gocyclo       # Cyclomatic complexity
    - godot         # Check comments end with period
    - gofmt         # Check code is gofmt-ed
    - goimports     # Check imports are sorted
    - gomnd         # Magic number detector
    - gosec         # Security problems
    - misspell      # Spelling mistakes
    - nakedret      # Naked returns in functions
    - nilerr        # Return nil after checking error
    - nilnil        # Return nil,nil
    - noctx         # HTTP requests without context
    - prealloc      # Find slice declarations that could be preallocated
    - predeclared   # Find shadowed predeclared identifiers
    - revive        # Fast, configurable linter
    - unconvert     # Unnecessary type conversions
    - unparam       # Unused function parameters
    - whitespace    # Whitespace issues

linters-settings:
  gocyclo:
    min-complexity: 15
  
  gocognit:
    min-complexity: 20
  
  goconst:
    min-len: 3
    min-occurrences: 3
  
  gomnd:
    checks:
      - argument
      - case
      - condition
      - return
    ignored-numbers:
      - '0'
      - '1'
      - '2'
      - '10'
      - '100'
    ignored-functions:
      - 'time.*'
      - 'context.*'
  
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
    disabled-checks:
      - ifElseChain
      - hugeParam
  
  revive:
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: exported
      - name: increment-decrement
      - name: var-naming
      - name: var-declaration
      - name: package-comments
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf
      - name: empty-block
      - name: superfluous-else
      - name: unreachable-code

issues:
  exclude-rules:
    # Exclude some linters from running on tests files
    - path: _test\.go
      linters:
        - dupl
        - gomnd
        - goconst
    
    # Exclude some linters from generated files
    - path: \.pb\.go
      linters:
        - all

  max-issues-per-linter: 50
  max-same-issues: 10
```

### Go Commands

```bash
# Format code
gofmt -w -s .
goimports -w .

# Run linter
golangci-lint run

# Run linter with auto-fix (where possible)
golangci-lint run --fix

# Run specific linters
golangci-lint run --enable=gosec,gocritic

# Generate lint report
golangci-lint run --out-format=json > lint-report.json
```

### Go Taskfile Tasks

```yaml
# Add to Taskfile.yml
tasks:
  fmt:
    desc: Format Go code
    cmds:
      - gofmt -w -s .
      - goimports -w .

  lint:
    desc: Run Go linters
    cmds:
      - golangci-lint run

  lint-fix:
    desc: Run Go linters with auto-fix
    cmds:
      - golangci-lint run --fix

  quality:
    desc: Run all code quality checks
    cmds:
      - task: fmt
      - task: lint
```

### Installing golangci-lint

```bash
# macOS
brew install golangci-lint

# Windows (via scoop)
scoop install golangci-lint

# Windows (via chocolatey)
choco install golangci-lint

# Linux/CI
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0

# Or via go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Python Code Quality

### Required Configuration Files

#### pyproject.toml

```toml
# pyproject.toml
[project]
name = "your-service"
version = "1.0.0"
requires-python = ">=3.11"

[tool.black]
line-length = 100
target-version = ['py311']
include = '\.pyi?$'
exclude = '''
/(
    \.eggs
  | \.git
  | \.hg
  | \.mypy_cache
  | \.tox
  | \.venv
  | _build
  | buck-out
  | build
  | dist
  | migrations
)/
'''

[tool.isort]
profile = "black"
line_length = 100
multi_line_output = 3
include_trailing_comma = true
force_grid_wrap = 0
use_parentheses = true
ensure_newline_before_comments = true
skip = [".venv", "venv", "migrations"]

[tool.ruff]
line-length = 100
target-version = "py311"
select = [
    "E",    # pycodestyle errors
    "W",    # pycodestyle warnings
    "F",    # pyflakes
    "I",    # isort
    "B",    # flake8-bugbear
    "C4",   # flake8-comprehensions
    "UP",   # pyupgrade
    "ARG",  # flake8-unused-arguments
    "SIM",  # flake8-simplify
    "TCH",  # flake8-type-checking
    "PTH",  # flake8-use-pathlib
    "ERA",  # eradicate (commented-out code)
    "PL",   # pylint
    "RUF",  # ruff-specific rules
]
ignore = [
    "E501",   # line too long (handled by black)
    "B008",   # do not perform function calls in argument defaults
    "PLR0913", # too many arguments
]
exclude = [
    ".git",
    ".venv",
    "venv",
    "__pycache__",
    "migrations",
    "*.egg-info",
]

[tool.ruff.per-file-ignores]
"__init__.py" = ["F401"]
"tests/*" = ["ARG", "PLR2004"]

[tool.mypy]
python_version = "3.11"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
disallow_incomplete_defs = true
check_untyped_defs = true
disallow_untyped_decorators = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true
warn_no_return = true
warn_unreachable = true
strict_equality = true

[[tool.mypy.overrides]]
module = "tests.*"
disallow_untyped_defs = false

[[tool.mypy.overrides]]
module = [
    "kafka.*",
    "redis.*",
]
ignore_missing_imports = true

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = "-v --cov=src --cov-report=term-missing"
```

### Python Commands

```bash
# Format with black
black .

# Sort imports with isort
isort .

# Lint with ruff (faster alternative to flake8)
ruff check .

# Lint with auto-fix
ruff check --fix .

# Type checking with mypy
mypy src/

# Run all checks
black . && isort . && ruff check . && mypy src/
```

### Python Taskfile Tasks

```yaml
# Add to Taskfile.yml
tasks:
  format:
    desc: Format Python code
    cmds:
      - black .
      - isort .

  lint:
    desc: Run Python linters
    cmds:
      - ruff check .
      - mypy src/

  lint-fix:
    desc: Run Python linters with auto-fix
    cmds:
      - ruff check --fix .

  quality:
    desc: Run all code quality checks
    cmds:
      - task: format
      - task: lint
```

### Installing Python Tools

```bash
# Using pip
pip install black isort ruff mypy

# Using poetry
poetry add --group dev black isort ruff mypy

# Using uv (faster)
uv pip install black isort ruff mypy
```

---

## TypeScript/React Code Quality

### Required Configuration Files

#### .prettierrc

```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100,
  "bracketSpacing": true,
  "arrowParens": "avoid",
  "endOfLine": "lf"
}
```

#### .prettierignore

```
node_modules
dist
build
coverage
*.min.js
*.min.css
package-lock.json
```

#### eslint.config.js (ESLint Flat Config - v9+)

```javascript
// eslint.config.js
import js from '@eslint/js';
import typescript from '@typescript-eslint/eslint-plugin';
import typescriptParser from '@typescript-eslint/parser';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import importPlugin from 'eslint-plugin-import';

export default [
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescriptParser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    plugins: {
      '@typescript-eslint': typescript,
      'react': react,
      'react-hooks': reactHooks,
      'jsx-a11y': jsxA11y,
      'import': importPlugin,
    },
    rules: {
      // TypeScript
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      
      // React
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      'react/jsx-uses-react': 'off',
      'react/jsx-key': 'error',
      'react/no-unescaped-entities': 'warn',
      
      // React Hooks
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      
      // Accessibility
      'jsx-a11y/alt-text': 'error',
      'jsx-a11y/anchor-is-valid': 'warn',
      'jsx-a11y/click-events-have-key-events': 'warn',
      'jsx-a11y/no-static-element-interactions': 'warn',
      
      // Imports
      'import/order': [
        'error',
        {
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling'],
            'index',
            'type',
          ],
          'newlines-between': 'always',
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
      'import/no-duplicates': 'error',
      
      // General
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },
  {
    ignores: ['node_modules', 'dist', 'build', 'coverage', '*.config.js'],
  },
];
```

#### Alternative: .eslintrc.cjs (Legacy Config)

```javascript
// .eslintrc.cjs
module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended',
    'plugin:jsx-a11y/recommended',
    'prettier', // Must be last to override other configs
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
  },
  plugins: ['@typescript-eslint', 'react', 'react-hooks', 'jsx-a11y', 'import'],
  rules: {
    // Same rules as above
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    'react/react-in-jsx-scope': 'off',
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'warn',
    'import/order': ['error', { 'newlines-between': 'always', alphabetize: { order: 'asc' } }],
    'no-console': ['warn', { allow: ['warn', 'error'] }],
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
  ignorePatterns: ['node_modules', 'dist', 'build', 'coverage'],
};
```

### TypeScript Commands

```bash
# Format with prettier
npx prettier --write "src/**/*.{ts,tsx,css,json}"

# Check formatting (CI mode)
npx prettier --check "src/**/*.{ts,tsx,css,json}"

# Lint with eslint
npx eslint src/

# Lint with auto-fix
npx eslint src/ --fix

# TypeScript type checking
npx tsc --noEmit

# Run all checks
npx prettier --write . && npx eslint src/ --fix && npx tsc --noEmit
```

### package.json Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "format": "prettier --write \"src/**/*.{ts,tsx,css,json}\"",
    "format:check": "prettier --check \"src/**/*.{ts,tsx,css,json}\"",
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix",
    "typecheck": "tsc --noEmit",
    "quality": "npm run format && npm run lint:fix && npm run typecheck"
  },
  "devDependencies": {
    "@typescript-eslint/eslint-plugin": "^7.0.0",
    "@typescript-eslint/parser": "^7.0.0",
    "eslint": "^9.0.0",
    "eslint-config-prettier": "^9.1.0",
    "eslint-plugin-import": "^2.29.0",
    "eslint-plugin-jsx-a11y": "^6.8.0",
    "eslint-plugin-react": "^7.33.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "prettier": "^3.2.0",
    "typescript": "^5.3.0"
  }
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/code-quality.yml
name: Code Quality

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  dotnet-quality:
    runs-on: ubuntu-latest
    if: hashFiles('**/*.csproj') != ''
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-dotnet@v4
        with:
          dotnet-version: '8.0.x'
      - name: Format Check
        run: dotnet format --verify-no-changes
      - name: Build with Analyzers
        run: dotnet build /p:TreatWarningsAsErrors=true

  go-quality:
    runs-on: ubuntu-latest
    if: hashFiles('go.mod') != ''
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Format Check
        run: test -z "$(gofmt -l .)"
      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  python-quality:
    runs-on: ubuntu-latest
    if: hashFiles('pyproject.toml') != ''
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Install tools
        run: pip install black isort ruff mypy
      - name: Format Check
        run: |
          black --check .
          isort --check .
      - name: Lint
        run: |
          ruff check .
          mypy src/

  typescript-quality:
    runs-on: ubuntu-latest
    if: hashFiles('package.json') != ''
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - name: Install
        run: npm ci
      - name: Format Check
        run: npx prettier --check "src/**/*.{ts,tsx,css,json}"
      - name: Lint
        run: npx eslint src/
      - name: Type Check
        run: npx tsc --noEmit
```

---

## Error Resolution Guide

### Common .NET Issues

| Error | Cause | Fix |
|-------|-------|-----|
| `IDE0044: Make field readonly` | Field never reassigned | Add `readonly` modifier |
| `CS8618: Non-nullable field not initialized` | Nullable warning | Initialize in constructor or mark nullable |
| `SA1200: Using directives should be placed correctly` | Using placement | Move using outside namespace |
| `CA1062: Validate parameter is non-null` | Missing null check | Add `ArgumentNullException.ThrowIfNull(param)` |

### Common Go Issues

| Error | Cause | Fix |
|-------|-------|-----|
| `G104: Errors unhandled` | Ignored error return | Handle or explicitly ignore with `_ = func()` |
| `SA4006: Assigned value is never used` | Unused assignment | Remove or use the variable |
| `ST1003: Incorrect naming` | Non-idiomatic name | Follow Go naming conventions |
| `errcheck: Error return value not checked` | Missing error handling | Check and handle the error |

### Common TypeScript/React Issues

| Error | Cause | Fix |
|-------|-------|-----|
| `@typescript-eslint/no-unused-vars` | Unused variable | Remove or prefix with `_` |
| `react-hooks/exhaustive-deps` | Missing dependency | Add to dependency array |
| `@typescript-eslint/no-explicit-any` | Using `any` type | Use specific type or `unknown` |
| `import/order` | Wrong import order | Reorder imports or run auto-fix |

---

## Suppressing Warnings (When Necessary)

### .NET

```csharp
// Single line
#pragma warning disable CA1062
ValidateInput(input);
#pragma warning restore CA1062

// Method/class level
[SuppressMessage("Design", "CA1062:Validate arguments", Justification = "Validated by framework")]
public void MyMethod(string input) { }
```

### Go

```go
// Single line
//nolint:errcheck
_ = file.Close()

// Function level
//nolint:gocyclo
func complexFunction() { }

// With reason
//nolint:gosec // G104: Intentionally ignoring this error
```

### TypeScript

```typescript
// Single line
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const data: any = response.data;

// Block
/* eslint-disable react-hooks/exhaustive-deps */
useEffect(() => {
  // Effect code
}, []); // Intentionally empty deps
/* eslint-enable react-hooks/exhaustive-deps */
```

---

## Taskfile.yml Complete Template

```yaml
# Taskfile.yml - Code Quality Tasks
version: '3'

vars:
  SERVICE_NAME: my-service

tasks:
  # ===========================================
  # .NET Tasks
  # ===========================================
  dotnet:format:
    desc: Format .NET code
    cmds:
      - dotnet format --verbosity normal

  dotnet:lint:
    desc: Run .NET analyzers
    cmds:
      - dotnet build /p:EnforceCodeStyleInBuild=true

  dotnet:quality:
    desc: Run all .NET code quality checks
    cmds:
      - task: dotnet:format
      - task: dotnet:lint

  # ===========================================
  # Go Tasks
  # ===========================================
  go:fmt:
    desc: Format Go code
    cmds:
      - gofmt -w -s .
      - goimports -w .

  go:lint:
    desc: Run Go linters
    cmds:
      - golangci-lint run

  go:quality:
    desc: Run all Go code quality checks
    cmds:
      - task: go:fmt
      - task: go:lint

  # ===========================================
  # Python Tasks
  # ===========================================
  py:format:
    desc: Format Python code
    cmds:
      - black .
      - isort .

  py:lint:
    desc: Run Python linters
    cmds:
      - ruff check .
      - mypy src/

  py:quality:
    desc: Run all Python code quality checks
    cmds:
      - task: py:format
      - task: py:lint

  # ===========================================
  # TypeScript/React Tasks
  # ===========================================
  ts:format:
    desc: Format TypeScript/React code
    dir: "{{.SERVICE_NAME}}-ui"
    cmds:
      - npx prettier --write "src/**/*.{ts,tsx,css,json}"

  ts:lint:
    desc: Run TypeScript/React linters
    dir: "{{.SERVICE_NAME}}-ui"
    cmds:
      - npx eslint src/ --fix
      - npx tsc --noEmit

  ts:quality:
    desc: Run all TypeScript code quality checks
    cmds:
      - task: ts:format
      - task: ts:lint

  # ===========================================
  # Combined Quality Check
  # ===========================================
  quality:
    desc: Run ALL code quality checks
    cmds:
      - task: dotnet:quality
      - task: go:quality
      - task: ts:quality
```

---

## Summary

```
╔══════════════════════════════════════════════════════════════════════════════════════════╗
║                        ✅ CODE QUALITY CHECKLIST                                         ║
╠══════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                          ║
║   Before proceeding to Phase 3 (Unit Tests), verify:                                     ║
║                                                                                          ║
║   □ All formatters have been run (auto-fix applied)                                      ║
║   □ All linters pass with no errors                                                      ║
║   □ Warnings are either fixed or documented with justification                           ║
║   □ Type checking passes (TypeScript/Python)                                             ║
║   □ Config files are committed (.editorconfig, .golangci.yml, etc.)                      ║
║                                                                                          ║
║   Command summary:                                                                       ║
║   • .NET:       dotnet format && dotnet build /p:EnforceCodeStyleInBuild=true           ║
║   • Go:         gofmt -w . && golangci-lint run                                         ║
║   • Python:     black . && isort . && ruff check . && mypy src/                         ║
║   • TypeScript: npx prettier --write . && npx eslint src/ && npx tsc --noEmit           ║
║                                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════════════════╝
```
````
