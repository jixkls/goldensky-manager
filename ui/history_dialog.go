package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
	"notinha/internal/storage"
)

func (a *App) showOrderHistoryDialog() {
	dates, err := storage.ListOrderDates()
	if err != nil {
		log.Printf("Erro ao listar datas: %v", err)
		dialog.ShowError(fmt.Errorf("erro ao carregar historico: %w", err), a.mainWindow)
		return
	}
	if len(dates) == 0 {
		dialog.ShowInformation("Historico", "Nenhum pedido registrado.", a.mainWindow)
		return
	}

	displayDates := make([]string, len(dates))
	for i, d := range dates {
		displayDates[i] = pos.FormatDateBR(d)
	}

	var orders []pos.Order
	detailLabel := widget.NewLabel("Selecione um pedido.")
	detailLabel.Wrapping = fyne.TextWrapWord
	detailScroll := container.NewVScroll(detailLabel)

	var orderList *widget.List
	orderList = widget.NewList(
		func() int { return len(orders) },
		func() fyne.CanvasObject {
			return widget.NewLabel("#000  00:00  R$ 0,00  Dinheiro")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(orders) {
				return
			}
			o := orders[id]
			timeStr := o.ClosedAt.Format("15:04")
			status := ""
			if o.Status == pos.StatusCancelado {
				status = " [CANCELADO]"
			}
			paymentDisplay := string(o.Payment)
			if o.IsSplitPayment() {
				paymentDisplay = "Dividido"
			}
			obj.(*widget.Label).SetText(
				fmt.Sprintf("#%d  %s  %s  %s%s",
					o.Number, timeStr, pos.FormatBRL(o.Total()), paymentDisplay, status),
			)
		},
	)

	orderList.OnSelected = func(id widget.ListItemID) {
		if id < len(orders) {
			detailLabel.SetText(formatOrderDetail(&orders[id]))
		}
	}

	loadOrders := func(isoDate string) {
		loaded, err := storage.LoadDayOrders(isoDate)
		if err != nil {
			log.Printf("Erro ao carregar pedidos: %v", err)
			orders = nil
		} else {
			orders = loaded
		}
		detailLabel.SetText("Selecione um pedido.")
		orderList.UnselectAll()
		orderList.Refresh()
	}

	dateSelect := widget.NewSelect(displayDates, func(selected string) {
		for i, d := range displayDates {
			if d == selected {
				loadOrders(dates[i])
				return
			}
		}
	})
	dateSelect.SetSelectedIndex(0)
	loadOrders(dates[0])

	leftPanel := container.NewBorder(dateSelect, nil, nil, nil, orderList)
	content := container.NewHSplit(leftPanel, detailScroll)
	content.SetOffset(0.4)

	d := dialog.NewCustom("Historico de Pedidos", "Fechar", content, a.mainWindow)
	d.Resize(fyne.NewSize(750, 500))
	d.Show()
}

func formatOrderDetail(o *pos.Order) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Pedido #%d\n", o.Number)
	fmt.Fprintf(&b, "Status: %s\n", o.Status)
	fmt.Fprintf(&b, "Criado: %s\n", o.CreatedAt.Format("02/01/2006 15:04"))
	if !o.ClosedAt.IsZero() {
		fmt.Fprintf(&b, "Fechado: %s\n", o.ClosedAt.Format("02/01/2006 15:04"))
	}
	if o.Customer != "" {
		fmt.Fprintf(&b, "Cliente: %s\n", o.Customer)
	}
	if o.Table != "" {
		fmt.Fprintf(&b, "Mesa: %s\n", o.Table)
	}
	if o.IsSplitPayment() {
		b.WriteString("Pagamentos:\n")
		for _, p := range o.Payments {
			fmt.Fprintf(&b, "  %s: %s\n", p.Method, pos.FormatBRL(p.Amount))
		}
	} else {
		fmt.Fprintf(&b, "Pagamento: %s\n", o.Payment)
	}

	b.WriteString("\n--- Itens ---\n")
	for _, orderItem := range o.Items {
		fmt.Fprintf(&b, "%dx %s  %s\n", orderItem.Quantity, orderItem.Item.Name, pos.FormatBRL(orderItem.Total()))
		if orderItem.Notes != "" {
			fmt.Fprintf(&b, "   * %s\n", orderItem.Notes)
		}
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "Subtotal: %s\n", pos.FormatBRL(o.Subtotal()))
	if o.Discount > 0 {
		fmt.Fprintf(&b, "Desconto: -%s\n", pos.FormatBRL(o.Discount))
	}
	fmt.Fprintf(&b, "Total: %s\n", pos.FormatBRL(o.Total()))

	if o.CashReceived > 0 {
		fmt.Fprintf(&b, "Valor Recebido: %s\n", pos.FormatBRL(o.CashReceived))
		fmt.Fprintf(&b, "Troco: %s\n", pos.FormatBRL(o.CashChange()))
	}

	return b.String()
}
