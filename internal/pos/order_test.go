package pos

import (
	"testing"
	"time"
)

func TestFormatBRL(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{12345, "R$ 123,45"},
		{100000, "R$ 1.000,00"},
		{0, "R$ 0,00"},
		{1, "R$ 0,01"},
		{999, "R$ 9,99"},
		{1234567, "R$ 12.345,67"},
	}
	for _, tc := range tests {
		got := FormatBRL(tc.input)
		if got != tc.expected {
			t.Errorf("FormatBRL(%d) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestOrderTotals(t *testing.T) {
	order := NewOrder(1)
	item := MenuItem{ID: 1, Name: "Cafe", Price: 550, Category: "Bebidas", Active: true}

	order.AddItem(item, 2, "")
	if order.Subtotal() != 1100 {
		t.Errorf("Subtotal = %d, want 1100", order.Subtotal())
	}

	order.Discount = 100
	if order.Total() != 1000 {
		t.Errorf("Total = %d, want 1000", order.Total())
	}

	// Adding same item should merge
	order.AddItem(item, 1, "")
	if len(order.Items) != 1 {
		t.Errorf("Items count = %d, want 1 (should merge)", len(order.Items))
	}
	if order.Items[0].Quantity != 3 {
		t.Errorf("Quantity = %d, want 3", order.Items[0].Quantity)
	}
}

func TestMenuCategories(t *testing.T) {
	menu := NewMenu()
	menu.AddItem(MenuItem{Name: "Cafe", Price: 550, Category: "Bebidas"})
	menu.AddItem(MenuItem{Name: "Bolo", Price: 1200, Category: "Doces"})
	menu.AddItem(MenuItem{Name: "Suco", Price: 800, Category: "Bebidas"})

	cats := menu.Categories()
	if len(cats) != 2 {
		t.Errorf("Categories count = %d, want 2", len(cats))
	}

	bebidas := menu.ItemsByCategory("Bebidas")
	if len(bebidas) != 2 {
		t.Errorf("Bebidas count = %d, want 2", len(bebidas))
	}
}

func TestFinalize(t *testing.T) {
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 1, "")

	before := time.Now()
	order.Finalize(PaymentPix)
	after := time.Now()

	if order.Status != StatusFinalizado {
		t.Errorf("Status = %q, want %q", order.Status, StatusFinalizado)
	}
	if order.Payment != PaymentPix {
		t.Errorf("Payment = %q, want %q", order.Payment, PaymentPix)
	}
	if order.ClosedAt.Before(before) || order.ClosedAt.After(after) {
		t.Errorf("ClosedAt = %v, want between %v and %v", order.ClosedAt, before, after)
	}
}

func TestCancel(t *testing.T) {
	order := NewOrder(2)
	order.Cancel()

	if order.Status != StatusCancelado {
		t.Errorf("Status = %q, want %q", order.Status, StatusCancelado)
	}
	if order.ClosedAt.IsZero() {
		t.Error("ClosedAt should not be zero after Cancel")
	}
}

func TestComputeDaySummary(t *testing.T) {
	orders := []Order{
		{
			Number:  1,
			Status:  StatusFinalizado,
			Payment: PaymentDinheiro,
			Items:   []OrderItem{{Item: MenuItem{Price: 1000}, Quantity: 2}},
		},
		{
			Number:  2,
			Status:  StatusFinalizado,
			Payment: PaymentPix,
			Items:   []OrderItem{{Item: MenuItem{Price: 500}, Quantity: 1}},
		},
		{
			Number: 3,
			Status: StatusCancelado,
		},
	}

	s := ComputeDaySummary("2026-02-05", orders)

	if s.TotalOrders != 3 {
		t.Errorf("TotalOrders = %d, want 3", s.TotalOrders)
	}
	if s.FinalizedOrders != 2 {
		t.Errorf("FinalizedOrders = %d, want 2", s.FinalizedOrders)
	}
	if s.CancelledOrders != 1 {
		t.Errorf("CancelledOrders = %d, want 1", s.CancelledOrders)
	}
	if s.TotalRevenue != 2500 {
		t.Errorf("TotalRevenue = %d, want 2500", s.TotalRevenue)
	}
	if s.ByPayment[PaymentDinheiro] != 2000 {
		t.Errorf("ByPayment[Dinheiro] = %d, want 2000", s.ByPayment[PaymentDinheiro])
	}
	if s.ByPayment[PaymentPix] != 500 {
		t.Errorf("ByPayment[Pix] = %d, want 500", s.ByPayment[PaymentPix])
	}
	if s.OrdersByPayment[PaymentDinheiro] != 1 {
		t.Errorf("OrdersByPayment[Dinheiro] = %d, want 1", s.OrdersByPayment[PaymentDinheiro])
	}
	if s.AverageTicket != 1250 {
		t.Errorf("AverageTicket = %d, want 1250", s.AverageTicket)
	}
}

func TestAddItemWithNotes(t *testing.T) {
	order := NewOrder(1)
	item := MenuItem{ID: 1, Name: "Pizza", Price: 3000, Active: true}

	order.AddItem(item, 1, "sem cebola")
	order.AddItem(item, 1, "extra queijo")

	if len(order.Items) != 2 {
		t.Errorf("Items count = %d, want 2 (different notes = separate lines)", len(order.Items))
	}

	// Same item with same notes should merge
	order.AddItem(item, 1, "sem cebola")
	if len(order.Items) != 2 {
		t.Errorf("Items count = %d, want 2 (same notes should merge)", len(order.Items))
	}
	if order.Items[0].Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", order.Items[0].Quantity)
	}
}

func TestUpdateNotes(t *testing.T) {
	order := NewOrder(1)
	item := MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}
	order.AddItem(item, 1, "")

	order.UpdateNotes(0, "sem acucar")
	if order.Items[0].Notes != "sem acucar" {
		t.Errorf("Notes = %q, want %q", order.Items[0].Notes, "sem acucar")
	}

	// Invalid indices should be no-ops
	order.UpdateNotes(-1, "nope")
	order.UpdateNotes(5, "nope")
	if order.Items[0].Notes != "sem acucar" {
		t.Errorf("Notes changed on invalid index")
	}
}

func TestFinalizeSplit(t *testing.T) {
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 2, "")

	payments := []PaymentSplit{
		{Method: PaymentDinheiro, Amount: 600},
		{Method: PaymentPix, Amount: 500},
	}

	before := time.Now()
	order.FinalizeSplit(payments)
	after := time.Now()

	if order.Status != StatusFinalizado {
		t.Errorf("Status = %q, want %q", order.Status, StatusFinalizado)
	}
	if len(order.Payments) != 2 {
		t.Errorf("Payments count = %d, want 2", len(order.Payments))
	}
	// Backward compat: Payment field set to first method
	if order.Payment != PaymentDinheiro {
		t.Errorf("Payment = %q, want %q", order.Payment, PaymentDinheiro)
	}
	if order.ClosedAt.Before(before) || order.ClosedAt.After(after) {
		t.Errorf("ClosedAt out of range")
	}
}

func TestEffectivePayments(t *testing.T) {
	// Single payment order (no splits)
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 1, "")
	order.Finalize(PaymentPix)

	ep := order.EffectivePayments()
	if len(ep) != 1 {
		t.Fatalf("EffectivePayments count = %d, want 1", len(ep))
	}
	if ep[0].Method != PaymentPix {
		t.Errorf("Method = %q, want %q", ep[0].Method, PaymentPix)
	}
	if ep[0].Amount != 550 {
		t.Errorf("Amount = %d, want 550", ep[0].Amount)
	}
	if order.IsSplitPayment() {
		t.Error("IsSplitPayment should be false for single payment")
	}
}

func TestComputeDaySummaryWithSplitPayment(t *testing.T) {
	orders := []Order{
		{
			Number:  1,
			Status:  StatusFinalizado,
			Payment: PaymentDinheiro,
			Payments: []PaymentSplit{
				{Method: PaymentDinheiro, Amount: 1000},
				{Method: PaymentPix, Amount: 500},
			},
			Items: []OrderItem{{Item: MenuItem{Price: 1500}, Quantity: 1}},
		},
		{
			Number:  2,
			Status:  StatusFinalizado,
			Payment: PaymentCartao,
			Items:   []OrderItem{{Item: MenuItem{Price: 800}, Quantity: 1}},
		},
	}

	s := ComputeDaySummary("2026-02-06", orders)

	if s.TotalRevenue != 2300 {
		t.Errorf("TotalRevenue = %d, want 2300", s.TotalRevenue)
	}
	if s.ByPayment[PaymentDinheiro] != 1000 {
		t.Errorf("ByPayment[Dinheiro] = %d, want 1000", s.ByPayment[PaymentDinheiro])
	}
	if s.ByPayment[PaymentPix] != 500 {
		t.Errorf("ByPayment[Pix] = %d, want 500", s.ByPayment[PaymentPix])
	}
	if s.ByPayment[PaymentCartao] != 800 {
		t.Errorf("ByPayment[Cartao] = %d, want 800", s.ByPayment[PaymentCartao])
	}
}

func TestCashChange(t *testing.T) {
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 2, "")
	order.Finalize(PaymentDinheiro)
	order.CashReceived = 2000

	change := order.CashChange()
	if change != 900 {
		t.Errorf("CashChange = %d, want 900", change)
	}
}

func TestCashChangeZeroWhenNotCash(t *testing.T) {
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 1, "")
	order.Finalize(PaymentPix)
	order.CashReceived = 1000

	change := order.CashChange()
	if change != 0 {
		t.Errorf("CashChange = %d, want 0 (no cash portion for Pix payment)", change)
	}
}

func TestCashChangeSplitPayment(t *testing.T) {
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Pizza", Price: 3000, Active: true}, 1, "")

	order.FinalizeSplit([]PaymentSplit{
		{Method: PaymentDinheiro, Amount: 1500},
		{Method: PaymentPix, Amount: 1500},
	})
	order.CashReceived = 2000

	// Cash portion is 1500 (Dinheiro split), received is 2000, change = 500
	change := order.CashChange()
	if change != 500 {
		t.Errorf("CashChange = %d, want 500", change)
	}
}

func TestFormatDateBR(t *testing.T) {
	got := FormatDateBR("2026-02-05")
	if got != "05/02/2026" {
		t.Errorf("FormatDateBR = %q, want %q", got, "05/02/2026")
	}
}
