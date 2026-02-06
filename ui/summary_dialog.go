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
	"notinha/internal/printer"
	"notinha/internal/storage"
)

func (a *App) showDaySummaryDialog() {
	dates, err := storage.ListOrderDates()
	if err != nil {
		log.Printf("Erro ao listar datas: %v", err)
		dialog.ShowError(fmt.Errorf("erro ao carregar historico: %w", err), a.mainWindow)
		return
	}
	if len(dates) == 0 {
		dialog.ShowInformation("Resumo", "Nenhum pedido registrado.", a.mainWindow)
		return
	}

	displayDates := make([]string, len(dates))
	for i, d := range dates {
		displayDates[i] = pos.FormatDateBR(d)
	}

	summaryLabel := widget.NewLabel("")
	summaryLabel.Wrapping = fyne.TextWrapWord
	summaryScroll := container.NewVScroll(summaryLabel)

	var currentDate string

	printBtn := widget.NewButton("Imprimir Resumo", func() {
		if currentDate == "" {
			return
		}
		a.printDaySummary(currentDate)
	})

	loadSummary := func(isoDate string) {
		currentDate = isoDate
		orders, err := storage.LoadDayOrders(isoDate)
		if err != nil {
			log.Printf("Erro ao carregar pedidos: %v", err)
			summaryLabel.SetText("Erro ao carregar dados.")
			return
		}
		s := pos.ComputeDaySummary(isoDate, orders)
		summaryLabel.SetText(formatSummaryText(s))
	}

	dateSelect := widget.NewSelect(displayDates, func(selected string) {
		for i, d := range displayDates {
			if d == selected {
				loadSummary(dates[i])
				return
			}
		}
	})
	dateSelect.SetSelectedIndex(0)
	loadSummary(dates[0])

	content := container.NewBorder(dateSelect, printBtn, nil, nil, summaryScroll)

	d := dialog.NewCustom("Resumo do Dia", "Fechar", content, a.mainWindow)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

func formatSummaryText(s pos.DaySummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Data: %s\n\n", pos.FormatDateBR(s.Date))
	fmt.Fprintf(&b, "Total de pedidos: %d\n", s.TotalOrders)
	fmt.Fprintf(&b, "Finalizados: %d\n", s.FinalizedOrders)
	fmt.Fprintf(&b, "Cancelados: %d\n", s.CancelledOrders)

	b.WriteString("\n--- Receita ---\n")
	fmt.Fprintf(&b, "Total: %s\n", pos.FormatBRL(s.TotalRevenue))
	fmt.Fprintf(&b, "Ticket medio: %s\n", pos.FormatBRL(s.AverageTicket))

	b.WriteString("\n--- Por Forma de Pagamento ---\n")
	for _, pm := range []pos.PaymentMethod{pos.PaymentDinheiro, pos.PaymentCartao, pos.PaymentPix} {
		revenue := s.ByPayment[pm]
		count := s.OrdersByPayment[pm]
		if count > 0 {
			fmt.Fprintf(&b, "%s: %d pedidos - %s\n", pm, count, pos.FormatBRL(revenue))
		}
	}

	return b.String()
}

func (a *App) printDaySummary(isoDate string) {
	if !a.requirePrinterConnected() {
		return
	}

	go func() {
		orders, err := storage.LoadDayOrders(isoDate)
		if err != nil {
			log.Printf("Erro ao carregar pedidos para impressao: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("erro ao carregar pedidos: %w", err), a.mainWindow)
			})
			return
		}

		summary := pos.ComputeDaySummary(isoDate, orders)
		data := printer.SummaryReceiptData{
			Restaurant:   a.config.Restaurant,
			Summary:      summary,
			CharsPerLine: a.config.Printer.CharsPerLine,
		}

		receipt := printer.BuildSummaryReceipt(data)
		if err := a.printer.Write(receipt); err != nil {
			log.Printf("Erro ao imprimir resumo: %v", err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("erro ao imprimir: %w", err), a.mainWindow)
			})
			return
		}
		fyne.Do(func() {
			dialog.ShowInformation("Sucesso", "Resumo impresso!", a.mainWindow)
		})
	}()
}
