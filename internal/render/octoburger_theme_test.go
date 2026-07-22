package render

import (
	"os"
	"strings"
	"testing"
)

func TestOctoburgerThemeUsesLocalAccessibleContract(t *testing.T) {
	templateBytes, err := os.ReadFile("../../templates/layout-octoburger-cunty.html")
	if err != nil {
		t.Fatal(err)
	}
	source := string(templateBytes)
	for name, marker := range map[string]string{
		"viewport":          `<meta name="viewport"`,
		"main landmark":     `<main id="main-content"`,
		"skip link":         `href="#main-content"`,
		"octoburger":        `OCTOBURGER`,
		"keyboard menu":     `<details class="octo-menu">`,
		"canonical opt-in":  `{{if .CanonicalURL}}`,
		"local foundations": `/assets/css/theme-foundations.css`,
		"local theme css":   `/assets/css/octoburger-cunty.css`,
	} {
		if !strings.Contains(source, marker) {
			t.Errorf("missing %s marker %q", name, marker)
		}
	}
	if strings.Contains(source, "cdn.") || strings.Contains(source, "tailwindcss") || strings.Contains(source, "daisyui") {
		t.Fatal("octoburger theme must not depend on CDN, Tailwind, or DaisyUI")
	}
	if strings.Count(source, "<h1") != 1 {
		t.Fatalf("expected one h1, found %d", strings.Count(source, "<h1"))
	}
}

func TestOctoburgerStylesAreScopedAndLocal(t *testing.T) {
	css, err := os.ReadFile("../../assets/css/octoburger-cunty.css")
	if err != nil {
		t.Fatal(err)
	}
	source := string(css)
	for _, marker := range []string{".octo-theme", ".octo-menu", ":focus-visible", "@media"} {
		if !strings.Contains(source, marker) {
			t.Errorf("theme CSS missing %s", marker)
		}
	}
}
