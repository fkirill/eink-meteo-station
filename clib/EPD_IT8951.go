package clib

// #cgo LDFLAGS: -lexample -L.
// #include "EPD_IT8951.h"
// typedef UBYTE* PUBYTE;
// typedef int8_t* PBYTE;
// typedef char* PCHAR;
// typedef void* PVOID;
import "C"
import (
	"fmt"
	"unsafe"
)

const INIT_Mode = 0
const A2_Mode = 6
const GC16_Mode = 2

const EPD_RST_PIN = 17
const EPD_CS_PIN = 8
const EPD_BUSY_PIN = 24

const HIGH = 0x1
const LOW = 0x0

const SYS_REG_BASE = 0x0000
const BLACK uint8 = 0x00
const WHITE uint8 = 0xFF

// Address of System Registers
const I80CPCR = (SYS_REG_BASE + 0x04)
const USDEF_I80_CMD_VCOM = 0x0039
const IT8951_TCON_SYS_RUN = 0x0001
const USDEF_I80_CMD_GET_DEV_INFO = 0x0302
const IT8951_TCON_REG_WR = 0x0011
const IT8951_LDIMG_L_ENDIAN = 0
const IT8951_2BPP = 0
const IT8951_3BPP = 1
const IT8951_4BPP = 2
const IT8951_8BPP = 3
const IT8951_ROTATE_0 = 0
const DISPLAY_REG_BASE = 0x1000            //Register RW access for I80 only
const LUTAFSR = (DISPLAY_REG_BASE + 0x224) //LUT Status Reg (status of All LUT Engines)
const IT8951_TCON_REG_RD = 0x0010
const MCSR_BASE_ADDR = 0x0200
const LISAR = (MCSR_BASE_ADDR + 0x0008)
const IT8951_TCON_LD_IMG_AREA = 0x0021
const IT8951_TCON_LD_IMG_END = 0x0022
const USDEF_I80_CMD_DPY_AREA = 0x0034
const USDEF_I80_CMD_DPY_BUF_AREA = 0x0037
const UP1SR = DISPLAY_REG_BASE + 0x138 //Update Parameter1 Setting Reg
const BGVR = DISPLAY_REG_BASE + 0x250  //Bitmap (1bpp) image color table

const (
	DOT_PIXEL_1X1 = 1 // 1 x 1
	DOT_PIXEL_2X2 = 2 // 2 X 2
	DOT_PIXEL_3X3 = 3 // 3 X 3
	DOT_PIXEL_4X4 = 4 // 4 X 4
	DOT_PIXEL_5X5 = 5 // 5 X 5
	DOT_PIXEL_6X6 = 6 // 6 X 6
	DOT_PIXEL_7X7 = 7 // 7 X 7
	DOT_PIXEL_8X8 = 8 // 8 X 8

	DRAW_FILL_EMPTY = 0
	DRAW_FILL_FULL  = 1

	MIRROR_NONE       = 0x00
	MIRROR_HORIZONTAL = 0x01
	MIRROR_VERTICAL   = 0x02
	MIRROR_ORIGIN     = 0x03

	ROTATE_0   = 0
	ROTATE_90  = 90
	ROTATE_180 = 180
	ROTATE_270 = 270

	LINE_STYLE_SOLID  = 0
	LINE_STYLE_DOTTED = 1

	DOT_FILL_AROUND  = 1 // dot pixel 1 x 1
	DOT_FILL_RIGHTUP = 2 // dot pixel 2 X 2

	DOT_STYLE_DFT = DOT_FILL_AROUND //Default dot pilex

	IMAGE_BACKGROUND = WHITE
	FONT_FOREGROUND  = BLACK
	FONT_BACKGROUND  = WHITE

	DOT_PIXEL_DFT = DOT_PIXEL_1X1 //Default dot pilex
	epd_mode      = 1             //1: no rotate, horizontal mirror, for 10.3inch
)

type IT8951_Dev_Info struct {
	Panel_W     uint16
	Panel_H     uint16
	Memory_Addr uint32
	FW_Version  [8]uint16
	LUT_Version [8]uint16
}

type IT8951_Load_Img_Info struct {
	Endian_Type        uint16 //little or Big Endian
	Pixel_Format       uint16 //bpp
	Rotate             uint16 //Rotate mode
	Source_Buffer      []byte //Start address of source Frame buffer
	Target_Memory_Addr uint32 //Base address of target image buffer
}

func (i *IT8951_Load_Img_Info) String() string {
	return fmt.Sprintf("Endian_Type = %v, Pixel_Format = %v, Rotate = %v, len(Source_Buffer) = %v, Target_Memory_Addr = %v",
		i.Endian_Type,
		i.Pixel_Format,
		i.Rotate,
		len(i.Source_Buffer),
		i.Target_Memory_Addr,
	)
}

type IT8951_Area_Img_Info struct {
	Area_X uint16
	Area_Y uint16
	Area_W uint16
	Area_H uint16
}

func (a *IT8951_Area_Img_Info) String() string {
	return fmt.Sprintf("x = %d, y = %d, w = %d, h = %d", a.Area_X, a.Area_Y, a.Area_W, a.Area_H)
}

type EInkScreen interface {
	GetScreenDimensions() (uint16, uint16)
	ClearScreen() error
	WriteScreenArea(x, y, w, h uint16, buf []uint8) error
	// len(buf) MUST be equal to w * h. Each byte represents a point.
	// The allowed colors are: 0x00, 0x11, 0x22, ..., 0xee, 0xff.
	WriteScreenAreaRefreshMode(x, y, w, h uint16, buf []uint8, mode uint8) error
	GetMemoryAddr() uint32
}

type eInkScreen struct {
	deviceW, deviceH uint16
	softRefreshMode  uint8
	memoryAddr       uint32
}

// void Enhance_Driving_Capability(void);
//
// void EPD_IT8951_SystemRun(void);
//
// void EPD_IT8951_Standby(void);
//
// void EPD_IT8951_Sleep(void);
//
// IT8951_Dev_Info EPD_IT8951_Init(UWORD VCOM);
func EPD_IT8951_Init(VCOM uint16) *IT8951_Dev_Info {
	devInfo := C.EPD_IT8951_Init(C.UWORD(VCOM))
	var fwVersion, lutVersion [8]uint16
	res := &IT8951_Dev_Info{
		Panel_W:     uint16(devInfo.Panel_W),
		Panel_H:     uint16(devInfo.Panel_H),
		Memory_Addr: uint32(devInfo.Memory_Addr_H)<<16 | uint32(devInfo.Memory_Addr_L),
	}
	for i := 0; i < 8; i++ {
		fwVersion[i] = uint16(devInfo.FW_Version[i])
		lutVersion[i] = uint16(devInfo.LUT_Version[i])
	}
	res.FW_Version = fwVersion
	res.LUT_Version = lutVersion
	return res
}

// void EPD_IT8951_Clear_Refresh(IT8951_Dev_Info Dev_Info,UDOUBLE Target_Memory_Addr, UWORD Mode);
func EPD_IT8951_Clear_Refresh(screenW, screenH uint16, Target_Memory_Addr uint32, Mode uint8) {
	var devInfo C.IT8951_Dev_Info
	devInfo.Panel_W = C.UWORD(screenW)
	devInfo.Panel_H = C.UWORD(screenH)

	C.EPD_IT8951_Clear_Refresh(devInfo, C.UDOUBLE(Target_Memory_Addr), C.UWORD(Mode))
}

// void EPD_IT8951_1bp_Refresh(UBYTE* Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H, UBYTE Mode, UDOUBLE Target_Memory_Addr, bool Packed_Write);
func EPD_IT8951_1bp_Refresh(Frame_Buf []uint8, X, Y, W, H uint16, Mode uint8, Target_Memory_Addr uint32, Packed_Write bool) {
	img := unsafe.Pointer(unsafe.SliceData(Frame_Buf))
	C.EPD_IT8951_1bp_Refresh(
		C.PUBYTE(img),
		C.UWORD(X),
		C.UWORD(Y),
		C.UWORD(W),
		C.UWORD(H),
		C.UBYTE(Mode),
		C.UDOUBLE(Target_Memory_Addr),
		C.bool(Packed_Write),
	)
}

// void EPD_IT8951_1bp_Multi_Frame_Write(UBYTE* Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H,UDOUBLE Target_Memory_Addr, bool Packed_Write);
// void EPD_IT8951_1bp_Multi_Frame_Refresh(UWORD X, UWORD Y, UWORD W, UWORD H,UDOUBLE Target_Memory_Addr);
//
// void EPD_IT8951_2bp_Refresh(UBYTE* Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H, bool Hold, UDOUBLE Target_Memory_Addr, bool Packed_Write);
//
// void EPD_IT8951_4bp_Refresh(UBYTE* Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H, bool Hold, UDOUBLE Target_Memory_Addr, bool Packed_Write);
func EPD_IT8951_4bp_Refresh(Frame_Buf []uint8, X, Y, W, H uint16, Hold bool, Target_Memory_Addr uint32, Packed_Write bool) {
	img := unsafe.Pointer(unsafe.SliceData(Frame_Buf))
	C.EPD_IT8951_4bp_Refresh(
		C.PUBYTE(img),
		C.UWORD(X),
		C.UWORD(Y),
		C.UWORD(W),
		C.UWORD(H),
		C.bool(Hold),
		C.UDOUBLE(Target_Memory_Addr),
		C.bool(Packed_Write),
	)
}

// void EPD_IT8951_4bp_Refresh_Mode(UBYTE* Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H, bool Hold, UDOUBLE Target_Memory_Addr, bool Packed_Write, UBYTE Mode)
func EPD_IT8951_4bp_Refresh_Mode(Frame_Buf []uint8, X, Y, W, H uint16, Hold bool, Target_Memory_Addr uint32, Packed_Write bool, Mode uint8) {
	img := unsafe.Pointer(unsafe.SliceData(Frame_Buf))
	C.EPD_IT8951_4bp_Refresh_Mode(
		C.PUBYTE(img),
		C.UWORD(X),
		C.UWORD(Y),
		C.UWORD(W),
		C.UWORD(H),
		C.bool(Hold),
		C.UDOUBLE(Target_Memory_Addr),
		C.bool(Packed_Write),
		C.UBYTE(Mode),
	)
}

//
//void EPD_IT8951_8bp_Refresh(UBYTE *Frame_Buf, UWORD X, UWORD Y, UWORD W, UWORD H, bool Hold, UDOUBLE Target_Memory_Addr);

func Epd_Mode(mode uint8) {
	if mode == 3 {
		Paint_SetRotate(ROTATE_0)
		Paint_SetMirroring(MIRROR_NONE)
		//isColor = 1
	} else if mode == 2 {
		Paint_SetRotate(ROTATE_0)
		Paint_SetMirroring(MIRROR_HORIZONTAL)
	} else if mode == 1 {
		Paint_SetRotate(ROTATE_0)
		Paint_SetMirroring(MIRROR_HORIZONTAL)
	} else {
		Paint_SetRotate(ROTATE_0)
		Paint_SetMirroring(MIRROR_NONE)
	}
}
