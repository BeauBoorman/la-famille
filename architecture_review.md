# Component Mapping & Micro-Improvement Audit: La Famille

## Part 1: Component Identification
Here is a mapping of the major components currently residing in the `internal/` directories and their responsibilities:

- **internal/generator**: Orchestrates the core build pipeline. It coordinates gathering metadata, building the graph/search indices, generating stubs for missing files, transforming content, rendering pages, and exporting outputs (RAG, taxonomy, search index).
- **internal/render**: Handles HTML template parsing, caching, and execution. It provides a safely cached `Renderer` struct and manages the injection of live-reload scripts for WatchMode.
- **internal/transform**: Manages Markdown AST transformations during rendering. This includes `LinkTransformer` (for rewriting internal links, tracking backlinks, and detecting missing files) and `EmojiKitchenParser` (an inline Goldmark parser).
- **internal/asset**: Manages the copying of static assets from the asset directory to the output directory, evaluating `.gitignore` patterns natively in Go without external processes.
- **internal/search**: Responsible for building a minified JSON search index from processed markdown files to power client-side search functionality.
- **internal/taxonomy**: Generates tag archive pages (HTML stubs) and organizes site content structurally by tags extracted from markdown frontmatter.
- **internal/graph**: Defines the `Graph` and `Node` structures representing the site's internal linking structure and writes this relational data to a `graph.json` file.
- **internal/ragexport**: Handles compiling the site's content into a single, cohesive Markdown file (RAG export) intended for AI ingestion or archiving, respecting explicit exclusions.
- **internal/config**: Loads, validates, and stores application configuration from a YAML file, handling defaults and enforcing safe-path validations.
- **internal/content**: Contains the core logic for walking the content directory, parsing Markdown files, and extracting YAML frontmatter into a structured `FileMeta` object.
- **internal/stub**: Generates placeholder HTML pages (stubs) for internal links that point to files that do not yet exist, showing backlinks to help users navigate.
- **internal/page**: Defines the data model (`Page` struct) that is passed to the HTML templates for rendering individual pages.
- **internal/sitedata**: Writes site metadata to `meta.json` and generates the standard `sitemap.xml` based on the gathered metadata.
- **internal/markdown**: Configures and instantiates the Goldmark Markdown rendering engine with custom extensions (GFM, Typographer) and custom AST transformers.
- **internal/git**: Provides utilities for executing basic local Git commands, checking for uncommitted changes, and parsing remote repository URLs.
- **internal/github**: A simple GitHub API client used to check the status of Pull Requests, sync repository data, and verify CI check runs.
- **internal/watcher**: Implements a file system watcher and an SSE (Server-Sent Events) HTTP handler for the live-reload functionality during development WatchMode.

## Part 2: Micro-Improvements

Here are 5 high-ROI micro-improvements focusing on localized technical debt, performance, and memory packing:

### 1. Optimize Struct Field Alignment (`internal/content/metadata.go`)
**Issue:** The `FileMeta` struct has fields ordered seemingly at random (mix of `[]byte`, `[]string`, `string`, `*bool`). This leads to unnecessary padding and wasted memory for every loaded file.
**Improvement:** Reorder the fields from largest alignment requirement to smallest (pointers/slices first, then strings, then booleans) to pack the struct tightly in memory.

### 2. Pre-allocate String Builders with `sb.Grow()` (`internal/sitedata/write.go` & `internal/stub/stub.go`)
**Issue:** Several places create a `strings.Builder` and immediately write multiple large strings in a loop without pre-allocating the internal buffer, causing multiple heap re-allocations.
**Improvement:** Use `sb.Grow()` based on an estimated or known size before writing. For `sitemap.xml` generation, calculate the expected byte size based on the number of keys. For `stub.go`, pre-calculate the size of the boilerplate HTML block.

### 3. Replace Linear Slice Scans with Maps for Ignore/Exclude Matching (`internal/ragexport/export.go`)
**Issue:** In `ExportRAG`, the `excludeFiles` check currently performs a linear `strings.HasPrefix` check inside a nested loop over all files (resulting in $O(N \times M)$ time complexity).
**Improvement:** Convert exact file exclusions to a `map[string]struct{}` lookup for $O(1)$ checks, falling back to prefix matching only for directory-level exclusions.

### 4. Optimize Slice Pre-allocation for Appends (`internal/generator/generator.go`)
**Issue:** In `generator.go`, when iterating to collect errors (`update.errs = append(update.errs, ...)`), or building intermediate lists, the slices are often appended to from a `nil` state even when the maximum capacity is known (e.g., collecting joinErrs).
**Improvement:** Pre-allocate slices where lengths are known using `make([]T, 0, len(items))` to prevent runtime reallocation overhead during tight loops.

### 5. Improve Error Wrapping Context in `GatherMetadata` (`internal/content/metadata.go`)
**Issue:** When parsing frontmatter fails, it returns `fmt.Errorf("failed to parse frontmatter: %w", err)` or similar, without including the file path that caused the failure.
**Improvement:** Update error wrapping to explicitly include the `path` being processed (e.g., `fmt.Errorf("failed to parse frontmatter for %s: %w", path, err)`) to vastly improve debugging context for invalid markdown files.
