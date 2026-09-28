# Quizle

Real-time multiplayer trivia game. Friends join a room with a code, answer four-choice questions against the clock, and score points for speed and streaks. Turkish and English UI, 2 to 20 players, no sign-up.

**Live:** https://quizleapp.vercel.app (API: https://quizle-api.fly.dev)

*Türkçe açıklama aşağıda.*

## Engineering highlights

- **One goroutine per room (actor pattern).** Every state change for a room goes through a single goroutine, so the game state itself needs no locks.
- **Server-authoritative timing.** The server owns the countdown and never sends the correct answer before the reveal phase, so a client cannot cheat by reading network traffic.
- **Reconnects keep progress.** A reconnect token maps a returning WebSocket back to the same player, so a dropped connection costs neither score nor streak.
- **Pure, deterministic game engine.** `internal/game` knows nothing about networking or storage: the input is a command plus a timestamp, the output is the new state plus the events to broadcast. All game logic is tested with a fake clock.
- **Every question cites a source.** A validator runs in CI over the whole bank (4,933 questions, 567 of them in both languages) and fails the build on a missing source, a duplicate id, or malformed choices.
- **Abuse limits.** Room creation is rate-limited per IP, and WebSocket messages have a size cap.

## Architecture

```
server/   Go: WebSocket hub + pure game engine (internal/game), in-memory state (v1)
web/      Next.js 16 (App Router) + TypeScript + Tailwind CSS
```

The question bank is stored as JSON files (`server/data/questions`). PostgreSQL and Redis are not used yet; v1 did not need them.

## Development

```bash
# Backend
cd server
go test -race ./...
go run ./cmd/quizle

# Frontend
cd web
npm install
npm run dev
```

CI (GitHub Actions) runs `go vet`, a `gofmt` check, `go test -race ./...`, and the question-bank validator on every push.

## Deployment

- Backend: Fly.io (`cd server && flyctl deploy`)
- Frontend: Vercel (`cd web && vercel --prod`); the `NEXT_PUBLIC_API_BASE` environment variable points at the backend URL

---

## Türkçe

Çok oyunculu, gerçek zamanlı bilgi yarışması. Oda kodu ile arkadaşlarınla katıl, 4 şıklı sorulara hızlı ve doğru cevap ver, puan topla. Türkçe ve İngilizce arayüz, 2–20 oyuncu, kayıt gerektirmez.

**Canlı:** https://quizleapp.vercel.app (backend: https://quizle-api.fly.dev)

### Neden bu proje?

Piyasadaki çok oyunculu bilgi yarışması sitelerinin çoğu ya kalitesiz sorular içeriyor ya da bağlantı koptuğunda oyunu öldürüyor. Quizle iki amaçla yazıldı:

1. **Doğru bir ürün:** Her sorunun bir kaynağı var, sunucu doğru cevabı zamanından önce asla göndermiyor, bağlantı kopunca oyuncu kaldığı yerden devam edebiliyor.
2. **Gerçek bir mühendislik problemi:** Oda başına durum makinesi, sunucu-otoriteli senkron zamanlayıcı ve eşzamanlılık, Firebase/Supabase gibi bir BaaS'a yaptırmak yerine elle çözüldü.

### Mimari

`internal/game` paketi ağdan ve veritabanından tamamen izole, saf bir Go paketidir: girdi komut + zaman damgası, çıktı yeni durum + yayılacak event listesi. Bu sayede tüm oyun mantığı sahte saatle deterministik olarak test edilir. Tasarım kararlarının gerekçesi: [`docs/superpowers/specs/2026-08-04-quizle-design.md`](docs/superpowers/specs/2026-08-04-quizle-design.md)
