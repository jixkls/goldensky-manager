package printer

import (
	"fmt"
)

// BuildKitchenTicket constructs a kitchen-only ticket (no prices) as ESC/POS bytes.
func BuildKitchenTicket(data ReceiptData) []byte {
	w := data.CharsPerLine
	if w <= 0 {
		w = 48
	}

	rb := NewReceiptBuilder()

	rb.AlignCenter().
		FontDouble().Bold().
		Line("*** COZINHA ***").
		FontNormal().NoBold()

	rb.Separator('-', w)

	rb.AlignLeft().Bold().FontDouble()
	rb.Line(fmt.Sprintf("Pedido: #%d", data.Order.Number))
	rb.Line(formatDateTime(data.Order.CreatedAt))

	if data.Order.Customer != "" {
		rb.Line("Cliente: " + truncate(data.Order.Customer, w/2-5))
	}
	if data.Order.Table != "" {
		rb.Line("Mesa: " + data.Order.Table)
	}

	rb.FontNormal().NoBold()
	rb.Separator('-', w)

	for _, oi := range data.Order.Items {
		rb.Bold().
			Line(fmt.Sprintf("%dx %s", oi.Quantity, truncate(oi.Item.Name, w-5)))
		rb.NoBold()
		if oi.Notes != "" {
			rb.Line("  * " + oi.Notes)
		}
	}

	rb.Separator('-', w)
	rb.Feed(4).PartialCut()

	return rb.Build()
}
