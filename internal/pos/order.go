package pos

import (
	"fmt"
	"strings"
	"time"
)

type PaymentMethod string

const (
	PaymentDinheiro PaymentMethod = "Dinheiro"
	PaymentCartao   PaymentMethod = "Cartao"
	PaymentPix      PaymentMethod = "Pix"
)

type OrderStatus string

const (
	StatusAberto     OrderStatus = "Aberto"
	StatusFinalizado OrderStatus = "Finalizado"
	StatusCancelado  OrderStatus = "Cancelado"
)

type MenuItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int64  `json:"price"` // centavos
	Category string `json:"category"`
	Active   bool   `json:"active"`
}

type OrderItem struct {
	Item     MenuItem `json:"item"`
	Quantity int      `json:"quantity"`
	Notes    string   `json:"notes"`
}

func (oi OrderItem) Total() int64 {
	return oi.Item.Price * int64(oi.Quantity)
}

type PaymentSplit struct {
	Method PaymentMethod `json:"method"`
	Amount int64         `json:"amount"`
}

type Order struct {
	Number       int              `json:"number"`
	Items        []OrderItem      `json:"items"`
	Customer     string           `json:"customer"`
	Table        string           `json:"table"`
	Discount     int64            `json:"discount"` // centavos
	Payment      PaymentMethod    `json:"payment"`
	Payments     []PaymentSplit   `json:"payments,omitempty"`
	CashReceived int64            `json:"cash_received,omitempty"`
	Status       OrderStatus      `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	ClosedAt     time.Time        `json:"closed_at,omitempty"`
}

func NewOrder(number int) *Order {
	return &Order{
		Number:    number,
		Items:     []OrderItem{},
		Status:    StatusAberto,
		Payment:   PaymentDinheiro,
		CreatedAt: time.Now(),
	}
}

func (o *Order) AddItem(item MenuItem, quantity int, notes string) {
	for i, oi := range o.Items {
		if oi.Item.ID == item.ID && oi.Notes == notes {
			o.Items[i].Quantity += quantity
			return
		}
	}
	o.Items = append(o.Items, OrderItem{
		Item:     item,
		Quantity: quantity,
		Notes:    notes,
	})
}

func PaymentMethodLabels() []string {
	return []string{string(PaymentDinheiro), string(PaymentCartao), string(PaymentPix)}
}

func (o *Order) isValidItemIndex(index int) bool {
	return index >= 0 && index < len(o.Items)
}

func (o *Order) RemoveItem(index int) {
	if !o.isValidItemIndex(index) {
		return
	}
	o.Items = append(o.Items[:index], o.Items[index+1:]...)
}

func (o *Order) UpdateQuantity(index, quantity int) {
	if !o.isValidItemIndex(index) {
		return
	}
	if quantity <= 0 {
		o.RemoveItem(index)
		return
	}
	o.Items[index].Quantity = quantity
}

func (o *Order) UpdateNotes(index int, notes string) {
	if !o.isValidItemIndex(index) {
		return
	}
	o.Items[index].Notes = notes
}

func (o *Order) Subtotal() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.Total()
	}
	return total
}

func (o *Order) Total() int64 {
	total := o.Subtotal() - o.Discount
	if total < 0 {
		return 0
	}
	return total
}

func (o *Order) Finalize(payment PaymentMethod) {
	o.Status = StatusFinalizado
	o.Payment = payment
	o.ClosedAt = time.Now()
}

func (o *Order) FinalizeSplit(payments []PaymentSplit) {
	o.Status = StatusFinalizado
	o.Payments = payments
	if len(payments) > 0 {
		o.Payment = payments[0].Method
	}
	o.ClosedAt = time.Now()
}

func (o *Order) IsSplitPayment() bool {
	return len(o.Payments) > 1
}

func (o *Order) EffectivePayments() []PaymentSplit {
	if len(o.Payments) > 0 {
		return o.Payments
	}
	return []PaymentSplit{{Method: o.Payment, Amount: o.Total()}}
}

func (o *Order) cashPortion() int64 {
	if len(o.Payments) > 0 {
		var cash int64
		for _, p := range o.Payments {
			if p.Method == PaymentDinheiro {
				cash += p.Amount
			}
		}
		return cash
	}
	if o.Payment == PaymentDinheiro {
		return o.Total()
	}
	return 0
}

func (o *Order) CashChange() int64 {
	portion := o.cashPortion()
	if portion == 0 {
		return 0
	}
	change := o.CashReceived - portion
	if change > 0 {
		return change
	}
	return 0
}

func (o *Order) Cancel() {
	o.Status = StatusCancelado
	o.ClosedAt = time.Now()
}

type DaySummary struct {
	Date             string                    `json:"date"`
	TotalOrders      int                       `json:"total_orders"`
	FinalizedOrders  int                       `json:"finalized_orders"`
	CancelledOrders  int                       `json:"cancelled_orders"`
	TotalRevenue     int64                     `json:"total_revenue"`
	ByPayment        map[PaymentMethod]int64   `json:"by_payment"`
	OrdersByPayment  map[PaymentMethod]int     `json:"orders_by_payment"`
	AverageTicket    int64                     `json:"average_ticket"`
}

func ComputeDaySummary(date string, orders []Order) DaySummary {
	s := DaySummary{
		Date:            date,
		ByPayment:       make(map[PaymentMethod]int64),
		OrdersByPayment: make(map[PaymentMethod]int),
	}

	for _, o := range orders {
		s.TotalOrders++
		switch o.Status {
		case StatusFinalizado:
			s.FinalizedOrders++
			s.TotalRevenue += o.Total()
			for _, p := range o.EffectivePayments() {
				s.ByPayment[p.Method] += p.Amount
				s.OrdersByPayment[p.Method]++
			}
		case StatusCancelado:
			s.CancelledOrders++
		}
	}

	if s.FinalizedOrders > 0 {
		s.AverageTicket = s.TotalRevenue / int64(s.FinalizedOrders)
	}

	return s
}

// FormatDateBR converts "2026-02-05" to "05/02/2026".
func FormatDateBR(isoDate string) string {
	t, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		return isoDate
	}
	return t.Format("02/01/2006")
}

type Menu struct {
	Items []MenuItem `json:"items"`
}

func NewMenu() *Menu {
	return &Menu{Items: []MenuItem{}}
}

func (m *Menu) Categories() []string {
	seen := map[string]bool{}
	var categories []string
	for _, item := range m.Items {
		if !item.Active {
			continue
		}
		if !seen[item.Category] {
			seen[item.Category] = true
			categories = append(categories, item.Category)
		}
	}
	return categories
}

func (m *Menu) ItemsByCategory(category string) []MenuItem {
	var items []MenuItem
	for _, item := range m.Items {
		if item.Category == category && item.Active {
			items = append(items, item)
		}
	}
	return items
}

func (m *Menu) AddItem(item MenuItem) {
	item.ID = m.NextID()
	item.Active = true
	m.Items = append(m.Items, item)
}

func (m *Menu) RemoveItem(id int) {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i].Active = false
			return
		}
	}
}

func (m *Menu) UpdateItem(updated MenuItem) {
	for i, item := range m.Items {
		if item.ID == updated.ID {
			m.Items[i] = updated
			return
		}
	}
}

func (m *Menu) NextID() int {
	maxID := 0
	for _, item := range m.Items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

// FormatBRL formats centavos as Brazilian Real: "R$ 1.234,56"
func FormatBRL(centavos int64) string {
	negative := centavos < 0
	if negative {
		centavos = -centavos
	}

	reais := centavos / 100
	cents := centavos % 100

	reaisStr := formatWithDotGrouping(reais)

	result := fmt.Sprintf("R$ %s,%02d", reaisStr, cents)
	if negative {
		result = "-" + result
	}
	return result
}

// FormatBRLPadded formats centavos as BRL right-aligned to the given width.
func FormatBRLPadded(centavos int64, width int) string {
	formatted := FormatBRL(centavos)
	if len(formatted) >= width {
		return formatted
	}
	return strings.Repeat(" ", width-len(formatted)) + formatted
}

func formatWithDotGrouping(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ".")
}
