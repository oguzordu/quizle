# Ana Sayfa Cilası + Soru Çeşitliliği — Tasarım

**Tarih:** 2026-08-05
**Durum:** Onaylandı, plana geçiliyor

## Bağlam

Vercel deploy sorunu çözüldü (Root Directory ayarı `web` olarak düzeltildi). Bu spec, sıradaki iki bağımsız iyileştirmeyi kapsıyor:

1. Ana sayfanın (lobi/giriş ekranı) "arkadaşlarla eğleneceğiz" hissini güçlendirmesi — özellikle avatar seçici zayıf ve genel enerji eksik bulundu.
2. Soru havuzundaki tekdüzelik — 265 sorunun 89'u (%33) Wikidata'dan otomatik üretilmiş "bu filmi kim yönetti" kalıbında, bu yüzden bir oyun turunda hep aynı tarz soru geliyormuş hissi oluşuyor.

İki iş birbirinden bağımsız, farklı katmanlarda (frontend UI vs. backend soru havuzu + veri), ama tek oturumda birlikte ele alınacak küçük ölçekli işler.

## 1) Avatar Seçici (`web/components/AvatarPicker.tsx`)

**Şu anki durum:** Küçük emoji ızgarası (6x8 grid) ve ayrı bir renk noktaları satırı. Emoji ve renk görsel olarak birbirinden kopuk. Seçili öğe sade bir mor çerçeveyle belli oluyor, statik.

**Değişiklikler:**
- Üstteki `Avatar` önizlemesi büyütülecek (52px → ~72px) ve arka planına seçili rengi alacak (şu an `Avatar` bileşeninin zaten `avatar` prop'undan emoji+renk ayrıştırdığını varsayıyoruz — `components/Avatar.tsx` bu mantığı zaten içeriyor).
- Her emoji/renk seçiminde önizlemeye kısa bir CSS keyframe animasyonu uygulanacak (ör. `scale` tabanlı "pop": 1 → 1.15 → 1, ~250ms). React state ile `key` değiştirerek animasyonu yeniden tetikleme (mevcut `AnimatedScore.tsx` bileşeninde benzer bir teknik olabilir, oradaki deseni takip et).
- Emoji butonları büyütülecek ve seçili olan, seçili rengin soft (opacity düşük) arka planını alacak — emoji ve renk seçimi görsel olarak birleşecek.
- İsim inputunun yanındaki 🎲 rastgele-isim butonuna paralel, avatar önizlemesinin yanına bir 🎲 "rastgele avatar" butonu eklenecek. Mevcut `randomAvatar()` fonksiyonu (`lib/quips.ts`) zaten var, sadece UI'a bağlanacak.
- Renk noktalarına seçildiğinde hafif bir halka pulse efekti (mevcut `ring-2` yapısı animasyonlu hale getirilecek).

**Kapsam dışı:** Ses efekti (kullanıcı onayladı ama "yeterli" dedi, ekstra scope istemedi), yeni emoji/renk seçenekleri eklemek.

## 2) Ana Sayfa Genel Enerjisi (`web/app/page.tsx`)

**Değişiklikler:**
- 🏆 başlık ikonuna sürekli, hafif bir "wobble" animasyonu (küçük dönüş + scale, birkaç saniyede bir, rahatsız etmeyecek genlikte).
- Tagline metni daha davetkâr bir tona çevrilecek:
  - TR: "Arkadaşlarınla çok oyunculu bilgi yarışması" → **"Arkadaşlarını topla, kim daha hızlı cevaplayacak?"**
  - EN: "Multiplayer trivia with your friends" → **"Gather your friends — who'll answer first?"**
  - (`web/lib/i18n.tsx` içindeki `tagline` anahtarları güncellenecek)
- Ana kart (`bg-white/95 ... rounded-3xl`) sayfa yüklendiğinde hafif aşağıdan yukarı kayarak + fade-in ile belirecek (CSS animation, mount'ta tetiklenir).
- "Oda Kur" ve "Katıl" butonlarına hover'da hafif bir wiggle/scale efekti (mevcut `active:scale-[0.97]` yapısına ek olarak `hover:` durumu).
- `FLOATERS` dizisindeki emojilere farklı animasyon süresi/genliği verilecek (şu an hepsi `3.5s ease-in-out` — bazıları 2.8s, bazıları 4.2s gibi çeşitlendirilecek ki arka plan daha organik görünsün). Mevcut `delay` alanları korunacak, `duration` alanı eklenecek.

**Kapsam dışı:** Sahte "şu an X kişi oynuyor" gibi backend'de karşılığı olmayan veri; illüstrasyon/görsel asset eklemek.

## 3) Soru Havuzu Genişletme + Kategori Dengesi

**Veri (`server/data/questions/general-knowledge.json`, `general-knowledge-2.json`):**
- ~60 yeni, elle yazılmış soru eklenecek. Mevcut kategoriler: `science`, `history`, `sports`, `arts`, `geography`, `movies`. Bunlara ek olarak `music` ve `animals` kategorileri eklenecek.
- Her soru mevcut formatı takip edecek: `id`, `category`, `correct` (index), `source_url` (gerçek, doğrulanabilir kaynak), `tr`/`en` içinde `text` + 4 `choices`.
- `movies-oscars.json`'a dokunulmayacak (89 soru olduğu gibi kalacak) — asıl sorun onun havuzda ezici oranda seçilmesiydi, bu 3. maddedeki dengeleme mantığıyla çözülecek.

**Kategori dengesi (`server/internal/hub/server.go` → `pickQuestions`):**
- Şu an: pool'u Fisher-Yates ile karıştırıp ilk `sampleSize` (10) taneyi alıyor — kategori dağılımına bakmıyor.
- Yeni mantık: karıştırılmış pool'dan sırayla seçim yaparken, bir kategoriden zaten `maxPerCategory` (4) tane seçilmişse o kategoriyi bu round için atla, sıradaki uygun soruya geç. 10 soru dolana kadar devam et. Havuzda yeterince çeşitlilik olduğu sürece (ki artık ~325 soru / 8 kategori var) bu limit rahatça karşılanabilir; teorik olarak havuz çok dar kalırsa (ör. tek kategori kalırsa) limit gevşetilip dolum tamamlanır (sonsuz döngüye girmeden, best-effort).
- Test (TDD, mevcut `server_test.go` konumuna paralel bir dosyada): pool'dan çok sayıda round simüle edip (ör. 500 iterasyon) hiçbirinde tek kategoriden `maxPerCategory`'den fazla soru gelmediğini doğrulayan bir test.

## Etkilenen Dosyalar

- `web/components/AvatarPicker.tsx`
- `web/components/Avatar.tsx` (muhtemelen, önizleme boyutu için prop kontrolü)
- `web/app/page.tsx`
- `web/app/globals.css` (yeni keyframe'ler için)
- `web/lib/i18n.tsx` (tagline metni)
- `server/data/questions/general-knowledge.json`
- `server/data/questions/general-knowledge-2.json`
- `server/internal/hub/server.go` (`pickQuestions`)
- `server/internal/hub/server_test.go` (yeni test)

## Test Stratejisi

- Backend: `pickQuestions` kategori dengesi için TDD — önce başarısız test, sonra implementasyon. `go test -race ./...` yeşil kalmalı.
- Frontend: Görsel/animasyon değişiklikleri olduğu için otomatik test yazılmayacak (mevcut proje deseni de böyle); tarayıcıda manuel doğrulama yapılacak (dev server üzerinden).
- Soru verisi: mevcut `validate.go` / soru doğrulama testleri (varsa) yeni sorular için de geçmeli — `go test ./...` ile kontrol edilecek.
