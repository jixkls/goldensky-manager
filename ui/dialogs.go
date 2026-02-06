package ui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
	"notinha/internal/storage"
)

func (a *App) showConfigDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(a.config.Restaurant.Name)

	addressEntry := widget.NewEntry()
	addressEntry.SetText(a.config.Restaurant.Address)

	phoneEntry := widget.NewEntry()
	phoneEntry.SetText(a.config.Restaurant.Phone)

	cnpjEntry := widget.NewEntry()
	cnpjEntry.SetText(a.config.Restaurant.CNPJ)

	footerEntry := widget.NewEntry()
	footerEntry.SetText(a.config.Restaurant.Footer)

	printerEntry := widget.NewEntry()
	printerEntry.SetText(a.config.Printer.DevicePath)

	charsEntry := widget.NewEntry()
	charsEntry.SetText(fmt.Sprintf("%d", a.config.Printer.CharsPerLine))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Nome", Widget: nameEntry},
			{Text: "Endereco", Widget: addressEntry},
			{Text: "Telefone", Widget: phoneEntry},
			{Text: "CNPJ", Widget: cnpjEntry},
			{Text: "Rodape", Widget: footerEntry},
			{Text: "Impressora", Widget: printerEntry},
			{Text: "Colunas", Widget: charsEntry},
		},
		OnSubmit: func() {},
	}

	d := dialog.NewCustomConfirm("Configuracoes", "Salvar", "Cancelar",
		container.NewVScroll(form), func(save bool) {
			if !save {
				return
			}
			a.config.Restaurant.Name = nameEntry.Text
			a.config.Restaurant.Address = addressEntry.Text
			a.config.Restaurant.Phone = phoneEntry.Text
			a.config.Restaurant.CNPJ = cnpjEntry.Text
			a.config.Restaurant.Footer = footerEntry.Text
			a.config.Printer.DevicePath = printerEntry.Text

			if chars, err := strconv.Atoi(charsEntry.Text); err == nil && chars > 0 {
				a.config.Printer.CharsPerLine = chars
			}

			if err := storage.SaveConfig(a.config); err != nil {
				log.Printf("Erro ao salvar config: %v", err)
				dialog.ShowError(fmt.Errorf("erro ao salvar: %w", err), a.mainWindow)
				return
			}

			a.reconnectPrinter()
		}, a.mainWindow)

	d.Resize(fyne.NewSize(500, 450))
	d.Show()
}

func (a *App) showMenuEditorDialog() {
	var itemList *widget.List
	var selectedIndex int = -1

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nome do item")

	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Preco (ex: 12,50)")

	categoryEntry := widget.NewEntry()
	categoryEntry.SetPlaceHolder("Categoria")

	activeItems := a.activeMenuItems()

	itemList = widget.NewList(
		func() int { return len(activeItems) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Item - R$ 0,00 (Categoria)")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(activeItems) {
				item := activeItems[id]
				obj.(*widget.Label).SetText(
					fmt.Sprintf("%s - %s (%s)", item.Name, pos.FormatBRL(item.Price), item.Category),
				)
			}
		},
	)

	itemList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		if id < len(activeItems) {
			item := activeItems[id]
			nameEntry.SetText(item.Name)
			priceEntry.SetText(formatPriceForEdit(item.Price))
			categoryEntry.SetText(item.Category)
		}
	}

	addBtn := widget.NewButton("Adicionar", func() {
		price := parsePrice(priceEntry.Text)
		if nameEntry.Text == "" || categoryEntry.Text == "" {
			dialog.ShowInformation("Aviso", "Preencha nome e categoria.", a.mainWindow)
			return
		}
		a.menu.AddItem(pos.MenuItem{
			Name:     nameEntry.Text,
			Price:    price,
			Category: categoryEntry.Text,
		})
		a.saveMenuAndRefresh(&activeItems, itemList, nameEntry, priceEntry, categoryEntry)
	})

	updateBtn := widget.NewButton("Atualizar", func() {
		if selectedIndex < 0 || selectedIndex >= len(activeItems) {
			return
		}
		item := activeItems[selectedIndex]
		item.Name = nameEntry.Text
		item.Price = parsePrice(priceEntry.Text)
		item.Category = categoryEntry.Text
		a.menu.UpdateItem(item)
		a.saveMenuAndRefresh(&activeItems, itemList, nameEntry, priceEntry, categoryEntry)
	})

	removeBtn := widget.NewButton("Remover", func() {
		if selectedIndex < 0 || selectedIndex >= len(activeItems) {
			return
		}
		a.menu.RemoveItem(activeItems[selectedIndex].ID)
		selectedIndex = -1
		a.saveMenuAndRefresh(&activeItems, itemList, nameEntry, priceEntry, categoryEntry)
	})

	formPanel := container.NewVBox(
		widget.NewLabel("Nome:"), nameEntry,
		widget.NewLabel("Preco:"), priceEntry,
		widget.NewLabel("Categoria:"), categoryEntry,
		container.NewHBox(addBtn, updateBtn, removeBtn),
	)

	content := container.NewBorder(nil, formPanel, nil, nil, itemList)

	d := dialog.NewCustom("Editar Cardapio", "Fechar", content, a.mainWindow)
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}

func (a *App) saveMenuAndRefresh(activeItems *[]pos.MenuItem, list *widget.List,
	nameEntry, priceEntry, categoryEntry *widget.Entry) {

	if err := storage.SaveMenu(a.menu); err != nil {
		log.Printf("Erro ao salvar cardapio: %v", err)
	}
	*activeItems = a.activeMenuItems()
	list.Refresh()
	nameEntry.SetText("")
	priceEntry.SetText("")
	categoryEntry.SetText("")
	a.refreshMenuTabs()
}

func (a *App) activeMenuItems() []pos.MenuItem {
	var items []pos.MenuItem
	for _, item := range a.menu.Items {
		if item.Active {
			items = append(items, item)
		}
	}
	return items
}

func parsePrice(text string) int64 {
	cents, _ := parseCurrencyInput(text)
	return cents
}

func formatPriceForEdit(centavos int64) string {
	return fmt.Sprintf("%.2f", float64(centavos)/100)
}
