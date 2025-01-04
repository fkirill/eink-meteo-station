package eink

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"image"
	"math"
)

type EInkScreen interface {
	GetScreenDimensions() (uint16, uint16)
	GetBufferAddress() uint32
	WriteScreenAreaRefreshMode(area image.Rectangle, raster []byte, mode uint8) error
}

type einkScreen struct {
	buffer         []uint8
	panelW, panelH uint16
	bufferAddress  uint32
	fwVersion      string
	lutVersion     string
}

func (e *einkScreen) WriteScreenAreaRefreshMode(area image.Rectangle, raster []byte, mode uint8) error {
	clib.EPD_IT8951_4bp_Refresh_Mode(
		raster,
		uint16(area.Min.X),
		uint16(area.Min.Y),
		uint16(area.Dx()),
		uint16(area.Dy()),
		false,
		e.bufferAddress,
		false,
		mode,
	)
	return nil
}

func (e *einkScreen) GetScreenDimensions() (uint16, uint16) {
	return e.panelW, e.panelH
}

func (e *einkScreen) GetBufferAddress() uint32 {
	return e.bufferAddress
}

func NewEInkScreen(vcom float64) (EInkScreen, error) {
	if clib.DEV_Module_Init() != 0 {
		panic("Failed to initialize eink screen")
	}
	vcomInt := uint16(math.Abs(vcom) * 1000)
	Dev_Info := clib.EPD_IT8951_Init(vcomInt)
	clib.Epd_Mode(1)
	clib.EPD_IT8951_Clear_Refresh(Dev_Info.Panel_W, Dev_Info.Panel_H, Dev_Info.Memory_Addr, clib.INIT_Mode)
	bufSize := Dev_Info.Panel_W * Dev_Info.Panel_H / 2
	return &einkScreen{
		buffer:        make([]uint8, bufSize, bufSize),
		panelW:        Dev_Info.Panel_W,
		panelH:        Dev_Info.Panel_H,
		bufferAddress: Dev_Info.Memory_Addr,
		fwVersion:     wordsToString(Dev_Info.FW_Version[:]),
		lutVersion:    wordsToString(Dev_Info.LUT_Version[:]),
	}, nil
}

func wordsToString(words []uint16) string {
	buffer := bytes.Buffer{}
	for _, w := range words {
		buffer.WriteByte(byte(w << 8))
		buffer.WriteByte(byte(w & 0xff))
	}
	buf := buffer.Bytes()
	var l int
	for i := range buf {
		if buf[i] == 0 {
			l = i
			break
		}
	}
	if l == 0 {
		return ""
	}
	return string(buf[0:l])
}
