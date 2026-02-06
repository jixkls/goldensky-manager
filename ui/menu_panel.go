package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
)

func (a *App) buildMenuPanel() fyne.CanvasObject {
	a.menuTabs = container.NewAppTabs()
	a.refreshMenuTabs()

	scrollable := container.NewVScroll(a.menuTabs)
	scrollable.SetMinSize(fyne.NewSize(420, 0))

	return scrollable
}

func (a *App) refreshMenuTabs() {
	a.menuTabs.Items = nil

	categories := a.menu.Categories()
	if len(categories) == 0 {
		placeholder := widget.NewLabel("Cardapio vazio.\nUse Opcoes > Editar Cardapio")
		a.menuTabs.Append(container.NewTabItem("Cardapio", placeholder))
		a.menuTabs.Refresh()
		return
	}

	for _, category := range categories {
		items := a.menu.ItemsByCategory(category)
		grid := a.buildCategoryGrid(items)
		tab := container.NewTabItem(category, grid)
		a.menuTabs.Append(tab)
	}
	a.menuTabs.Refresh()
}

func (a *App) buildCategoryGrid(items []pos.MenuItem) fyne.CanvasObject {
	var buttons []fyne.CanvasObject
	for _, item := range items {
		item := item // capture loop variable
		label := item.Name + "\n" + pos.FormatBRL(item.Price)
		btn := widget.NewButton(label, func() {
			a.addItemToOrder(item)
		})
		buttons = append(buttons, btn)
	}

	if len(buttons) == 0 {
		return widget.NewLabel("Sem itens")
	}

	grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 70)), buttons...)
	return container.NewVScroll(grid)
}
