package pos

import (
	"encoding/json"
	"strings"
	"testing"
)

// oldOrderJSON simulates a JSON file saved before the new features were added.
// It has no "payments", "cash_received", or item "notes" fields.
const oldOrderJSON = `[
  {
    "number": 1,
    "items": [
      {
        "item": {"id": 1, "name": "Cafe", "price": 550, "category": "Bebidas", "active": true},
        "quantity": 2,
        "notes": ""
      },
      {
        "item": {"id": 2, "name": "Bolo", "price": 1200, "category": "Doces", "active": true},
        "quantity": 1,
        "notes": ""
      }
    ],
    "customer": "Maria",
    "table": "3",
    "discount": 100,
    "payment": "Dinheiro",
    "status": "Finalizado",
    "created_at": "2026-01-15T10:00:00Z",
    "closed_at": "2026-01-15T10:30:00Z"
  },
  {
    "number": 2,
    "items": [
      {
        "item": {"id": 3, "name": "Suco", "price": 800, "category": "Bebidas", "active": true},
        "quantity": 1,
        "notes": ""
      }
    ],
    "customer": "",
    "table": "",
    "discount": 0,
    "payment": "Pix",
    "status": "Finalizado",
    "created_at": "2026-01-15T11:00:00Z",
    "closed_at": "2026-01-15T11:05:00Z"
  },
  {
    "number": 3,
    "items": [],
    "customer": "",
    "table": "",
    "discount": 0,
    "payment": "Dinheiro",
    "status": "Cancelado",
    "created_at": "2026-01-15T12:00:00Z",
    "closed_at": "2026-01-15T12:01:00Z"
  }
]`

// newOrderJSON simulates a JSON file with all new fields populated.
const newOrderJSON = `[
  {
    "number": 10,
    "items": [
      {
        "item": {"id": 1, "name": "Pizza", "price": 3000, "category": "Pizzas", "active": true},
        "quantity": 1,
        "notes": "sem cebola"
      },
      {
        "item": {"id": 2, "name": "Refrigerante", "price": 700, "category": "Bebidas", "active": true},
        "quantity": 2,
        "notes": ""
      }
    ],
    "customer": "Joao",
    "table": "5",
    "discount": 0,
    "payment": "Dinheiro",
    "payments": [
      {"method": "Dinheiro", "amount": 2200},
      {"method": "Pix", "amount": 2200}
    ],
    "cash_received": 3000,
    "status": "Finalizado",
    "created_at": "2026-02-06T14:00:00Z",
    "closed_at": "2026-02-06T14:15:00Z"
  }
]`

func TestBackwardCompatLoadOldOrders(t *testing.T) {
	var orders []Order
	if err := json.Unmarshal([]byte(oldOrderJSON), &orders); err != nil {
		t.Fatalf("Failed to unmarshal old order JSON: %v", err)
	}

	if len(orders) != 3 {
		t.Fatalf("Expected 3 orders, got %d", len(orders))
	}

	// Order 1: Dinheiro, has items, has discount
	o1 := orders[0]
	if o1.Number != 1 {
		t.Errorf("Order 1 number = %d, want 1", o1.Number)
	}
	if o1.Customer != "Maria" {
		t.Errorf("Order 1 customer = %q, want %q", o1.Customer, "Maria")
	}
	if o1.Payment != PaymentDinheiro {
		t.Errorf("Order 1 payment = %q, want %q", o1.Payment, PaymentDinheiro)
	}
	if o1.Status != StatusFinalizado {
		t.Errorf("Order 1 status = %q, want %q", o1.Status, StatusFinalizado)
	}
	if len(o1.Items) != 2 {
		t.Fatalf("Order 1 items count = %d, want 2", len(o1.Items))
	}
	if o1.Subtotal() != 2300 {
		t.Errorf("Order 1 subtotal = %d, want 2300", o1.Subtotal())
	}
	if o1.Total() != 2200 {
		t.Errorf("Order 1 total = %d, want 2200", o1.Total())
	}

	// New fields should have zero values
	if len(o1.Payments) != 0 {
		t.Errorf("Order 1 Payments should be empty, got %d", len(o1.Payments))
	}
	if o1.CashReceived != 0 {
		t.Errorf("Order 1 CashReceived = %d, want 0", o1.CashReceived)
	}
	if o1.IsSplitPayment() {
		t.Error("Order 1 should not be split payment")
	}

	// EffectivePayments should fall back to single payment
	ep := o1.EffectivePayments()
	if len(ep) != 1 {
		t.Fatalf("Order 1 EffectivePayments count = %d, want 1", len(ep))
	}
	if ep[0].Method != PaymentDinheiro {
		t.Errorf("Order 1 effective method = %q, want %q", ep[0].Method, PaymentDinheiro)
	}
	if ep[0].Amount != 2200 {
		t.Errorf("Order 1 effective amount = %d, want 2200", ep[0].Amount)
	}

	// CashChange should be 0 (no CashReceived)
	if o1.CashChange() != 0 {
		t.Errorf("Order 1 CashChange = %d, want 0", o1.CashChange())
	}

	// Order 2: Pix payment
	o2 := orders[1]
	if o2.Payment != PaymentPix {
		t.Errorf("Order 2 payment = %q, want %q", o2.Payment, PaymentPix)
	}
	ep2 := o2.EffectivePayments()
	if ep2[0].Method != PaymentPix {
		t.Errorf("Order 2 effective method = %q, want %q", ep2[0].Method, PaymentPix)
	}

	// Order 3: Cancelled
	o3 := orders[2]
	if o3.Status != StatusCancelado {
		t.Errorf("Order 3 status = %q, want %q", o3.Status, StatusCancelado)
	}
}

func TestBackwardCompatDaySummaryOldOrders(t *testing.T) {
	var orders []Order
	if err := json.Unmarshal([]byte(oldOrderJSON), &orders); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	s := ComputeDaySummary("2026-01-15", orders)

	if s.TotalOrders != 3 {
		t.Errorf("TotalOrders = %d, want 3", s.TotalOrders)
	}
	if s.FinalizedOrders != 2 {
		t.Errorf("FinalizedOrders = %d, want 2", s.FinalizedOrders)
	}
	if s.CancelledOrders != 1 {
		t.Errorf("CancelledOrders = %d, want 1", s.CancelledOrders)
	}
	// Order 1 total: 2300-100=2200, Order 2 total: 800
	if s.TotalRevenue != 3000 {
		t.Errorf("TotalRevenue = %d, want 3000", s.TotalRevenue)
	}
	if s.ByPayment[PaymentDinheiro] != 2200 {
		t.Errorf("ByPayment[Dinheiro] = %d, want 2200", s.ByPayment[PaymentDinheiro])
	}
	if s.ByPayment[PaymentPix] != 800 {
		t.Errorf("ByPayment[Pix] = %d, want 800", s.ByPayment[PaymentPix])
	}
}

func TestNewOrderFieldsLoadCorrectly(t *testing.T) {
	var orders []Order
	if err := json.Unmarshal([]byte(newOrderJSON), &orders); err != nil {
		t.Fatalf("Failed to unmarshal new order JSON: %v", err)
	}

	if len(orders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(orders))
	}

	o := orders[0]

	// Item notes
	if o.Items[0].Notes != "sem cebola" {
		t.Errorf("Item 0 notes = %q, want %q", o.Items[0].Notes, "sem cebola")
	}
	if o.Items[1].Notes != "" {
		t.Errorf("Item 1 notes = %q, want empty", o.Items[1].Notes)
	}

	// Split payments
	if !o.IsSplitPayment() {
		t.Error("Should be split payment")
	}
	if len(o.Payments) != 2 {
		t.Fatalf("Payments count = %d, want 2", len(o.Payments))
	}
	if o.Payments[0].Method != PaymentDinheiro || o.Payments[0].Amount != 2200 {
		t.Errorf("Payment 0 = %v, want {Dinheiro 2200}", o.Payments[0])
	}
	if o.Payments[1].Method != PaymentPix || o.Payments[1].Amount != 2200 {
		t.Errorf("Payment 1 = %v, want {Pix 2200}", o.Payments[1])
	}

	// EffectivePayments returns the splits
	ep := o.EffectivePayments()
	if len(ep) != 2 {
		t.Fatalf("EffectivePayments count = %d, want 2", len(ep))
	}

	// Backward compat: Payment field set to first split method
	if o.Payment != PaymentDinheiro {
		t.Errorf("Payment = %q, want %q (backward compat)", o.Payment, PaymentDinheiro)
	}

	// Cash received and change
	if o.CashReceived != 3000 {
		t.Errorf("CashReceived = %d, want 3000", o.CashReceived)
	}
	// Cash portion = 2200 (Dinheiro split), received = 3000, change = 800
	if o.CashChange() != 800 {
		t.Errorf("CashChange = %d, want 800", o.CashChange())
	}
}

func TestNewOrderDaySummaryWithSplits(t *testing.T) {
	var orders []Order
	if err := json.Unmarshal([]byte(newOrderJSON), &orders); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	s := ComputeDaySummary("2026-02-06", orders)

	// Total: 3000+1400=4400
	if s.TotalRevenue != 4400 {
		t.Errorf("TotalRevenue = %d, want 4400", s.TotalRevenue)
	}
	if s.ByPayment[PaymentDinheiro] != 2200 {
		t.Errorf("ByPayment[Dinheiro] = %d, want 2200", s.ByPayment[PaymentDinheiro])
	}
	if s.ByPayment[PaymentPix] != 2200 {
		t.Errorf("ByPayment[Pix] = %d, want 2200", s.ByPayment[PaymentPix])
	}
}

func TestRoundTripSerialization(t *testing.T) {
	// Create an order with all new features, serialize, deserialize, verify
	order := NewOrder(42)
	order.AddItem(MenuItem{ID: 1, Name: "Pizza", Price: 3000, Category: "Pizzas", Active: true}, 1, "sem cebola")
	order.AddItem(MenuItem{ID: 2, Name: "Suco", Price: 800, Category: "Bebidas", Active: true}, 2, "")
	order.Customer = "Carlos"
	order.Table = "7"
	order.Discount = 200

	order.FinalizeSplit([]PaymentSplit{
		{Method: PaymentDinheiro, Amount: 2300},
		{Method: PaymentCartao, Amount: 2300},
	})
	order.CashReceived = 2500

	// Serialize
	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Deserialize
	var loaded Order
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify all fields
	if loaded.Number != 42 {
		t.Errorf("Number = %d, want 42", loaded.Number)
	}
	if len(loaded.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(loaded.Items))
	}
	if loaded.Items[0].Notes != "sem cebola" {
		t.Errorf("Item 0 notes = %q, want %q", loaded.Items[0].Notes, "sem cebola")
	}
	if loaded.Customer != "Carlos" {
		t.Errorf("Customer = %q, want %q", loaded.Customer, "Carlos")
	}
	if loaded.Discount != 200 {
		t.Errorf("Discount = %d, want 200", loaded.Discount)
	}
	if loaded.Status != StatusFinalizado {
		t.Errorf("Status = %q, want %q", loaded.Status, StatusFinalizado)
	}

	// Split payments
	if !loaded.IsSplitPayment() {
		t.Error("Should be split payment")
	}
	if len(loaded.Payments) != 2 {
		t.Fatalf("Payments count = %d, want 2", len(loaded.Payments))
	}
	// Backward compat Payment field
	if loaded.Payment != PaymentDinheiro {
		t.Errorf("Payment = %q, want %q", loaded.Payment, PaymentDinheiro)
	}

	// Cash
	if loaded.CashReceived != 2500 {
		t.Errorf("CashReceived = %d, want 2500", loaded.CashReceived)
	}
	if loaded.CashChange() != 200 {
		t.Errorf("CashChange = %d, want 200", loaded.CashChange())
	}

	// Total: (3000 + 1600) - 200 = 4400
	if loaded.Total() != 4400 {
		t.Errorf("Total = %d, want 4400", loaded.Total())
	}
}

func TestOmitemptyNewFields(t *testing.T) {
	// Old-style order (no splits, no cash) should NOT emit the new fields in JSON
	order := NewOrder(1)
	order.AddItem(MenuItem{ID: 1, Name: "Cafe", Price: 550, Active: true}, 1, "")
	order.Finalize(PaymentPix)

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	jsonStr := string(data)

	// "payments" should not appear (omitempty + nil slice)
	if containsField(jsonStr, "payments") {
		t.Errorf("JSON should not contain 'payments' field for non-split order, got: %s", jsonStr)
	}
	// "cash_received" should not appear (omitempty + zero)
	if containsField(jsonStr, "cash_received") {
		t.Errorf("JSON should not contain 'cash_received' field when zero, got: %s", jsonStr)
	}
}

func TestMixedOldAndNewOrders(t *testing.T) {
	// Simulate a day file with both old-format and new-format orders
	mixedJSON := `[
		{
			"number": 1,
			"items": [{"item": {"id": 1, "name": "Cafe", "price": 550, "category": "Bebidas", "active": true}, "quantity": 2, "notes": ""}],
			"customer": "", "table": "", "discount": 0,
			"payment": "Dinheiro",
			"status": "Finalizado",
			"created_at": "2026-02-06T09:00:00Z",
			"closed_at": "2026-02-06T09:05:00Z"
		},
		{
			"number": 2,
			"items": [{"item": {"id": 2, "name": "Pizza", "price": 3000, "category": "Pizzas", "active": true}, "quantity": 1, "notes": "extra queijo"}],
			"customer": "Ana", "table": "2", "discount": 0,
			"payment": "Dinheiro",
			"payments": [
				{"method": "Dinheiro", "amount": 1500},
				{"method": "Cartao", "amount": 1500}
			],
			"cash_received": 2000,
			"status": "Finalizado",
			"created_at": "2026-02-06T12:00:00Z",
			"closed_at": "2026-02-06T12:30:00Z"
		}
	]`

	var orders []Order
	if err := json.Unmarshal([]byte(mixedJSON), &orders); err != nil {
		t.Fatalf("Failed to unmarshal mixed JSON: %v", err)
	}

	s := ComputeDaySummary("2026-02-06", orders)

	// Order 1: 1100 all Dinheiro; Order 2: 1500 Dinheiro + 1500 Cartao
	if s.TotalRevenue != 4100 {
		t.Errorf("TotalRevenue = %d, want 4100", s.TotalRevenue)
	}
	if s.ByPayment[PaymentDinheiro] != 2600 {
		t.Errorf("ByPayment[Dinheiro] = %d, want 2600", s.ByPayment[PaymentDinheiro])
	}
	if s.ByPayment[PaymentCartao] != 1500 {
		t.Errorf("ByPayment[Cartao] = %d, want 1500", s.ByPayment[PaymentCartao])
	}

	// Order 2 cash change: received 2000, cash portion 1500, change 500
	if orders[1].CashChange() != 500 {
		t.Errorf("Order 2 CashChange = %d, want 500", orders[1].CashChange())
	}

	// Order 1 has no cash received set
	if orders[0].CashChange() != 0 {
		t.Errorf("Order 1 CashChange = %d, want 0", orders[0].CashChange())
	}

	// Item notes preserved
	if orders[1].Items[0].Notes != "extra queijo" {
		t.Errorf("Order 2 item notes = %q, want %q", orders[1].Items[0].Notes, "extra queijo")
	}
}

func containsField(jsonStr, field string) bool {
	return strings.Contains(jsonStr, `"`+field+`"`)
}
