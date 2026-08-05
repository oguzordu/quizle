package questions

import "testing"

func validQuestion() Question {
	return Question{
		ID:        "flag-tr",
		Category:  "flags",
		Correct:   1,
		SourceURL: "https://en.wikipedia.org/wiki/Flag_of_Turkey",
		TR:        Localized{Text: "Bu hangi ülkenin bayrağı?", Choices: []string{"Almanya", "Türkiye", "İtalya", "İspanya"}},
		EN:        Localized{Text: "Which country's flag is this?", Choices: []string{"Germany", "Turkey", "Italy", "Spain"}},
	}
}

func TestValidatePack_acceptsWellFormedQuestion(t *testing.T) {
	errs := ValidatePack([]Question{validQuestion()})
	if len(errs) != 0 {
		t.Fatalf("errs = %v, want none", errs)
	}
}

func TestValidatePack_rejectsWrongChoiceCount(t *testing.T) {
	q := validQuestion()
	q.TR.Choices = []string{"Almanya", "Türkiye", "İtalya"}
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for 3 choices, got none")
	}
}

func TestValidatePack_rejectsOutOfRangeCorrectIndex(t *testing.T) {
	q := validQuestion()
	q.Correct = 4
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for out-of-range Correct, got none")
	}
}

func TestValidatePack_rejectsDuplicateChoicesCaseInsensitive(t *testing.T) {
	q := validQuestion()
	q.EN.Choices = []string{"Germany", "germany", "Italy", "Spain"}
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate choices, got none")
	}
}

func TestValidatePack_rejectsMissingTranslation(t *testing.T) {
	q := validQuestion()
	q.EN.Text = ""
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for missing EN text, got none")
	}
}

// Questions imported from English-only sources (OpenTDB) ship without a TR
// block; they're served only to English rooms rather than blocked entirely.
func TestValidatePack_acceptsEnglishOnlyQuestion(t *testing.T) {
	q := validQuestion()
	q.TR = Localized{}
	errs := ValidatePack([]Question{q})
	if len(errs) != 0 {
		t.Fatalf("errs = %v, want none for an English-only question", errs)
	}
}

func TestValidatePack_rejectsMissingEnglish(t *testing.T) {
	q := validQuestion()
	q.EN = Localized{}
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for a question with no EN block, got none")
	}
}

// A present-but-malformed TR block is still a hard error: optional means
// "may be absent", not "may be broken".
func TestValidatePack_rejectsMalformedPresentTurkishBlock(t *testing.T) {
	q := validQuestion()
	q.TR.Choices = []string{"Almanya", "Türkiye"}
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for a TR block with 2 choices, got none")
	}
}

func TestValidatePack_rejectsMissingSourceURL(t *testing.T) {
	q := validQuestion()
	q.SourceURL = ""
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for missing source_url, got none")
	}
}

func TestValidatePack_rejectsDuplicateIDs(t *testing.T) {
	q1 := validQuestion()
	q2 := validQuestion()
	errs := ValidatePack([]Question{q1, q2})
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate question IDs, got none")
	}
}

func TestValidatePack_rejectsMissingCategory(t *testing.T) {
	q := validQuestion()
	q.Category = ""
	errs := ValidatePack([]Question{q})
	if len(errs) == 0 {
		t.Fatal("expected error for missing category, got none")
	}
}
