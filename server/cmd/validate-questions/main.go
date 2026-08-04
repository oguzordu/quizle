// Command validate-questions checks every JSON pack in server/data/questions
// against questions.ValidatePack and exits non-zero if anything fails. CI
// runs this on every push so a badly formed or unsourced question can never
// reach players.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oguzordu/quizle/internal/questions"
)

func main() {
	dir := "data/questions"
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "validate-questions:", err)
		os.Exit(1)
	}

	totalErrors := 0
	totalQuestions := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			totalErrors++
			continue
		}

		var qs []questions.Question
		if err := json.Unmarshal(b, &qs); err != nil {
			fmt.Fprintf(os.Stderr, "%s: invalid JSON: %v\n", path, err)
			totalErrors++
			continue
		}

		totalQuestions += len(qs)
		for _, verr := range questions.ValidatePack(qs) {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, verr)
			totalErrors++
		}
	}

	if totalErrors > 0 {
		fmt.Fprintf(os.Stderr, "\n%d hata bulundu\n", totalErrors)
		os.Exit(1)
	}
	fmt.Printf("%d soru doğrulandı, hata yok\n", totalQuestions)
}
