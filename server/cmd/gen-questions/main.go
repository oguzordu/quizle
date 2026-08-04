// Command gen-questions writes the programmatically generated flags and
// capitals packs to server/data/questions/. Re-run it whenever
// questions.WellKnownCountries changes.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oguzordu/quizle/internal/questions"
)

func main() {
	flags, capitals := questions.GenerateFlagAndCapitalQuestions(questions.WellKnownCountries)

	outDir := "data/questions"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "gen-questions:", err)
		os.Exit(1)
	}

	if err := writePack(filepath.Join(outDir, "flags.json"), flags); err != nil {
		fmt.Fprintln(os.Stderr, "gen-questions:", err)
		os.Exit(1)
	}
	if err := writePack(filepath.Join(outDir, "capitals.json"), capitals); err != nil {
		fmt.Fprintln(os.Stderr, "gen-questions:", err)
		os.Exit(1)
	}

	fmt.Printf("%d bayrak sorusu ve %d başkent sorusu yazıldı\n", len(flags), len(capitals))
}

func writePack(path string, qs []questions.Question) error {
	b, err := json.MarshalIndent(qs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
