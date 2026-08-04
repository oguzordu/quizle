import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { LocaleProvider } from "@/lib/i18n";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const siteUrl = "https://quizle.example";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: "Quizle — Çok Oyunculu Bilgi Yarışması",
  description:
    "Arkadaşlarınla oda kodu ile katıldığın, gerçek zamanlı çok oyunculu bilgi yarışması. Multiplayer trivia game you can play with friends in real time.",
  keywords: [
    "bilgi yarışması",
    "çok oyunculu quiz",
    "multiplayer trivia",
    "quiz game",
    "arkadaşlarla oyun",
  ],
  openGraph: {
    title: "Quizle — Çok Oyunculu Bilgi Yarışması",
    description: "Arkadaşlarınla oda kodu ile katıldığın, gerçek zamanlı bilgi yarışması.",
    type: "website",
    locale: "tr_TR",
    alternateLocale: "en_US",
  },
  twitter: {
    card: "summary",
    title: "Quizle — Çok Oyunculu Bilgi Yarışması",
    description: "Arkadaşlarınla oda kodu ile katıldığın, gerçek zamanlı bilgi yarışması.",
  },
};

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "WebApplication",
  name: "Quizle",
  applicationCategory: "GameApplication",
  description: "Arkadaşlarınla oda kodu ile oynanan, gerçek zamanlı çok oyunculu bilgi yarışması.",
  inLanguage: ["tr", "en"],
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="tr"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
        <LocaleProvider>{children}</LocaleProvider>
      </body>
    </html>
  );
}
