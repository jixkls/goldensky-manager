//go:build windows

package printer

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	winspool          = syscall.NewLazyDLL("winspool.drv")
	openPrinterW      = winspool.NewProc("OpenPrinterW")
	closePrinter      = winspool.NewProc("ClosePrinter")
	startDocPrinterW  = winspool.NewProc("StartDocPrinterW")
	endDocPrinter     = winspool.NewProc("EndDocPrinter")
	startPagePrinter  = winspool.NewProc("StartPagePrinter")
	endPagePrinter    = winspool.NewProc("EndPagePrinter")
	writePrinter      = winspool.NewProc("WritePrinter")
	enumPrintersW     = winspool.NewProc("EnumPrintersW")
	getPrinterW       = winspool.NewProc("GetPrinterW")
)

const (
	printerStatusError        = 0x00000002
	printerStatusOffline      = 0x00000080
	printerStatusNotAvailable = 0x00001000
	printerStatusMask         = printerStatusError | printerStatusOffline | printerStatusNotAvailable
)

// docInfo1 mirrors the Windows DOC_INFO_1W structure.
type docInfo1 struct {
	docName    *uint16
	outputFile *uint16
	datatype   *uint16
}

// printerInfo4 mirrors the Windows PRINTER_INFO_4W structure.
type printerInfo4 struct {
	printerName *uint16
	serverName  *uint16
	attributes  uint32
}

// SpoolerWriter sends raw data to a Windows printer via the Spooler API.
type SpoolerWriter struct {
	handle uintptr
}

func openSpooler(printerName string) (*SpoolerWriter, error) {
	namePtr, err := syscall.UTF16PtrFromString(printerName)
	if err != nil {
		return nil, fmt.Errorf("nome da impressora invalido: %w", err)
	}

	var handle uintptr
	r, _, e := openPrinterW.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&handle)),
		0,
	)
	if r == 0 {
		return nil, fmt.Errorf("abrir impressora %s: %w", printerName, e)
	}
	return &SpoolerWriter{handle: handle}, nil
}

func (sw *SpoolerWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	docName, _ := syscall.UTF16PtrFromString("GoldenSky POS")
	datatype, _ := syscall.UTF16PtrFromString("RAW")
	doc := docInfo1{
		docName:  docName,
		datatype: datatype,
	}

	r, _, e := startDocPrinterW.Call(sw.handle, 1, uintptr(unsafe.Pointer(&doc)))
	if r == 0 {
		return 0, fmt.Errorf("iniciar documento: %w", e)
	}

	r, _, e = startPagePrinter.Call(sw.handle)
	if r == 0 {
		endDocPrinter.Call(sw.handle)
		return 0, fmt.Errorf("iniciar pagina: %w", e)
	}

	var written uint32
	r, _, e = writePrinter.Call(
		sw.handle,
		uintptr(unsafe.Pointer(&p[0])),
		uintptr(len(p)),
		uintptr(unsafe.Pointer(&written)),
	)
	if r == 0 {
		endPagePrinter.Call(sw.handle)
		endDocPrinter.Call(sw.handle)
		return int(written), fmt.Errorf("escrever dados: %w", e)
	}

	endPagePrinter.Call(sw.handle)
	endDocPrinter.Call(sw.handle)

	return int(written), nil
}

func (sw *SpoolerWriter) Close() error {
	if sw.handle == 0 {
		return nil
	}
	r, _, e := closePrinter.Call(sw.handle)
	sw.handle = 0
	if r == 0 {
		return fmt.Errorf("fechar impressora: %w", e)
	}
	return nil
}

// printerStatus queries the spooler for the printer's status flags
// using GetPrinterW with PRINTER_INFO_6 (level 6), which returns
// only the dwStatus DWORD.
func printerStatus(handle uintptr) (uint32, error) {
	var needed uint32
	getPrinterW.Call(handle, 6, 0, 0, uintptr(unsafe.Pointer(&needed)))
	if needed == 0 {
		return 0, fmt.Errorf("nao foi possivel obter tamanho do buffer")
	}

	buf := make([]byte, needed)
	r, _, e := getPrinterW.Call(
		handle,
		6,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
	)
	if r == 0 {
		return 0, fmt.Errorf("consultar status da impressora: %w", e)
	}

	status := *(*uint32)(unsafe.Pointer(&buf[0]))
	return status, nil
}

// Open opens a connection to a Windows printer via the Spooler API.
// After opening the spooler handle, it queries status flags to detect
// offline or unavailable printers, since OpenPrinterW succeeds for any
// registered driver even without a physical device.
func Open(printerName string) (*Printer, error) {
	sw, err := openSpooler(printerName)
	if err != nil {
		return nil, err
	}

	status, err := printerStatus(sw.handle)
	if err != nil {
		sw.Close()
		return nil, fmt.Errorf("impressora nao respondeu em %s: %w", printerName, err)
	}
	if status&printerStatusMask != 0 {
		sw.Close()
		return nil, fmt.Errorf("impressora %s offline ou indisponivel (status: 0x%X)", printerName, status)
	}

	return &Printer{device: sw, path: printerName}, nil
}

const (
	printerEnumLocal = 0x00000002
)

// DetectPrinters returns names of locally installed printers on Windows.
func DetectPrinters() []string {
	var needed, count uint32

	// First call to get required buffer size.
	enumPrintersW.Call(
		printerEnumLocal,
		0,
		4, // level 4 = PRINTER_INFO_4
		0,
		0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&count)),
	)
	if needed == 0 {
		return nil
	}

	buf := make([]byte, needed)
	r, _, _ := enumPrintersW.Call(
		printerEnumLocal,
		0,
		4,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&count)),
	)
	if r == 0 {
		return nil
	}

	printers := make([]string, 0, count)
	infoSize := unsafe.Sizeof(printerInfo4{})
	for i := uint32(0); i < count; i++ {
		info := (*printerInfo4)(unsafe.Pointer(&buf[uintptr(i)*infoSize]))
		if info.printerName != nil {
			name := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(info.printerName))[:])
			printers = append(printers, name)
		}
	}
	return printers
}
