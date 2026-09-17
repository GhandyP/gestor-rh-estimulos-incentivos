package main

import (
	"strings"
	"testing"
)

// Spec (operable-delivery): "Clean checkout passes checks" y threat matrix
// "PR commands": el workflow de CI debe correr las cuatro verificaciones
// (test/build/vet/gofmt) fijadas a la raíz del repo y fallar en cualquier
// salida no cero, sin componer input de usuario ni escribir en el repo.
// RED: el workflow no existe todavía.

func TestCIWorkflowRunsAllFourChecksAtRepoRoot(t *testing.T) {
	ci := readRepoFile(t, ".github/workflows/ci.yml")

	for _, want := range []string{
		"go test ./...",
		"go build ./...",
		"go vet ./...",
		"gofmt -l .",
		"working-directory: ${{ github.workspace }}", // fijado a la raíz del repo
		"push:",
		"pull_request:",
		"go-version-file: go.mod",
		"./scripts/compose-smoke.sh",
	} {
		if !strings.Contains(ci, want) {
			t.Errorf("ci.yml no contiene %q", want)
		}
	}
}

// Threat matrix: "CI does not commit" — el workflow nunca escribe en el repo.
func TestCIHasNoWriteAutomation(t *testing.T) {
	ci := readRepoFile(t, ".github/workflows/ci.yml")
	for _, bad := range []string{"git commit", "git push", "git config"} {
		if strings.Contains(ci, bad) {
			t.Errorf("ci.yml no debe contener %q (CI no escribe en el repo)", bad)
		}
	}
}

// Threat matrix: "no composed user input" — ningún evento del PR/issue se
// interpola en los scripts de shell (github.event.* queda fuera del run).
func TestCINoComposedUserInput(t *testing.T) {
	ci := readRepoFile(t, ".github/workflows/ci.yml")
	if strings.Contains(ci, "github.event") {
		t.Error("ci.yml no debe componer input del evento en los scripts")
	}
}
