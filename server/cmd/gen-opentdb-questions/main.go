// Command gen-opentdb-questions imports the Open Trivia Database into our
// own question-pack format, the same way cmd/gen-wikidata-questions imports
// Oscar facts from Wikidata. OpenTDB is CC BY-SA 4.0 and needs no API key,
// but it is English-only — the questions it produces carry an "en" block and
// no "tr" block, so they are served only to English rooms until someone
// writes a Turkish translation for them.
//
// The API hands out at most 50 questions per call and rate-limits each IP to
// one call every 5 seconds, so a full import takes roughly ten minutes. A
// session token is used so the API never returns the same question twice.
//
// Run it to refresh the pack:
//
//	go run ./cmd/gen-opentdb-questions
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/oguzordu/quizle/internal/questions"
)

// rateLimitDelay is OpenTDB's documented per-IP limit (one call per 5s),
// plus a little headroom so a slow clock never trips it.
const rateLimitDelay = 5500 * time.Millisecond

const outputPath = "data/questions/opentdb.json"

// attributionURL credits OpenTDB as the source, as CC BY-SA 4.0 requires.
const attributionURL = "https://opentdb.com"

// categorySlugs maps OpenTDB's category IDs onto the slugs Quizle already
// uses, so imported questions land alongside the hand-written ones instead
// of forming a parallel taxonomy. Categories absent here are skipped.
var categorySlugs = map[int]string{
	9:  "general",
	10: "books",
	11: "movies",
	12: "music",
	13: "theatre",
	14: "tv",
	15: "videogames",
	16: "boardgames",
	17: "science",
	18: "computers",
	19: "math",
	20: "mythology",
	21: "sports",
	22: "geography",
	23: "history",
	24: "politics",
	25: "arts",
	26: "celebrities",
	27: "animals",
	28: "vehicles",
	29: "comics",
	30: "gadgets",
	31: "anime",
	32: "cartoons",
}

type apiResponse struct {
	ResponseCode int `json:"response_code"`
	Results      []struct {
		Question         string   `json:"question"`
		CorrectAnswer    string   `json:"correct_answer"`
		IncorrectAnswers []string `json:"incorrect_answers"`
	} `json:"results"`
}

type tokenResponse struct {
	ResponseCode int    `json:"response_code"`
	Token        string `json:"token"`
}

func main() {
	token, err := requestToken()
	if err != nil {
		log.Fatalf("oturum anahtarı alınamadı: %v", err)
	}

	// Re-runs top the pack up rather than replacing it: OpenTDB hands out
	// questions in a different order each time, so starting from scratch
	// would churn the file and could drop questions a previous run caught.
	all := loadExisting()
	seen := make(map[string]bool, len(all))
	for _, q := range all {
		seen[q.ID] = true
	}
	fmt.Printf("mevcut pakette %d soru var\n\n", len(all))

	ids := make([]int, 0, len(categorySlugs))
	for id := range categorySlugs {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		slug := categorySlugs[id]
		got := fetchCategory(id, slug, token, seen)
		all = append(all, got...)
		fmt.Printf("%-12s (%2d): %d soru\n", slug, id, len(got))
	}

	// Sort by ID so re-running the importer produces a reviewable diff
	// rather than reshuffling the whole file.
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	if errs := questions.ValidatePack(all); len(errs) != 0 {
		log.Fatalf("üretilen sorular geçersiz: %v", errs[:min(len(errs), 10)])
	}

	b, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outputPath, append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\ntoplam %d soru %s dosyasına yazıldı\n", len(all), outputPath)
}

// OpenTDB response codes.
const (
	codeSuccess    = 0
	codeNoResults  = 1
	codeTokenEmpty = 4
	codeRateLimit  = 5
)

// fetchCategory drains one category, 50 questions at a time, until the API
// reports it has no more to give (response code 4 with a session token).
// batchSizes is tried in order. OpenTDB answers "no results" (code 1) when
// a category holds fewer questions than the amount asked for, rather than
// returning what it has — so a small category looks empty unless we step the
// batch size down. Each step also drains the tail of a large category.
var batchSizes = []int{50, 10, 3, 1}

func fetchCategory(categoryID int, slug, token string, seen map[string]bool) []questions.Question {
	var out []questions.Question
	rateLimitRetries := 0
	sizeIdx := 0

	for {
		time.Sleep(rateLimitDelay)

		url := fmt.Sprintf(
			"https://opentdb.com/api.php?amount=%d&category=%d&type=multiple&encode=base64&token=%s",
			batchSizes[sizeIdx], categoryID, token,
		)
		resp, err := fetchAPI(url)
		if err != nil {
			log.Printf("%s: istek başarısız (%v), bu kategori atlanıyor", slug, err)
			return out
		}

		// Being told to slow down is not "no more data" — back off and try
		// the same page again, or a whole category silently comes back empty.
		if resp.ResponseCode == codeRateLimit {
			if rateLimitRetries >= 5 {
				log.Printf("%s: hız sınırı aşılamadı, kategori atlanıyor", slug)
				return out
			}
			rateLimitRetries++
			time.Sleep(time.Duration(rateLimitRetries) * 2 * time.Second)
			continue
		}
		rateLimitRetries = 0

		// Both "no results" (1) and "token empty" (4) really mean "I can't
		// serve a batch that big" — a category holding 30 questions answers
		// 4 to a request for 50, so treating either as "category finished"
		// silently loses every category smaller than the batch size. Step
		// down instead, and only stop once even one question can't be had.
		if resp.ResponseCode == codeNoResults || resp.ResponseCode == codeTokenEmpty {
			sizeIdx++
			if sizeIdx >= len(batchSizes) {
				return out
			}
			continue
		}
		if resp.ResponseCode != codeSuccess {
			log.Printf("%s: beklenmeyen yanıt kodu %d, kategori atlanıyor", slug, resp.ResponseCode)
			return out
		}

		for _, r := range resp.Results {
			q, ok := buildQuestion(r.Question, r.CorrectAnswer, r.IncorrectAnswers, slug)
			if !ok || seen[q.ID] {
				continue
			}
			seen[q.ID] = true
			out = append(out, q)
		}
	}
}

// buildQuestion decodes one API result into our format, returning false for
// anything malformed (wrong answer count, empty or duplicate choices) rather
// than letting a broken question reach players.
func buildQuestion(encQuestion, encCorrect string, encIncorrect []string, slug string) (questions.Question, bool) {
	text, err := decode(encQuestion)
	if err != nil || text == "" {
		return questions.Question{}, false
	}
	correct, err := decode(encCorrect)
	if err != nil || correct == "" {
		return questions.Question{}, false
	}
	if len(encIncorrect) != 3 {
		return questions.Question{}, false
	}

	wrong := make([]string, 0, 3)
	for _, e := range encIncorrect {
		d, err := decode(e)
		if err != nil || d == "" {
			return questions.Question{}, false
		}
		wrong = append(wrong, d)
	}

	// Derive the answer's slot from the question text so a re-import puts
	// it in the same place — a stable file beats a freshly shuffled one.
	sum := sha256.Sum256([]byte(text))
	correctPos := int(sum[0]) % 4

	choices := make([]string, 4)
	choices[correctPos] = correct
	wi := 0
	for slot := 0; slot < 4; slot++ {
		if slot == correctPos {
			continue
		}
		choices[slot] = wrong[wi]
		wi++
	}

	if hasDuplicates(choices) {
		return questions.Question{}, false
	}

	return questions.Question{
		ID:        fmt.Sprintf("otdb-%x", sum[:6]),
		Category:  slug,
		Correct:   correctPos,
		SourceURL: attributionURL,
		EN:        questions.Localized{Text: text, Choices: choices},
	}, true
}

func hasDuplicates(choices []string) bool {
	seen := make(map[string]bool, len(choices))
	for _, c := range choices {
		if seen[c] {
			return true
		}
		seen[c] = true
	}
	return false
}

func decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// loadExisting reads the pack written by a previous run, returning an empty
// slice when there isn't one yet.
func loadExisting() []questions.Question {
	b, err := os.ReadFile(outputPath)
	if err != nil {
		return nil
	}
	var qs []questions.Question
	if err := json.Unmarshal(b, &qs); err != nil {
		log.Printf("mevcut paket okunamadı (%v), sıfırdan başlanıyor", err)
		return nil
	}
	return qs
}

func requestToken() (string, error) {
	body, err := get("https://opentdb.com/api_token.php?command=request")
	if err != nil {
		return "", err
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", err
	}
	if tr.ResponseCode != 0 || tr.Token == "" {
		return "", fmt.Errorf("beklenmeyen yanıt kodu %d", tr.ResponseCode)
	}
	return tr.Token, nil
}

func fetchAPI(url string) (*apiResponse, error) {
	body, err := get(url)
	if err != nil {
		return nil, err
	}
	var r apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
