# Chart typeface

The unmodified Latin variable-weight WOFF2 is from
[`@fontsource-variable/inter@5.3.0`](https://www.npmjs.com/package/@fontsource-variable/inter/v/5.3.0),
upstream file `files/inter-latin-wght-normal.woff2`. [Inter](https://rsms.me/inter/)
is by the Inter Project Authors; its SIL Open Font License is retained in
`LICENSE` and included with generated reports.

`fonts.go` embeds the font in the HTML and exported chart images. Reports wait
for it before measuring chart text. No installed font or network is required.
Other writing systems use the existing system fallback.

To update, copy the same font and license from a pinned release and update the
version here and in `chartFontCSS`. Do not edit the vendored font binary.
