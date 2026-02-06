package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
)

func (a *App) buildOrderPanel() fyne.CanvasObject {
	a.totalLabel = widget.NewLabel(pos.FormatBRL(0))
	a.totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	a.orderList = widget.NewList(
		func() int {
			return len(a.order.Items)
		},
		func() fyne.CanvasObject {
			qty := widget.NewLabel("0x")
			name := widget.NewLabel("Item Name")
			name.Wrapping = fyne.TextTruncate
			notes := widget.NewLabel("")
			notes.TextStyle = fyne.TextStyle{Italic: true}
			notes.Wrapping = fyne.TextTruncate
			price := widget.NewLabel("R$ 0,00")
			plus := widget.NewButton("+", nil)
			minus := widget.NewButton("-", nil)
			remove := widget.NewButton("X", nil)
			editNotes := widget.NewButton("Obs", nil)

			controls := container.NewHBox(minus, qty, plus)
			nameBlock := container.NewVBox(name, notes)
			row := container.NewBorder(
				nil, nil,
				controls,
				container.NewHBox(price, editNotes, remove),
				nameBlock,
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.order.Items) {
				return
			}
			orderItem := a.order.Items[id]

			row := obj.(*fyne.Container)
			nameBlock := row.Objects[0].(*fyne.Container)
			leftBox := row.Objects[1].(*fyne.Container)
			rightBox := row.Objects[2].(*fyne.Container)

			nameLabel := nameBlock.Objects[0].(*widget.Label)
			notesLabel := nameBlock.Objects[1].(*widget.Label)

			qtyLabel := leftBox.Objects[1].(*widget.Label)
			minusBtn := leftBox.Objects[0].(*widget.Button)
			plusBtn := leftBox.Objects[2].(*widget.Button)

			priceLabel := rightBox.Objects[0].(*widget.Label)
			editNotesBtn := rightBox.Objects[1].(*widget.Button)
			removeBtn := rightBox.Objects[2].(*widget.Button)

			qtyLabel.SetText(fmt.Sprintf("%dx", orderItem.Quantity))
			nameLabel.SetText(orderItem.Item.Name)
			priceLabel.SetText(pos.FormatBRL(orderItem.Total()))

			if orderItem.Notes != "" {
				notesLabel.SetText("* " + orderItem.Notes)
				notesLabel.Show()
			} else {
				notesLabel.SetText("")
				notesLabel.Hide()
			}

			idx := id
			plusBtn.OnTapped = func() {
				if idx < len(a.order.Items) {
					a.order.UpdateQuantity(idx, a.order.Items[idx].Quantity+1)
					a.refreshOrderDisplay()
				}
			}
			minusBtn.OnTapped = func() {
				if idx < len(a.order.Items) {
					a.order.UpdateQuantity(idx, a.order.Items[idx].Quantity-1)
					a.refreshOrderDisplay()
				}
			}
			removeBtn.OnTapped = func() {
				if idx < len(a.order.Items) {
					a.order.RemoveItem(idx)
					a.refreshOrderDisplay()
				}
			}
			editNotesBtn.OnTapped = func() {
				if idx < len(a.order.Items) {
					a.showEditNotesDialog(idx)
				}
			}
		},
	)

	header := widget.NewLabelWithStyle(
		fmt.Sprintf("Pedido #%d", a.order.Number),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	totalRow := container.New(
		layout.NewHBoxLayout(),
		layout.NewSpacer(),
		widget.NewLabelWithStyle("TOTAL:", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		a.totalLabel,
	)

	return container.NewBorder(header, totalRow, nil, nil, a.orderList)
}
