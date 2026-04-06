import type { Metadata } from 'next'
import { Plus_Jakarta_Sans, IBM_Plex_Mono } from 'next/font/google'
import './globals.css'

const plusJakarta = Plus_Jakarta_Sans({
  subsets: ['latin'],
  variable: '--font-sans',
  weight: ['400', '500', '600'],
})

const ibmPlexMono = IBM_Plex_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  weight: ['400', '500'],
})

export const metadata: Metadata = {
  title: 'Property Inspector',
  description: 'AI-powered property inspection tool',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${plusJakarta.variable} ${ibmPlexMono.variable}`}>
      <body className="font-sans">{children}</body>
    </html>
  )
}
