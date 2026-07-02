# Soundcheck brand assets

**v1 draft — pending founder pass, per EL-071.** These are working SVGs for the
app reskin (EL-078) and drafts for the founder's review. A logo is ultimately the
founder's call; treat geometry, kerning, and the wordmark outline as not-yet-final.

## Direction

- **Mark — "waveform-check":** a checkmark whose bar tops trace the tick (short
  arm down to a low vertex, long arm climbing to the tallest bar) while the bars
  simultaneously read as an audio equalizer — "check" and "sound" at once.
- **Wordmark — "soundcheck":** lowercase Space Grotesk (600). The counter of the
  "o" carries a single lit VU-meter segment in Cue Amber.
- **Lockup:** mark + wordmark.

## Files

| File | What it is | Use |
|------|-----------|-----|
| `lockup-on-dark.svg` | Primary lockup, Booth Black card + light text | Default (dark "booth") context |
| `mark.svg` | Standalone mark (Cue Amber→Deck Cyan gradient) | App icon, cert badge, UI glyph |
| `wordmark.svg` | Standalone wordmark + amber VU "o" detail | Text-only lockups |
| `one-color.svg` | Flat single-color lockup, no gradient | Silkscreen / sticker / hoodie |
| `favicon.svg` | Crisp full mark on Booth tile | Favicon ≥ 32px |
| `favicon-16.svg` | Low-detail 4-bar variant | 16px — stays a waveform-check, not a plain tick |
| `app-icon-maskable.svg` | Full-bleed Booth tile, mark in the maskable safe zone | PWA/app icon (`purpose: maskable`) — EL-077 |
| `status-cleared.svg` | Green full waveform-check | "Cleared / Certified" status pill |
| `status-pending.svg` | Amber half-meter | "Pending" status pill |
| `status-blocked.svg` | Clip Red + clip-ceiling bar | "Blocked / Error" status pill |

Status glyphs are sized for 16–24px and, per the design system, are always paired
with a text label (status is never color-alone).

## Palette

Cue Amber `#FF9E2C` · Deck Cyan `#22D3EE` · Booth Black `#0B0D10` · Stage White
`#F7F8FA` · text neutral-100 `#E6EAF0` · Cleared `#34D399` · Pending `#F5B544` ·
Clip Red `#F4505B`. Matches the EL-072 `--sc-*` tokens in `src/styles.css`.

## Outstanding (needs founder pass)

- **Wordmark** uses live Space Grotesk `<text>` (kept editable for the draft). For
  final delivery convert to outlines + a manual kern pass; the amber "o" segment
  position is tuned to Space Grotesk metrics and should be re-checked when outlined.
- **PNG export pending** — no rasterizer was available in the build env. Export
  `favicon.svg` (and `favicon-16.svg` at 16px) to PNG at 16/32/64px for `.ico`.
