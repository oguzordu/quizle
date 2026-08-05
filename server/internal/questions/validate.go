// Package questions defines Quizle's question-bank format and validates it.
// A question with a validation error must never reach players: this package
// is what CI runs against every JSON file in server/data/questions before a
// pack can ship.
package questions

import (
	"fmt"
	"strings"
)

// Localized holds one language's rendering of a question.
type Localized struct {
	Text    string   `json:"text"`
	Choices []string `json:"choices"`
}

// isEmpty reports whether this language block was omitted entirely.
func (l Localized) isEmpty() bool {
	return strings.TrimSpace(l.Text) == "" && len(l.Choices) == 0
}

// Question is one round with a mandatory source citation. EN is required;
// TR is optional, and a question without it is served only to English rooms.
type Question struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Correct   int       `json:"correct"`
	SourceURL string    `json:"source_url"`
	Image     string    `json:"image,omitempty"`
	TR        Localized `json:"tr"`
	EN        Localized `json:"en"`
}

const requiredChoiceCount = 4

// ValidatePack checks every question in a pack and returns one error per
// problem found (not just the first), so a CI run reports everything wrong
// at once instead of forcing fix-rerun-fix cycles.
func ValidatePack(qs []Question) []error {
	var errs []error
	seenIDs := make(map[string]bool)

	for i, q := range qs {
		label := q.ID
		if label == "" {
			label = fmt.Sprintf("#%d", i)
		}

		if q.ID == "" {
			errs = append(errs, fmt.Errorf("%s: missing id", label))
		} else if seenIDs[q.ID] {
			errs = append(errs, fmt.Errorf("%s: duplicate id", label))
		}
		seenIDs[q.ID] = true

		if q.Category == "" {
			errs = append(errs, fmt.Errorf("%s: missing category", label))
		}
		if strings.TrimSpace(q.SourceURL) == "" {
			errs = append(errs, fmt.Errorf("%s: missing source_url", label))
		}

		// English is mandatory; Turkish is optional so English-only imports
		// (OpenTDB) can ship and be served to English rooms. A TR block that
		// is present must still be complete — optional means "may be
		// absent", not "may be broken".
		errs = append(errs, validateLocalized(label, "en", q.EN, q.Correct)...)
		if !q.TR.isEmpty() {
			errs = append(errs, validateLocalized(label, "tr", q.TR, q.Correct)...)
		}
	}

	return errs
}

func validateLocalized(label, lang string, l Localized, correct int) []error {
	var errs []error

	if strings.TrimSpace(l.Text) == "" {
		errs = append(errs, fmt.Errorf("%s: missing %s text", label, lang))
	}

	if len(l.Choices) != requiredChoiceCount {
		errs = append(errs, fmt.Errorf("%s: %s has %d choices, want %d", label, lang, len(l.Choices), requiredChoiceCount))
		return errs // index/duplicate checks below assume exactly 4 choices
	}

	if correct < 0 || correct >= len(l.Choices) {
		errs = append(errs, fmt.Errorf("%s: correct index %d out of range for %s choices", label, correct, lang))
	}

	seen := make(map[string]bool, len(l.Choices))
	for _, c := range l.Choices {
		key := strings.ToLower(strings.TrimSpace(c))
		if key == "" {
			errs = append(errs, fmt.Errorf("%s: %s has an empty choice", label, lang))
			continue
		}
		if seen[key] {
			errs = append(errs, fmt.Errorf("%s: %s has duplicate choice %q", label, lang, c))
		}
		seen[key] = true
	}

	return errs
}
