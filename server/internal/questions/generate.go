package questions

import "fmt"

// Country holds everything needed to generate a flag question and a capital
// question for one country, in both supported languages, with a real,
// checkable source for each fact.
type Country struct {
	ISO              string
	NameTR, NameEN   string
	CapitalTR        string
	CapitalEN        string
	FlagSourceURL    string
	CountrySourceURL string
}

// GenerateFlagAndCapitalQuestions builds one flag-identification and one
// capital-identification question per country. Distractors are the next
// three countries in the list (wrapping around), so output is fully
// deterministic and reviewable — no randomness to second-guess in a diff.
func GenerateFlagAndCapitalQuestions(countries []Country) (flags, capitals []Question) {
	n := len(countries)
	flags = make([]Question, 0, n)
	capitals = make([]Question, 0, n)

	for i, c := range countries {
		distractorIdx := []int{(i + 1) % n, (i + 2) % n, (i + 3) % n}
		correctPos := i % 4

		trChoices, enChoices := buildChoices(correctPos, c.NameTR, c.NameEN, countries, distractorIdx, func(d Country) (string, string) {
			return d.NameTR, d.NameEN
		})
		flags = append(flags, Question{
			ID:        fmt.Sprintf("flag-%s", c.ISO),
			Category:  "flags",
			Correct:   correctPos,
			SourceURL: c.FlagSourceURL,
			Image:     fmt.Sprintf("flag:%s", c.ISO),
			TR:        Localized{Text: "Bu hangi ülkenin bayrağı?", Choices: trChoices},
			EN:        Localized{Text: "Which country does this flag belong to?", Choices: enChoices},
		})

		trCapChoices, enCapChoices := buildChoices(correctPos, c.CapitalTR, c.CapitalEN, countries, distractorIdx, func(d Country) (string, string) {
			return d.CapitalTR, d.CapitalEN
		})
		capitals = append(capitals, Question{
			ID:        fmt.Sprintf("capital-%s", c.ISO),
			Category:  "capitals",
			Correct:   correctPos,
			SourceURL: c.CountrySourceURL,
			TR:        Localized{Text: fmt.Sprintf("%s ülkesinin başkenti neresidir?", c.NameTR), Choices: trCapChoices},
			EN:        Localized{Text: fmt.Sprintf("What is the capital of %s?", c.NameEN), Choices: enCapChoices},
		})
	}

	return flags, capitals
}

// buildChoices places the correct TR/EN answer at correctPos and fills the
// remaining three slots with the corresponding field from the countries at
// distractorIdx, in order.
func buildChoices(correctPos int, correctTR, correctEN string, countries []Country, distractorIdx []int, pick func(Country) (string, string)) (tr, en []string) {
	tr = make([]string, 4)
	en = make([]string, 4)
	tr[correctPos] = correctTR
	en[correctPos] = correctEN

	di := 0
	for slot := 0; slot < 4; slot++ {
		if slot == correctPos {
			continue
		}
		d := countries[distractorIdx[di]]
		dTR, dEN := pick(d)
		tr[slot] = dTR
		en[slot] = dEN
		di++
	}
	return tr, en
}
