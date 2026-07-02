# Soundcheck domain icons (EL-073)

Branded, domain-specific glyphs that off-the-shelf icon packs don't carry.
Drawn to match the EL-071 v1 mark (equalizer / waveform bar language) and the
EL-072 design tokens.

**v1 — consistent with the EL-071 v1 mark (mark being refined); not yet wired
into components.** These are asset files only. The per-surface reskin tickets
wire them in later (EL-078 owns the shared components).

## Style

- 24px `viewBox`, medium stroke (~1.75px; equalizer bars use 3px), rounded
  joins/caps — friendly, not surgical. Legible at 16px and 24px.
- Line variant by default; **filled = active / selected / status.**
- Single-color and `currentColor`-friendly so icons inherit the token color in
  use — **except** the two status-semantic glyphs, whose *filled* variant bakes
  the status color because color carries meaning there (pair with a text label;
  status is never color-alone).

## Icons

| File | Variant | Meaning / use | Color |
| --- | --- | --- | --- |
| `waveform-check-line.svg` | line | cleared / certified | `currentColor` |
| `waveform-check-filled.svg` | filled (status) | cleared / certified | Cleared `#34D399` |
| `half-meter-line.svg` | line | pending / in progress | `currentColor` |
| `half-meter-filled.svg` | filled (status) | pending / in progress | Pending `#F5B544` |
| `fader-line.svg` | line | skills / levels (mixer channel faders) | `currentColor` |
| `fader-filled.svg` | filled | skills / levels (selected) | `currentColor` |
| `booth-line.svg` | line | a gig / the venue (DJ booth, top-down) | `currentColor` |
| `booth-filled.svg` | filled | a gig / the venue (selected) | `currentColor` |
| `certificate-line.svg` | line | the "certified" credential (seal + ribbon) | `currentColor` |
| `certificate-filled.svg` | filled | the "certified" credential (selected) | `currentColor` |

### Notes on individual glyphs

- **waveform-check** — equalizer bars whose tops trace a checkmark (short dip,
  long climb), the EL-071 mark concept. The check reading is strongest at 24px+;
  at 16px it reads mainly as rising level bars (the filled/status color carries
  the "cleared" meaning). The filled variant reuses the `status-cleared` geometry.
- **half-meter** — a level capsule filled to the half mark: the climb hasn't
  finished. The filled variant is Pending amber.
- **fader** — two mixer channel faders at different levels.
- **booth / deck** — top-down console with two jog wheels and a center mixer
  strip. The filled variant knocks the jog wheels out with `fill-rule="evenodd"`.
- **certificate** — a seal with ribbon tails and an inner check. The filled
  variant knocks the check out of the seal with `fill-rule="evenodd"`.

## Usage sketch (for the reskin tickets)

Inline the SVG (or load via `<img>`) and set `color` to a token to tint the
`currentColor` variants:

```html
<span class="text-[color:var(--sc-cue-amber)]">
  <!-- inline fader-line.svg here -->
</span>
```

The status-semantic filled glyphs already carry their color; still render an
adjacent text label.
