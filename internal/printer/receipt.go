package printer

import (
	"fmt"
	"strings"
	"time"

	"notinha/internal/pos"
	"notinha/internal/storage"
)

// ReceiptData holds all information needed to build a receipt.
type ReceiptData struct {
	Restaurant   storage.RestaurantInfo
	Order        *pos.Order
	CharsPerLine int
}

// BuildReceipt constructs a full receipt and returns the ESC/POS bytes.
func BuildReceipt(data ReceiptData) []byte {
	w := data.CharsPerLine
	if w <= 0 {
		w = 48
	}

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
	if data.Restaurant.CNPJ != "" {
		rb.Line("CNPJ: " + data.Restaurant.CNPJ)
	}

	rb.Separator('-', w)

	// Order info
	rb.AlignLeft().
		Line(formatDateTime(data.Order.CreatedAt)).
		Line(fmt.Sprintf("Pedido: #%d", data.Order.Number))

	if data.Order.Customer != "" {
		rb.Line("Cliente: " + data.Order.Customer)
	}
	if data.Order.Table != "" {
		rb.Line("Mesa: " + data.Order.Table)
	}

	rb.Separator('-', w)

	// Column header
	rb.Bold().Line(formatItemLine("QTD", "ITEM", "VALOR", w)).NoBold()

	// Items
	for _, oi := range data.Order.Items {
		qty := fmt.Sprintf("%dx", oi.Quantity)
		price := pos.FormatBRL(oi.Total())
		rb.Line(formatItemLine(qty, oi.Item.Name, price, w))
		if oi.Notes != "" {
			rb.Line("  * " + oi.Notes)
		}
	}

	rb.Separator('-', w)

	// Totals
	subtotal := data.Order.Subtotal()
	rb.Line(formatTotalLine("Subtotal:", pos.FormatBRL(subtotal), w))

	if data.Order.Discount > 0 {
		rb.Line(formatTotalLine("Desconto:", "-"+pos.FormatBRL(data.Order.Discount), w))
	}

	rb.Bold().
		Line(formatTotalLine("TOTAL:", pos.FormatBRL(data.Order.Total()), w)).
		NoBold()

	rb.Separator('-', w)

	// Payment
	if data.Order.IsSplitPayment() {
		rb.Bold().Line("Pagamentos:").NoBold()
		for _, p := range data.Order.Payments {
			rb.Line(fmt.Sprintf("  %s: %s", p.Method, pos.FormatBRL(p.Amount)))
		}
	} else {
		rb.Line(fmt.Sprintf("Pagamento: %s", data.Order.Payment))
	}

	// Cash received / change
	if data.Order.CashReceived > 0 {
		rb.Line(formatTotalLine("Valor Recebido:", pos.FormatBRL(data.Order.CashReceived), w))
		rb.Line(formatTotalLine("Troco:", pos.FormatBRL(data.Order.CashChange()), w))
	}

	// Footer
	if data.Restaurant.Footer != "" {
		rb.Separator('-', w).
			AlignCenter().
			Line(data.Restaurant.Footer)
	}

	rb.Feed(4).PartialCut()

	return rb.Build()
}

func formatItemLine(qty, name, price string, width int) string {
	qtyWidth := 5
	priceWidth := 12
	nameWidth := width - qtyWidth - priceWidth

	qty = padRight(qty, qtyWidth)
	name = truncate(name, nameWidth)
	name = padRight(name, nameWidth)
	price = padLeft(price, priceWidth)

	return qty + name + price
}

func formatTotalLine(label, value string, width int) string {
	valueWidth := len(value)
	labelWidth := width - valueWidth
	if labelWidth < 0 {
		labelWidth = 0
	}
	return padRight(label, labelWidth) + value
}

func formatDateTime(t time.Time) string {
	return t.Format("02/01/2006 15:04")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 0 {
		return ""
	}
	return s[:maxLen]
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
