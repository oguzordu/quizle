# Kategori Seçimi + OpenTDB Soru Havuzu — Tasarım

**Tarih:** 2026-08-05
**Durum:** Onaylandı

## Amaç

İki bağlı hedef:

1. Oyuncu oda kurarken bir **kategori** seçsin (genel kültür, müzik, coğrafya, tarih, oyun, dizi, film, ünlüler, ...). Odaya katılanlar o kategoride oynar.
2. Soru havuzu hem **çeşitlensin** hem **büyüsün** — şu an 325 soru var ve %29'u film kategorisinde.

## Soru kaynağı: Open Trivia Database

[OpenTDB](https://opentdb.com) seçildi:

- **Lisans: CC BY-SA 4.0** — ticari kullanıma izin verir, atıf + aynı lisansla paylaşım ister. (Alternatif *The Trivia API* 14.400 soruyla daha büyük ama **CC BY-NC** — ticari kullanımı yasaklıyor, ileride eli bağlar.)
- API anahtarı gerektirmez, ~5.300 soru, 24 kategori, hepsi 4 şıklı.
- İstek başına 50 soru, IP başına 5 saniyede 1 istek. Session token ile tekrar dönmez → tüm veritabanı ~9 dakikada çekilebilir.

**Önemli:** OpenTDB yalnızca İngilizce. Türkçe için hazır bir "eğlenceli bilgi yarışması" veri seti aranıp bulunamadı (mevcut Türkçe setler ya okuduğunu-anlama, ya KPSS/TUS sınav soruları ve CC BY-NC lisanslı). Türkçe içerik elle üretilecek.

Mevcut `gen-wikidata-questions` deseni takip edilerek `gen-opentdb-questions` komutu yazılır: çalışma anında dış API'ye bağımlılık yok, sorular tek seferlik JSON'a dökülür.

## Mimari değişiklikler

### 1. Soru şemasında dil zorunluluğunun gevşetilmesi

Bugün `questions.ValidatePack` hem `tr` hem `en` alanını zorunlu tutuyor. OpenTDB soruları yalnızca İngilizce olacağı için:

- `en` zorunlu kalır.
- `tr` **isteğe bağlı** olur. Yoksa soru sadece İngilizce havuzda yer alır.
- Bir dil bloğu varsa, o blok kendi içinde tam doğrulanmaya devam eder (4 şık, boş/yinelenen şık yok, `correct` indeksi geçerli).

### 2. Sunucuda dil farkındalığı

Bugünkü `cmd/quizle/main.go` soruları yüklerken yalnızca `q.TR` okuyor — arayüz İngilizce'ye alınsa bile sorular Türkçe geliyor. Bu düzeltilir:

- `loadQuestionPool` her paketi **iki havuza** ayırır: `tr` ve `en`. Bir soru, o dilde metni varsa ilgili havuza girer.
- `hub.Server` tek `questionPool` yerine `pools map[string][]game.Question` tutar.
- `game.Question` dil bilmez (saf oyun motoru olarak kalır) — dil seçimi havuz seviyesinde çözülür.

### 3. Kategori seçimi

- `POST /rooms` iki opsiyonel sorgu parametresi alır: `lang` (varsayılan `tr`) ve `category`.
- `category` verilmişse havuz o kategoriye filtrelenir; verilmemişse (veya `mixed` ise) bugünkü kategori-dengeli rastgele seçim korunur.
- Bilinmeyen/boş kategori veya o dilde yetersiz soru varsa `400` döner; istemci kullanıcıya anlamlı bir hata gösterir.
- `GET /categories?lang=tr` yeni bir uç nokta: o dilde **en az bir oyunluk (10) sorusu olan** kategorileri döner. Böylece istemci, içeriği olmayan bir kategoriyi hiç göstermez — Türkçe ve İngilizce farklı listeler görebilir ve liste içerik büyüdükçe kendiliğinden genişler.

### 4. Kategori kimlikleri

Mevcut slug'lar korunur (`movies`, `capitals`, `flags`, `science`, `history`, `arts`, `sports`, `geography`, `music`, `animals`). OpenTDB kategorileri bunlara eşlenir, karşılığı olmayanlar yeni slug alır:

| OpenTDB | slug |
|---|---|
| General Knowledge | `general` |
| Entertainment: Film | `movies` |
| Entertainment: Music | `music` |
| Entertainment: Television | `tv` |
| Entertainment: Video Games | `videogames` |
| Celebrities | `celebrities` |
| Geography | `geography` |
| History | `history` |
| Sports | `sports` |
| Science & Nature | `science` |
| Animals | `animals` |
| Art | `arts` |

Diğer OpenTDB kategorileri (Mitoloji, Anime, Çizgi Roman, Kitaplar, Araçlar, Bilgisayar, Matematik, ...) da kendi slug'larıyla alınır — İngilizce oyunculara zengin bir liste sunar.

### 5. Türkçe içerik

Kullanıcının saydığı 8 kategori için (`general`, `music`, `geography`, `history`, `videogames`, `tv`, `movies`, `celebrities`) kategori başına **~40 Türkçe soru** hedeflenir. Kaynak OpenTDB'den seçilip elle çevrilir/uyarlanır; çeviriyle anlamını yitiren veya Türkiye'den oynayan biri için anlamsız olan sorular (ör. yalnızca ABD'ye özgü popüler kültür) elenir, yerlerine uygun olanlar seçilir.

Mevcut Türkçe sorular korunur; `flags` ve `capitals` Quizle'a özgü kategoriler olarak kalır.

### 6. Frontend

- Ana sayfada, "Yeni Oda Oluştur" butonunun üstünde bir **kategori seçici** yer alır. Seçenekler `GET /categories` çıktısından gelir, arayüz diline göre filtrelenir.
- Her kategorinin bir emoji + yerelleştirilmiş adı olur (`lib/i18n.tsx`).
- "Karışık" seçeneği listenin başında durur ve varsayılandır — bugünkü davranış.
- Seçilen kategori `POST /rooms`'a iletilir; oda ekranında hangi kategoride oynandığı başlıkta görünür.

## Test stratejisi

TDD, backend'de her davranış için önce başarısız test:

- `ValidatePack`: `tr` bloğu olmayan soru geçerli; `en` olmayan soru geçersiz; var olan blok yine tam doğrulanıyor.
- `loadQuestionPool`: yalnızca İngilizce bir soru `en` havuzuna girer, `tr` havuzuna girmez.
- `pickQuestions`: kategori verildiğinde yalnızca o kategoriden soru döner; verilmediğinde mevcut kategori-dengesi (kategori başına en fazla 4) korunur.
- `CreateRoomHandler`: `lang`/`category` parametrelerini uygular; bilinmeyen kategoride 400 döner.
- `CategoriesHandler`: yalnızca o dilde ≥10 sorusu olan kategorileri döner.

Frontend'de görsel/etkileşim değişiklikleri manuel olarak tarayıcıda doğrulanır (mevcut proje deseni).

`go test -race ./...`, `go vet`, `validate-questions` ve frontend `tsc`/`lint` yeşil kalmalı.

## Uygulama sırası

1. Şema: `tr` opsiyonel (validate + testler)
2. `gen-opentdb-questions` komutu, soruların çekilip JSON'a yazılması
3. Sunucuda dil havuzları (`loadQuestionPool`, `Server.pools`)
4. Kategori filtresi + `POST /rooms` parametreleri + `GET /categories`
5. Türkçe çeviri partisi (8 kategori × ~40 soru)
6. Frontend kategori seçici + i18n etiketleri
7. Deploy (Fly.io + Vercel), canlı doğrulama
