package printer

import (
	"fmt"

	"notinha/internal/pos"
	"notinha/internal/storage"
)

type SummaryReceiptData struct {
	Restaurant   storage.RestaurantInfo
	Summary      pos.DaySummary
	CharsPerLine int
}

func BuildSummaryReceipt(data SummaryReceiptData) []byte {
	w := data.CharsPerLine
	if w <= 0 {
		w = 48
	}
	s := data.Summary

	rb := NewReceiptBuilder()

	// Header
	rb.AlignCenter().
		FontDouble().Bold().
		Line(data.Restaurant.Name).
		FontNormal().NoBold()

	if data.Restaurant.Address != "" {
		rb.Line(data.Restaurant.Address)
	}
	if data.Restaurant.Phone != "" {
		rb.Line("Tel: " + data.Restaurant.Phone)
	}

	rb.Separator('-', w)

	rb.FontDouble().Bold().
		Line("RESUMO DO DIA").
		FontNormal().NoBold()

	rb.Line(pos.FormatDateBR(s.Date))
	rb.Separator('-', w)

	// Order counts
	rb.AlignLeft()
	rb.Line(formatTotalLine("Total de pedidos:", fmt.Sprintf("%d", s.TotalOrders), w))
	rb.Line(formatTotalLine("Finalizados:", fmt.Sprintf("%d", s.FinalizedOrders), w))
	rb.Line(formatTotalLine("Cancelados:", fmt.Sprintf("%d", s.CancelledOrders), w))

	rb.Separator('-', w)

	// Revenue
	rb.Bold().
		Line(formatTotalLine("RECEITA TOTAL:", pos.FormatBRL(s.TotalRevenue), w)).
		NoBold()

	rb.Separator('-', w)

	// Payment breakdown
	rb.AlignCenter().
		Bold().Line("POR FORMA DE PAGAMENTO").NoBold()
	rb.AlignLeft()

	for _, pm := range []pos.PaymentMethod{pos.PaymentDinheiro, pos.PaymentCartao, pos.PaymentPix} {
		count := s.OrdersByPayment[pm]
		if count == 0 {
			continue
		}
		revenue := s.ByPayment[pm]
		label := fmt.Sprintf("%s (%d):", pm, count)
		rb.Line(formatTotalLine(label, pos.FormatBRL(revenue), w))
	}

	rb.Separator('-', w)

	// Average ticket
	rb.Line(formatTotalLine("Ticket medio:", pos.FormatBRL(s.AverageTicket), w))

	// Footer
	if data.Restaurant.Footer != "" {
		rb.Separator('-', w).
			AlignCenter().
			Line(data.Restaurant.Footer)
	}

	rb.Feed(4).PartialCut()

	return rb.Build()
}
