// dcinit — scaffold a lean devcontainer for a chosen set of languages.
//
// Usage:
//   dcinit <lang> [<lang> ...] [--no-tools] [--name NAME] [--force]
//
// Example:
//   dcinit python rust lua
//
// Writes ./.devcontainer/devcontainer.json and ./.mise.toml, pulling the lean
// base image, adding only the build-deps Features the chosen languages need,
// and listing pinned (never "latest") tool versions in .mise.toml.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ─── Configuration: the single source of truth ──────────────────────────────
// To add a language later, add one entry here. Nothing else changes.

const (
	baseImage       = "ghcr.io/yjalil/devbase:latest"
	featureNS       = "ghcr.io/yjalil/devcontainers-toolkit"
	devToolsFeature = featureNS + "/dev-tools:1"
)

// lang describes how one language is provisioned.
//   miseTool:    the key written into [tools] in .mise.toml
//   miseVersion: pinned LTS/stable default (NEVER "latest")
//   feature:     "" if no root system libs are needed, else the Feature id (e.g. "ruby")
//   persistDirs: home-relative dirs this language's runtime writes to that must be
//                persisted in the volume (e.g. rustup/cargo for rust). Without this,
//                the toolchain re-downloads on every rebuild. Names are relative to $HOME.
type lang struct {
	miseTool    string
	miseVersion string
	feature     string
	persistDirs []string
}

// languages maps a CLI argument -> provisioning.
// Versions are deliberate pins; bump them here when you choose to move.
var languages = map[string]lang{
	"python":     {miseTool: "python", miseVersion: "3.12", feature: ""},
	"node":       {miseTool: "node", miseVersion: "lts", feature: "", persistDirs: []string{".npm"}},
	"typescript": {miseTool: "node", miseVersion: "lts", feature: "", persistDirs: []string{".npm"}}, // alias: TS rides on node
	"go":         {miseTool: "go", miseVersion: "1.23", feature: "", persistDirs: []string{"go"}},
	"rust":       {miseTool: "rust", miseVersion: "stable", feature: "", persistDirs: []string{".rustup", ".cargo"}},
	"zig":        {miseTool: "zig", miseVersion: "0.13.0", feature: ""},
	"ruby":       {miseTool: "ruby", miseVersion: "3.3", feature: "ruby"},
	"haskell":    {miseTool: "ghc", miseVersion: "9.8", feature: "haskell", persistDirs: []string{".ghcup", ".cabal", ".stack"}},
	"lua":        {miseTool: "lua", miseVersion: "5.4", feature: "lua"},
}

// ─── Templates ───────────────────────────────────────────────────────────────

// basePersist are the dirs every container persists: mise toolchains, generic
// cache, uv, and zsh history. Language-specific dirs (rustup, go, etc.) are
// appended based on the chosen languages.
//   mapping: HOME-relative dir  ->  volume subdir name
var basePersist = [][2]string{
	{".local/share/mise", "mise"},
	{".cache", "cache"},
	{".local/share/uv", "uv"},
}

// buildOnCreate constructs the onCreateCommand: it redirects tool/state dirs
// into the persistent volume (via symlink) so toolchains and caches survive
// rebuilds instead of re-downloading. extraDirs are HOME-relative dirs from the
// chosen languages (e.g. ".rustup", "go").
func buildOnCreate(extraDirs []string) string {
	// Build the full mapping: base + language dirs. Volume subdir = last path
	// element with dots stripped (".rustup" -> "rustup", "go" -> "go").
	type pair struct{ home, vol string }
	pairs := []pair{}
	for _, b := range basePersist {
		pairs = append(pairs, pair{b[0], b[1]})
	}
	for _, d := range extraDirs {
		vol := strings.TrimPrefix(filepath.Base(d), ".")
		pairs = append(pairs, pair{d, vol})
	}

	// zsh history dir is always created (history lives there), no symlink needed.
	volDirs := []string{"~/persistent-data/zsh"}
	var rmTargets, links []string
	for _, p := range pairs {
		volDirs = append(volDirs, "~/persistent-data/"+p.vol)
		rmTargets = append(rmTargets, "~/"+p.home)
		links = append(links, fmt.Sprintf("ln -sfn ~/persistent-data/%s ~/%s", p.vol, p.home))
	}

	parts := []string{
		"sudo chown -R vscode:vscode ~/persistent-data",
		"mkdir -p " + strings.Join(volDirs, " "),
		"rm -rf " + strings.Join(rmTargets, " "),
		"mkdir -p ~/.local/share", // parent for mise/uv symlinks
	}
	parts = append(parts, links...)
	return strings.Join(parts, " && ")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	var (
		langArgs []string
		noTools  bool
		force    bool
		name     string
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--no-tools":
			noTools = true
		case a == "--force":
			force = true
		case a == "--name":
			if i+1 >= len(args) {
				fail("--name requires a value")
			}
			i++
			name = args[i]
		case a == "-h" || a == "--help":
			usage()
			os.Exit(0)
		case strings.HasPrefix(a, "-"):
			fail("unknown flag: " + a)
		default:
			langArgs = append(langArgs, strings.ToLower(a))
		}
	}

	if len(langArgs) == 0 {
		fail("specify at least one language (see --help)")
	}

	// Resolve & validate languages.
	chosen := map[string]lang{} // keyed by miseTool to dedupe (node+typescript)
	featureSet := map[string]bool{}
	persistSet := map[string]bool{} // HOME-relative dirs to persist, deduped
	for _, name := range langArgs {
		l, ok := languages[name]
		if !ok {
			fail(fmt.Sprintf("unknown language %q (supported: %s)", name, supportedList()))
		}
		chosen[l.miseTool] = l
		if l.feature != "" {
			featureSet[l.feature] = true
		}
		for _, d := range l.persistDirs {
			persistSet[d] = true
		}
	}

	// Project name defaults to the current directory's basename.
	if name == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fail("cannot determine working directory: " + err.Error())
		}
		name = filepath.Base(cwd)
	}

	onCreate := buildOnCreate(sortedSet(persistSet))
	devcontainerJSON := renderDevcontainer(name, featureSet, !noTools, onCreate)
	miseTOML := renderMise(chosen)

	// Write files.
	if err := os.MkdirAll(".devcontainer", 0o755); err != nil {
		fail("mkdir .devcontainer: " + err.Error())
	}
	writeFile(filepath.Join(".devcontainer", "devcontainer.json"), devcontainerJSON, force)
	writeFile(".mise.toml", miseTOML, force)

	fmt.Printf("✓ scaffolded devcontainer for: %s\n", strings.Join(sortedKeys(chosen), ", "))
	if len(featureSet) > 0 {
		fmt.Printf("  features: %s\n", strings.Join(sortedSet(featureSet), ", "))
	}
	if !noTools {
		fmt.Println("  dev-tools: included (use --no-tools to omit)")
	}
}

// renderDevcontainer builds the devcontainer.json text.
func renderDevcontainer(name string, featureSet map[string]bool, withTools bool, onCreate string) string {
	// Collect feature reference lines.
	var featureLines []string
	if withTools {
		featureLines = append(featureLines, fmt.Sprintf("    %q: {}", devToolsFeature))
	}
	for _, f := range sortedSet(featureSet) {
		ref := fmt.Sprintf("%s/%s:1", featureNS, f)
		featureLines = append(featureLines, fmt.Sprintf("    %q: {}", ref))
	}

	features := "{}"
	if len(featureLines) > 0 {
		features = "{\n" + strings.Join(featureLines, ",\n") + "\n  }"
	}

	return fmt.Sprintf(`{
  "name": %q,
  "image": %q,
  "features": %s,

  "workspaceFolder": "/workspaces/${localWorkspaceFolderBasename}",

  "mounts": [
    "source=${localWorkspaceFolderBasename}-data,target=/home/vscode/persistent-data,type=volume"
  ],

  "onCreateCommand": %q,
  "postCreateCommand": "mise install -y && echo ready",

  "remoteUser": "vscode"
}
`, name, baseImage, features, onCreate)
}

// renderMise builds the .mise.toml text with pinned versions.
func renderMise(chosen map[string]lang) string {
	var b strings.Builder
	b.WriteString("# Pinned versions — never \"latest\". Bump deliberately as needed.\n")
	b.WriteString("[tools]\n")
	for _, tool := range sortedKeys(chosen) {
		l := chosen[tool]
		b.WriteString(fmt.Sprintf("%s = %q\n", l.miseTool, l.miseVersion))
	}
	return b.String()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func writeFile(path, content string, force bool) {
	if !force {
		if _, err := os.Stat(path); err == nil {
			fail(fmt.Sprintf("%s already exists (use --force to overwrite)", path))
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fail("write " + path + ": " + err.Error())
	}
}

func supportedList() string {
	keys := make([]string, 0, len(languages))
	for k := range languages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func sortedKeys(m map[string]lang) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedSet(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func usage() {
	fmt.Print(`dcinit — scaffold a lean devcontainer for a set of languages

Usage:
  dcinit <lang> [<lang> ...] [flags]

Flags:
  --no-tools     omit the dev-tools feature (jq, psql, debug kit)
  --name NAME    project name (default: current directory)
  --force        overwrite existing files
  -h, --help     show this help

Supported languages:
  python, node, typescript, go, rust, zig, ruby, haskell, lua

Examples:
  dcinit python
  dcinit python rust lua
  dcinit go --no-tools
`)
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "dcinit: "+msg)
	os.Exit(1)
}