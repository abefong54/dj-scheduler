module.exports = {
  content: ["./src/**/*.{html,ts}"],
  theme: {
    extend: {
      fontFamily: {
        // EL-072: 'Noto Sans TC' added to the sans + display stacks so
        // Traditional Chinese (Taiwan) has coverage under the Soundcheck type
        // system. Additive — existing families and order are unchanged.
        sans: ['Inter', 'Noto Sans TC', 'system-ui', 'sans-serif'],
        display: ['Space Grotesk', 'Noto Sans TC', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'SFMono-Regular', 'monospace'],
      },
    },
  },
  plugins: [],
}
