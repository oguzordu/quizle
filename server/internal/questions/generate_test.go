package questions

import "testing"

func TestGenerateFlagAndCapitalQuestions_producesFourChoicesEach(t *testing.T) {
	countries := []Country{
		{ISO: "TR", NameTR: "Türkiye", NameEN: "Turkey", CapitalTR: "Ankara", CapitalEN: "Ankara", FlagSourceURL: "https://en.wikipedia.org/wiki/Flag_of_Turkey", CountrySourceURL: "https://en.wikipedia.org/wiki/Turkey"},
		{ISO: "DE", NameTR: "Almanya", NameEN: "Germany", CapitalTR: "Berlin", CapitalEN: "Berlin", FlagSourceURL: "https://en.wikipedia.org/wiki/Flag_of_Germany", CountrySourceURL: "https://en.wikipedia.org/wiki/Germany"},
		{ISO: "FR", NameTR: "Fransa", NameEN: "France", CapitalTR: "Paris", CapitalEN: "Paris", FlagSourceURL: "https://en.wikipedia.org/wiki/Flag_of_France", CountrySourceURL: "https://en.wikipedia.org/wiki/France"},
		{ISO: "IT", NameTR: "İtalya", NameEN: "Italy", CapitalTR: "Roma", CapitalEN: "Rome", FlagSourceURL: "https://en.wikipedia.org/wiki/Flag_of_Italy", CountrySourceURL: "https://en.wikipedia.org/wiki/Italy"},
	}

	flags, capitals := GenerateFlagAndCapitalQuestions(countries)

	if len(flags) != len(countries) {
		t.Fatalf("len(flags) = %d, want %d", len(flags), len(countries))
	}
	if len(capitals) != len(countries) {
		t.Fatalf("len(capitals) = %d, want %d", len(capitals), len(countries))
	}

	for _, q := range flags {
		if q.Category != "flags" {
			t.Errorf("flags[%s].Category = %q, want flags", q.ID, q.Category)
		}
	}
	for _, q := range capitals {
		if q.Category != "capitals" {
			t.Errorf("capitals[%s].Category = %q, want capitals", q.ID, q.Category)
		}
	}

	all := append(append([]Question{}, flags...), capitals...)
	if errs := ValidatePack(all); len(errs) != 0 {
		t.Fatalf("generated pack fails validation: %v", errs)
	}
}

func TestGenerateFlagAndCapitalQuestions_correctChoiceMatchesCountry(t *testing.T) {
	countries := []Country{
		{ISO: "TR", NameTR: "Türkiye", NameEN: "Turkey", CapitalTR: "Ankara", CapitalEN: "Ankara", FlagSourceURL: "u1", CountrySourceURL: "u2"},
		{ISO: "DE", NameTR: "Almanya", NameEN: "Germany", CapitalTR: "Berlin", CapitalEN: "Berlin", FlagSourceURL: "u3", CountrySourceURL: "u4"},
		{ISO: "FR", NameTR: "Fransa", NameEN: "France", CapitalTR: "Paris", CapitalEN: "Paris", FlagSourceURL: "u5", CountrySourceURL: "u6"},
		{ISO: "IT", NameTR: "İtalya", NameEN: "Italy", CapitalTR: "Roma", CapitalEN: "Rome", FlagSourceURL: "u7", CountrySourceURL: "u8"},
	}

	flags, capitals := GenerateFlagAndCapitalQuestions(countries)

	trFlag := flags[0]
	if trFlag.EN.Choices[trFlag.Correct] != "Turkey" {
		t.Errorf("flag correct choice = %q, want Turkey", trFlag.EN.Choices[trFlag.Correct])
	}

	trCapital := capitals[0]
	if trCapital.EN.Choices[trCapital.Correct] != "Ankara" {
		t.Errorf("capital correct choice = %q, want Ankara", trCapital.EN.Choices[trCapital.Correct])
	}
}
