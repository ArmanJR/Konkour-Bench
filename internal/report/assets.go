package report

import _ "embed"

//go:embed page.html
var page string

//go:embed report.js
var script string

//go:embed charts.js
var chartScript string

//go:embed report.css
var stylesheet string

// These assets are generated from the dashboard's pinned KaTeX dependency by
// make report-assets. Fonts are data URLs so a shared HTML file works offline.
//
//go:embed math.generated.js
var mathScript string

//go:embed math.generated.css
var mathStylesheet string
