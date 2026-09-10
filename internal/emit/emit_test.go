package emit_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/emit"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
)

func view(t *testing.T) *emit.View {
	t.Helper()
	return emit.NewView(palette.Default(), "test")
}

// TestEveryTargetRenders catches a template that references a field the View does
// not have — which text/template only reports at execution time, so a broken
// template compiles and ships otherwise.
func TestEveryTargetRenders(t *testing.T) {
	v := view(t)
	p := palette.Default()
	bg := p.MustColor("background")

	for _, target := range emit.Targets() {
		t.Run(target.Name, func(t *testing.T) {
			out, err := emit.Render(target, v)
			if err != nil {
				t.Fatal(err)
			}
			s := string(out)
			if len(strings.TrimSpace(s)) == 0 {
				t.Fatal("rendered empty")
			}
			// Every format carries the background somewhere, in one spelling or
			// the other. If it does not, the template is wired to the wrong field.
			if !strings.Contains(s, bg.Hex()) && !strings.Contains(s, strings.TrimPrefix(bg.Hex(), "#")) &&
				!strings.Contains(s, "0.0784313725") {
				t.Errorf("output never mentions the background color %s", bg.Hex())
			}
			if strings.Contains(s, "<no value>") {
				t.Error("template referenced a field the View does not have")
			}
			if !strings.HasSuffix(s, "\n") {
				t.Error("output does not end in a newline")
			}
		})
	}
}

func TestJSONTargetsAreValidJSON(t *testing.T) {
	v := view(t)
	for _, target := range emit.Targets() {
		if !strings.HasSuffix(target.Path, ".json") {
			continue
		}
		out, err := emit.Render(target, v)
		if err != nil {
			t.Fatal(err)
		}
		var any map[string]interface{}
		if err := json.Unmarshal(out, &any); err != nil {
			t.Errorf("%s is not valid JSON: %v", target.Name, err)
		}
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	// dist/ is committed and CI diffs it, so any map-iteration order leaking into
	// output would make the check flap.
	for _, target := range emit.Targets() {
		a, err := emit.Render(target, view(t))
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			b, err := emit.Render(target, view(t))
			if err != nil {
				t.Fatal(err)
			}
			if string(a) != string(b) {
				t.Fatalf("%s renders differently between runs", target.Name)
			}
		}
	}
}

func TestNamed(t *testing.T) {
	all, err := emit.Named("")
	if err != nil || len(all) != len(emit.Targets()) {
		t.Fatalf("empty filter should select everything: %d, %v", len(all), err)
	}
	one, err := emit.Named("ghostty")
	if err != nil || len(one) != 1 || one[0].Name != "ghostty" {
		t.Fatalf("single filter: %v, %v", one, err)
	}
	two, err := emit.Named("ghostty, kitty")
	if err != nil || len(two) != 2 {
		t.Fatalf("comma filter: %v, %v", two, err)
	}
	_, err = emit.Named("nosuchterminal")
	if err == nil {
		t.Fatal("unknown target should error")
	}
	if !strings.Contains(err.Error(), "ghostty") {
		t.Errorf("error should list the known targets, got: %v", err)
	}
}
