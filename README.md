# Quizle

Çok oyunculu, gerçek zamanlı bilgi yarışması. Oda kodu ile arkadaşlarınla katıl, 4 şıklı sorulara hızlı ve doğru cevap ver, puan topla.

Türkçe ve İngilizce arayüz. 2–20 oyuncu. Kayıt gerektirmez.

**Canlı:** https://web-ruddy-ten-69.vercel.app (backend: https://quizle-api.fly.dev)

## Neden bu proje?

Piyasadaki çok oyunculu bilgi yarışması sitelerinin çoğu ya kalitesiz sorular içeriyor ya da bağlantı koptuğunda oyunu öldürüyor. Quizle iki amaçla yazıldı:

1. **Doğru bir ürün:** Her sorunun bir kaynağı var, sunucu doğru cevabı zamanından önce asla göndermiyor, bağlantı kopunca oyuncu kaldığı yerden devam edebiliyor.
2. **Gerçek bir mühendislik problemi:** Oda başına durum makinesi, sunucu-otoriteli senkron zamanlayıcı, eşzamanlılık — Firebase/Supabase gibi bir BaaS'a yaptırmak yerine elle çözüldü.

## Mimari

```
server/   Go — WebSocket hub + saf oyun motoru (internal/game), tamamen bellek içi (v1)
web/      Next.js 15 (App Router) + TypeScript + Tailwind
```

Soru bankası JSON dosyaları olarak tutuluyor (`server/data/questions`); PostgreSQL/Redis henüz eklenmedi — v1 kapsamında gerek duyulmadı (bkz. design doc'taki YAGNI notu).

`internal/game` paketi ağdan, veritabanından ve Redis'ten tamamen izole saf bir Go paketidir: girdi komut + zaman damgası, çıktı yeni durum + yayılacak event listesi. Bu sayede tüm oyun mantığı sahte saatle deterministik olarak test edilir.

Tasarım kararlarının tam gerekçesi için: [`docs/superpowers/specs/2026-08-04-quizle-design.md`](docs/superpowers/specs/2026-08-04-quizle-design.md)

## Geliştirme

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

## Yayına alma

- Backend: Fly.io (`cd server && flyctl deploy`)
- Frontend: Vercel (`cd web && vercel --prod`), `NEXT_PUBLIC_API_BASE` ortam değişkeni backend URL'ini gösterir

## Durum

Aktif geliştirme aşamasında (v1). Yol haritası için design doc'a bakınız.
