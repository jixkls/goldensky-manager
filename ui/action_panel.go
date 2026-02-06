package ui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
	"notinha/internal/printer"
	"notinha/internal/storage"
)

func (a *App) buildActionPanel() fyne.CanvasObject {
	// Customer and table entries
	a.customerEntry = widget.NewEntry()
	a.customerEntry.SetPlaceHolder("Nome do cliente")
	a.tableEntry = widget.NewEntry()
	a.tableEntry.SetPlaceHolder("Mesa")

	// Payment section: cash entry, radio, split button
	a.cashReceivedEntry = widget.NewEntry()
	a.cashReceivedEntry.SetPlaceHolder("Valor recebido (R$)")
	a.cashReceivedEntry.OnChanged = func(_ string) {
		a.updateChangeDisplay()
	}
	a.changeLabel = widget.NewLabel("")
	a.changeLabel.TextStyle = fyne.TextStyle{Bold: true}
	a.cashSection = container.NewVBox(
		widget.NewLabel("Troco:"),
		a.cashReceivedEntry,
		a.changeLabel,
	)
	a.paymentRadio = widget.NewRadioGroup(
		pos.PaymentMethodLabels(),
		func(selected string) {
			a.order.Payment = pos.PaymentMethod(selected)
			if selected == string(pos.PaymentDinheiro) {
				a.cashSection.Show()
			} else {
				a.cashSection.Hide()
			}
		},
	)
	a.paymentRadio.SetSelected(string(pos.PaymentDinheiro))
	splitPaymentBtn := widget.NewButton("Dividir Pagamento", func() {
		a.showSplitPaymentDialog()
	})

	// Discount and action buttons
	a.discountEntry = widget.NewEntry()
	a.discountEntry.SetPlaceHolder("Desconto (R$)")
	finalizeBtn := widget.NewButton("Finalizar Pedido", func() {
		a.finalizeOrder()
	})
	finalizeBtn.Importance = widget.HighImportance
	newOrderBtn := widget.NewButton("Novo Pedido", func() {
		a.newOrder()
	})
	a.kitchenCheck = widget.NewCheck("Comanda de Cozinha", func(checked bool) {
		a.config.KitchenTicket = checked
		_ = storage.SaveConfig(a.config)
	})
	a.kitchenCheck.SetChecked(a.config.KitchenTicket)

	// Printer controls
	printTestBtn := widget.NewButton("Teste Impressao", func() {
		a.printTest()
	})
	openDrawerBtn := widget.NewButton("Abrir Gaveta", func() {
		a.openDrawer()
	})

	// Layout assembly
	panel := container.NewVBox(
		widget.NewLabelWithStyle("Dados do Pedido", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Cliente:"),
		a.customerEntry,
		widget.NewLabel("Mesa:"),
		a.tableEntry,
		widget.NewSeparator(),
		widget.NewLabel("Pagamento:"),
		a.paymentRadio,
		splitPaymentBtn,
		a.cashSection,
		widget.NewSeparator(),
		widget.NewLabel("Desconto:"),
		a.discountEntry,
		widget.NewSeparator(),
		a.kitchenCheck,
		finalizeBtn,
		newOrderBtn,
		layout.NewSpacer(),
		widget.NewSeparator(),
		printTestBtn,
		openDrawerBtn,
	)

	scrollable := container.NewVScroll(panel)
	scrollable.SetMinSize(fyne.NewSize(300, 0))
	return scrollable
}

func (a *App) updateChangeDisplay() {
	received, ok := parseCurrencyInput(a.cashReceivedEntry.Text)
	if !ok {
		a.changeLabel.SetText("")
		return
	}
	total := a.order.Total()
	if received >= total {
		a.changeLabel.SetText(fmt.Sprintf("Troco: %s", pos.FormatBRL(received-total)))
	} else {
		a.changeLabel.SetText(fmt.Sprintf("Falta: %s", pos.FormatBRL(total-received)))
	}
}

func (a *App) hasCashPayment() bool {
	if len(a.splitPayments) > 0 {
		for _, s := range a.splitPayments {
			if s.Method == pos.PaymentDinheiro {
				return true
			}
		}
		return false
	}
	return a.order.Payment == pos.PaymentDinheiro
}

func (a *App) finalizeOrder() {
	if len(a.order.Items) == 0 {
		dialog.ShowInformation("Aviso", "Adicione itens ao pedido.", a.mainWindow)
		return
	}

	a.order.Customer = a.customerEntry.Text
	a.order.Table = a.tableEntry.Text

	if cents, ok := parseCurrencyInput(a.discountEntry.Text); ok {
		a.order.Discount = cents
	}

	if cents, ok := parseCurrencyInput(a.cashReceivedEntry.Text); ok {
		a.order.CashReceived = cents
	}

	if len(a.splitPayments) > 0 {
		a.order.FinalizeSplit(a.splitPayments)
	} else {
		a.order.Finalize(a.order.Payment)
	}

	if a.order.CashReceived > 0 && a.hasCashPayment() {
		change := a.order.CashChange()
		msg := fmt.Sprintf("Total: %s\nRecebido: %s\nTroco: %s",
			pos.FormatBRL(a.order.Total()),
			pos.FormatBRL(a.order.CashReceived),
			pos.FormatBRL(change))
		dialog.ShowConfirm("Confirmar Troco", msg, func(ok bool) {
			if ok {
				a.executeFinalizeOrder()
			}
		}, a.mainWindow)
		return
	}

	a.executeFinalizeOrder()
}

func (a *App) executeFinalizeOrder() {
	if err := storage.SaveOrder(a.order); err != nil {
		log.Printf("Erro ao salvar pedido: %v", err)
	}

	if a.printer != nil && a.printer.IsConnected() {
		printKitchen := a.config.KitchenTicket
		go func() {
			data := printer.ReceiptData{
				Restaurant:   a.config.Restaurant,
				Order:        a.order,
				CharsPerLine: a.config.Printer.CharsPerLine,
			}
			receipt := printer.BuildReceipt(data)
			if err := a.printer.Write(receipt); err != nil {
				log.Printf("Erro ao imprimir: %v", err)
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("erro ao imprimir: %w", err), a.mainWindow)
				})
				return
			}
			if printKitchen {
				ticket := printer.BuildKitchenTicket(data)
				if err := a.printer.Write(ticket); err != nil {
					log.Printf("Erro ao imprimir comanda de cozinha: %v", err)
				}
			}
			fyne.Do(func() {
				dialog.ShowInformation("Sucesso", "Pedido impresso!", a.mainWindow)
			})
		}()
	} else {
		dialog.ShowInformation("Pedido Finalizado",
			fmt.Sprintf("Pedido #%d finalizado.\nTotal: %s\n(Impressora nao conectada)",
				a.order.Number, pos.FormatBRL(a.order.Total())),
			a.mainWindow)
	}

	a.newOrder()
}

func (a *App) showSplitPaymentDialog() {
	total := a.order.Total()
	if total <= 0 {
		dialog.ShowInformation("Aviso", "Adicione itens ao pedido.", a.mainWindow)
		return
	}

	// State setup
	splits := make([]pos.PaymentSplit, 0)
	var remaining int64 = total

	// Display labels
	remainingLabel := widget.NewLabel(fmt.Sprintf("Restante: %s", pos.FormatBRL(remaining)))
	remainingLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel := widget.NewLabel(fmt.Sprintf("Total: %s", pos.FormatBRL(total)))
	splitsList := widget.NewLabel("")

	// Input widgets
	methodRadio := widget.NewRadioGroup(pos.PaymentMethodLabels(), nil)
	methodRadio.SetSelected(string(pos.PaymentDinheiro))
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Valor (R$)")

	refreshSplitSummary := func() {
		remaining = total
		for _, s := range splits {
			remaining -= s.Amount
		}
		remainingLabel.SetText(fmt.Sprintf("Restante: %s", pos.FormatBRL(remaining)))

		var lines []string
		for _, s := range splits {
			lines = append(lines, fmt.Sprintf("  %s: %s", s.Method, pos.FormatBRL(s.Amount)))
		}
		if len(lines) > 0 {
			splitsList.SetText(strings.Join(lines, "\n"))
		} else {
			splitsList.SetText("")
		}
	}

	addBtn := widget.NewButton("Adicionar", func() {
		cents, ok := parseCurrencyInput(amountEntry.Text)
		if !ok {
			dialog.ShowInformation("Aviso", "Valor invalido.", a.mainWindow)
			return
		}
		amount := cents
		method := pos.PaymentMethod(methodRadio.Selected)
		splits = append(splits, pos.PaymentSplit{Method: method, Amount: amount})
		amountEntry.SetText("")
		refreshSplitSummary()
	})

	clearBtn := widget.NewButton("Limpar", func() {
		splits = splits[:0]
		amountEntry.SetText("")
		refreshSplitSummary()
	})

	content := container.NewVBox(
		totalLabel,
		remainingLabel,
		widget.NewSeparator(),
		widget.NewLabel("Pagamentos:"),
		splitsList,
		widget.NewSeparator(),
		methodRadio,
		amountEntry,
		container.NewHBox(addBtn, clearBtn),
	)

	d := dialog.NewCustomConfirm("Dividir Pagamento", "Confirmar", "Cancelar", content, func(ok bool) {
		if !ok {
			return
		}
		if len(splits) == 0 {
			return
		}

		var splitTotal int64
		for _, s := range splits {
			splitTotal += s.Amount
		}
		if splitTotal != total {
			dialog.ShowInformation("Aviso",
				fmt.Sprintf("Soma dos pagamentos (%s) difere do total (%s).",
					pos.FormatBRL(splitTotal), pos.FormatBRL(total)),
				a.mainWindow)
			return
		}

		a.splitPayments = splits
		a.paymentRadio.Disable()
		if a.hasCashPayment() {
			a.cashSection.Show()
		} else {
			a.cashSection.Hide()
		}
	}, a.mainWindow)
	d.Resize(fyne.NewSize(400, 450))
	d.Show()
}

func (a *App) printTest() {
	if !a.requirePrinterConnected() {
		return
	}
	go func() {
		if err := a.printer.PrintTest(); err != nil {
			log.Printf("Erro no teste: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("erro no teste: %w", err), a.mainWindow)
			})
		}
	}()
}

func (a *App) openDrawer() {
	if !a.requirePrinterConnected() {
		return
	}
	go func() {
		if err := a.printer.OpenDrawer(); err != nil {
			log.Printf("Erro ao abrir gaveta: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("erro ao abrir gaveta: %w", err), a.mainWindow)
			})
		}
	}()
}

func (a *App) reconnectPrinter() {
	if a.printer != nil {
		a.printer.Close()
		a.printer = nil
	}
	a.connectPrinter()
	a.updatePrinterStatus()
}

func sanitizeDecimal(s string) string {
	// Accept both "10,50" and "10.50"
	result := make([]byte, 0, len(s))
	for _, c := range s {
		if c == ',' {
			result = append(result, '.')
		} else if (c >= '0' && c <= '9') || c == '.' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

func parseCurrencyInput(text string) (int64, bool) {
	text = sanitizeDecimal(text)
	if text == "" {
		return 0, false
	}
	val, err := strconv.ParseFloat(text, 64)
	if err != nil || val <= 0 {
		return 0, false
	}
	return int64(val * 100), true
}

func (a *App) requirePrinterConnected() bool {
	if a.printer != nil && a.printer.IsConnected() {
		return true
	}
	dialog.ShowInformation("Aviso", "Impressora nao conectada.", a.mainWindow)
	return false
}
