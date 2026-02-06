package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *App) buildStatusBar() fyne.CanvasObject {
	a.statusLabel = widget.NewLabel("")
	a.updatePrinterStatus()

	reconnectBtn := widget.NewButton("Reconectar", func() {
		a.reconnectPrinter()
	})

	bar := container.New(
		layout.NewHBoxLayout(),
		a.statusLabel,
		layout.NewSpacer(),
		reconnectBtn,
	)
	return bar
}

func (a *App) updatePrinterStatus() {
	if a.printer != nil && a.printer.IsConnected() {
		a.statusLabel.SetText("Impressora: Conectada (" + a.printer.Path() + ")")
	} else {
		a.statusLabel.SetText("Impressora: Desconectada")
	}
}
