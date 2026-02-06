package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed goldensky_icon.png
var iconBytes []byte

var appIcon = fyne.NewStaticResource("goldensky_icon.png", iconBytes)
