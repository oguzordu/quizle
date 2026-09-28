# Quizle — Çok Oyunculu Bilgi Yarışması (Tasarım)

## Context

İnternetteki Kahoot tarzı çok oyunculu bilgi yarışması sitelerinin çoğu ya kalitesiz sorular içeriyor ya da bağlantı koptuğunda oyunu öldürüyor. Bu proje hem doğru bir ürün (kaynaklı sorular, kopunca ölmeyen oyun) hem de gerçek bir mühendislik problemi (Go ile oda başına state machine, sunucu-otoriteli senkron timer) hedefliyor.

## Kararlar

| Konu | Karar |
|---|---|
| Backend | Go + ham WebSocket + Redis + PostgreSQL |
| Frontend | Next.js 15 (App Router) + TypeScript + Tailwind |
| v1 kapsamı | Oda kodu ile davet, 2–20 kişi, küratörlü paketler, TR+EN, canlı skor |
| Profil | Fotoğraf yükleme, alternatif olarak DiceBear tarzı özelleştirilebilir avatar (göz/kaş/saç/renk parça parça seçilebilir, insansı veya şirin/emoji tarzı) |
| İsim | Quizle |
| Streak | ×1.5 tavanlı çarpan |

## Puanlama

```
raw = 1000 * (0.5 + 0.5 * kalanSüre/toplamSüre)
streakMult = 1.0 (0-2) | 1.2 (3-4) | 1.35 (5-6) | 1.5 (7+)
puan = round(raw * streakMult)
```

## Mimari

`internal/game`: saf oyun motoru, ağ/DB/Redis bilmez. Oda başına tek goroutine (actor pattern). Sunucu doğru cevabı Reveal fazına kadar göndermez. `playerToken` ile yeniden bağlanma.

Tam gerekçeler ve faz planı için: proje kökündeki plan dosyası (`~/.claude/plans/bir-oyun-yapmak-istiyorum-enchanted-kettle.md`) referans alınmıştır; bu dosya o planın repo'ya kopyasıdır.

## Fazlar

0) Repo iskeleti · 1) Saf oyun motoru (TDD) · 2) WebSocket hub · 3) Soru bankası · 4) Next.js istemci · 5) Profil/mizah · 6) i18n/SEO · 7) Yayın

## Doğrulama

`go test -race ./...`, soru doğrulama scripti, Playwright E2E (reconnect dahil), elle oynama testi.
