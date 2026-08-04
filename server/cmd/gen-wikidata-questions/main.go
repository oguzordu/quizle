// Command gen-wikidata-questions builds pop-culture trivia questions from
// Wikidata's public SPARQL endpoint, the same way cmd/gen-questions builds
// flag/capital questions from a static country list — but for facts too
// numerous to hand-maintain (Oscar-winning films, their directors, and
// release years). Wikidata's own multilingual labels are used directly
// (not machine translation), which is what keeps the Turkish text trustworthy.
//
// Run it whenever you want to refresh this pack:
//
//	go run ./cmd/gen-wikidata-questions
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/oguzordu/quizle/internal/questions"
)

const sparqlQuery = `
SELECT DISTINCT ?film ?filmLabelTR ?filmLabelEN ?director ?directorLabel ?year WHERE {
  ?film wdt:P166 wd:Q102427.
  ?film wdt:P577 ?date.
  BIND(YEAR(?date) AS ?year)
  ?film wdt:P57 ?director.
  ?film rdfs:label ?filmLabelTR. FILTER(LANG(?filmLabelTR) = "tr")
  ?film rdfs:label ?filmLabelEN. FILTER(LANG(?filmLabelEN) = "en")
  ?director rdfs:label ?directorLabel. FILTER(LANG(?directorLabel) = "en")
}
ORDER BY ?year
`

type sparqlResponse struct {
	Results struct {
		Bindings []map[string]struct {
			Value string `json:"value"`
		} `json:"bindings"`
	} `json:"results"`
}

type oscarFilm struct {
	WikidataID string
	TitleTR    string
	TitleEN    string
	Director   string
	Year       string
}

func main() {
	films, err := fetchOscarBestPictureWinners()
	if err != nil {
		log.Fatalf("wikidata sorgusu başarısız: %v", err)
	}
	fmt.Printf("%d benzersiz film bulundu\n", len(films))

	qs := buildQuestions(films)
	if errs := questions.ValidatePack(qs); len(errs) != 0 {
		log.Fatalf("üretilen sorular geçersiz: %v", errs)
	}

	b, err := json.MarshalIndent(qs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	out := "data/questions/movies-oscars.json"
	if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d soru %s dosyasına yazıldı\n", len(qs), out)
}

func fetchOscarBestPictureWinners() ([]oscarFilm, error) {
	endpoint := "https://query.wikidata.org/sparql?" + url.Values{
		"query":  {sparqlQuery},
		"format": {"json"},
	}.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	// Wikidata asks bots to identify themselves; an anonymous default Go
	// User-Agent gets rate-limited or blocked.
	req.Header.Set("User-Agent", "QuizleQuestionGenerator/1.0 (https://github.com/oguzordu/quizle)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikidata döndürdü: %s: %s", resp.Status, body)
	}

	var parsed sparqlResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	// Some films carry more than one P57 (director) statement in Wikidata
	// with no way to tell which is the "real" credited director from this
	// query alone (e.g. an uncredited co-director sharing equal rank with
	// the actual one — this genuinely produced a wrong answer on the first
	// run, for "Wings"). Rather than guess, track every distinct director
	// seen per film and drop any film where they disagree, so this pack
	// only ever asserts a fact it's actually sure of.
	directorsByFilm := make(map[string]map[string]bool)
	meta := make(map[string]oscarFilm)
	for _, row := range parsed.Results.Bindings {
		id := row["film"].Value
		if directorsByFilm[id] == nil {
			directorsByFilm[id] = make(map[string]bool)
		}
		directorsByFilm[id][row["directorLabel"].Value] = true
		meta[id] = oscarFilm{
			WikidataID: id,
			TitleTR:    row["filmLabelTR"].Value,
			TitleEN:    row["filmLabelEN"].Value,
			Director:   row["directorLabel"].Value,
			Year:       row["year"].Value,
		}
	}

	byFilm := make(map[string]oscarFilm)
	skipped := 0
	for id, directors := range directorsByFilm {
		if len(directors) != 1 {
			skipped++
			continue
		}
		byFilm[id] = meta[id]
	}
	if skipped > 0 {
		fmt.Printf("%d film birden fazla/çelişkili yönetmen kaydı yüzünden atlandı\n", skipped)
	}

	films := make([]oscarFilm, 0, len(byFilm))
	for _, f := range byFilm {
		films = append(films, f)
	}
	sort.Slice(films, func(i, j int) bool { return films[i].Year < films[j].Year })
	return films, nil
}

// buildQuestions generates one "who directed this Best Picture winner"
// question per film. Distractors are the next films in the (year-sorted)
// list, skipping any whose director is the same person (some directors won
// more than once) so no question ever offers the correct name twice.
//
// A "which year did it win" question was deliberately left out: Wikidata's
// P577 (publication date) is the film's release date, not its ceremony
// year, and those two occasionally disagree — not a fact this pack should
// risk getting wrong.
func buildQuestions(films []oscarFilm) []questions.Question {
	var qs []questions.Question

	for i, f := range films {
		distractors := pickDistractors(films, i, f.Director, 3)
		if len(distractors) < 3 {
			continue // not enough distinct-director films to build 4 choices
		}
		correctPos := i % 4

		trChoices := make([]string, 4)
		enChoices := make([]string, 4)
		trChoices[correctPos] = f.Director
		enChoices[correctPos] = f.Director
		di := 0
		for slot := 0; slot < 4; slot++ {
			if slot == correctPos {
				continue
			}
			d := films[distractors[di]].Director
			trChoices[slot] = d
			enChoices[slot] = d
			di++
		}

		qs = append(qs, questions.Question{
			ID:        "oscar-director-" + lastPathSegment(f.WikidataID),
			Category:  "movies",
			Correct:   correctPos,
			SourceURL: "https://www.wikidata.org/wiki/" + lastPathSegment(f.WikidataID),
			TR: questions.Localized{
				Text:    fmt.Sprintf("\"%s\" filminin yönetmeni kimdir?", f.TitleTR),
				Choices: trChoices,
			},
			EN: questions.Localized{
				Text:    fmt.Sprintf("Who directed \"%s\"?", f.TitleEN),
				Choices: enChoices,
			},
		})
	}

	return qs
}

// pickDistractors walks forward from i (wrapping around) collecting up to
// count film indices whose director differs from excludeDirector and from
// every distractor already chosen.
func pickDistractors(films []oscarFilm, i int, excludeDirector string, count int) []int {
	n := len(films)
	seen := map[string]bool{excludeDirector: true}
	var picked []int
	for step := 1; step < n && len(picked) < count; step++ {
		idx := (i + step) % n
		d := films[idx].Director
		if seen[d] {
			continue
		}
		seen[d] = true
		picked = append(picked, idx)
	}
	return picked
}

func lastPathSegment(uri string) string {
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			return uri[i+1:]
		}
	}
	return uri
}
