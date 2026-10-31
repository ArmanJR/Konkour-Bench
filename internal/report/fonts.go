package report

import (
	_ "embed"
	"encoding/base64"
)

//go:embed fonts/inter-latin-wght-normal.woff2
var interFont []byte

//go:embed fonts/LICENSE
var interLicense string

// The same embedded face is used by the HTML and each exported SVG image.
func chartFontCSS() string {
	return "/* Inter, @fontsource-variable/inter 5.3.0\n" + interLicense + "*/\n" +
		"@font-face{font-family:Inter;font-style:normal;font-weight:100 900;font-display:block;src:url(data:font/woff2;base64," +
		base64.StdEncoding.EncodeToString(interFont) + ") format('woff2');}"
}
