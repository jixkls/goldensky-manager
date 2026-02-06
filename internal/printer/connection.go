package printer

import (
	"fmt"
	"io"
	"sync"
)

// Printer manages the connection to a thermal printer.
type Printer struct {
	device io.WriteCloser
	path   string
	mu     sync.Mutex
}

// Close closes the printer connection.
func (p *Printer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.device != nil {
		err := p.device.Close()
		p.device = nil
		return err
	}
	return nil
}

// Write sends raw bytes to the printer.
func (p *Printer) Write(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.device == nil {
		return fmt.Errorf("impressora nao conectada")
	}
	_, err := p.device.Write(data)
	return err
}

// Print builds the receipt and sends it to the printer.
func (p *Printer) Print(rb *ReceiptBuilder) error {
	return p.Write(rb.Build())
}

// PrintTest sends a test page to the printer.
func (p *Printer) PrintTest() error {
	rb := NewReceiptBuilder()
	rb.AlignCenter().
		FontDouble().Bold().
		Line("TESTE DE IMPRESSAO").
		FontNormal().NoBold().
		Line("GoldenSky GS-T80E").
		Separator('-', 48).
		AlignLeft().
		Line("Caracteres especiais:").
		Line("aeiou AEIOU").
		Line("acucar cafe pao nao").
		Separator('-', 48).
		AlignCenter().
		Line("Impressora funcionando!").
		Feed(4).
		PartialCut()
	return p.Print(rb)
}

// OpenDrawer sends the command to open the cash drawer.
func (p *Printer) OpenDrawer() error {
	rb := NewReceiptBuilder()
	rb.OpenDrawer()
	return p.Print(rb)
}

// IsConnected returns true if the printer device is open.
func (p *Printer) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.device != nil
}

// Path returns the device path of this printer.
func (p *Printer) Path() string {
	return p.path
}
