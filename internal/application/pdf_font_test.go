package application

import (
	"os"
	"path/filepath"
	"testing"
)

func repoFontDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../fonts")
	if err != nil {
		t.Skip(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "NotoSansDevanagari-Regular.ttf")); err != nil {
		t.Skip("devanagari font not present")
	}
	return dir
}

func TestLoadFonts_Devanagari(t *testing.T) {
	loaded := map[string]bool{}
	fonts := loadFonts(repoFontDir(t), loaded)
	if !loaded["devanagari"] {
		t.Fatal("devanagari family not loaded")
	}
	if len(fonts) == 0 {
		t.Fatal("no fonts returned")
	}
}

func TestSplitRuns_RoutesByScript(t *testing.T) {
	g := &PDFGenerator{loaded: map[string]bool{"devanagari": true, "latin": true}, family: "latin"}
	runs := g.splitRuns("Flat 4, गांधी नगर")
	if len(runs) != 2 {
		t.Fatalf("got %d runs, want 2: %+v", len(runs), runs)
	}
	if runs[0].family != "latin" {
		t.Errorf("latin run family = %q, want latin", runs[0].family)
	}
	if runs[1].family != "devanagari" {
		t.Errorf("hindi run family = %q, want devanagari", runs[1].family)
	}
	if got := g.splitRuns("Meera Devi"); len(got) != 1 || got[0].family != "latin" {
		t.Errorf("ascii value routed to %+v, want single latin run", got)
	}
	if got := g.splitRuns("पूरा नाम"); len(got) != 1 || got[0].family != "devanagari" {
		t.Errorf("hindi value routed to %+v, want single devanagari run", got)
	}
	// Digits and punctuation attach to the current run instead of splitting.
	if got := g.splitRuns("मकान 12"); len(got) != 1 || got[0].family != "devanagari" {
		t.Errorf("hindi+digits routed to %+v, want single devanagari run", got)
	}
}
