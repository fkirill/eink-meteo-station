package clib

// #cgo LDFLAGS: -lexample -lbcm2835 -lm -L.
// #include "GUI_Paint.h"
// typedef UBYTE* PUBYTE;
// typedef sFONT* PsFONT;
import "C"
import (
	"fmt"
	"unsafe"
)

// void Paint_NewImage(UBYTE *image, UWORD Width, UWORD Height, UWORD Rotate, UWORD Color);
func Paint_NewImage(image []uint8, width, height uint16, rotate, color uint8) {
	img := unsafe.Pointer(unsafe.SliceData(image))
	C.Paint_NewImage(C.PUBYTE(img), C.UWORD(width), C.UWORD(height), C.UWORD(rotate), C.UWORD(color))
}

// void Paint_SelectImage(UBYTE *image);
func Paint_SelectImage(image []uint8) {
	img := unsafe.Pointer(unsafe.SliceData(image))
	C.Paint_SelectImage(C.PUBYTE(img))
}

// void Paint_SetRotate(UWORD Rotate);
func Paint_SetRotate(Rotate uint8) {
	C.Paint_SetRotate(C.UWORD(Rotate))
}

// void Paint_SetMirroring(UBYTE mirror);
func Paint_SetMirroring(mirror uint8) {
	C.Paint_SetMirroring(C.UBYTE(mirror))
}

// void Paint_SetBitsPerPixel(UBYTE bpp);
func Paint_SetBitsPerPixel(bpp uint8) {
	C.Paint_SetBitsPerPixel(C.UBYTE(bpp))
}

// void Paint_SetPixel(UWORD Xpoint, UWORD Ypoint, UWORD Color);
//
// void Paint_Clear(UWORD Color);
func Paint_Clear(color uint8) {
	C.Paint_Clear(C.UWORD(color))
}

// void Paint_ClearWindows(UWORD Xstart, UWORD Ystart, UWORD Xend, UWORD Yend, UWORD Color);
//
// //Drawing
// void Paint_DrawPoint(UWORD Xpoint, UWORD Ypoint, UWORD Color, DOT_PIXEL Dot_Pixel, DOT_STYLE Dot_FillWay);
// void Paint_DrawLine(UWORD Xstart, UWORD Ystart, UWORD Xend, UWORD Yend, UWORD Color, DOT_PIXEL Line_width, LINE_STYLE Line_Style);
// void Paint_DrawRectangle(UWORD Xstart, UWORD Ystart, UWORD Xend, UWORD Yend, UWORD Color, DOT_PIXEL Line_width, DRAW_FILL Draw_Fill);
func Paint_DrawRectangle(Xstart, Ystart, Xend, Yend uint16, Color, Line_width, Draw_Fill uint8) {
	C.Paint_DrawRectangle(
		C.UWORD(Xstart),
		C.UWORD(Ystart),
		C.UWORD(Xend),
		C.UWORD(Yend),
		C.UWORD(Color),
		C.DOT_PIXEL(Line_width),
		C.DRAW_FILL(Draw_Fill),
	)
}

// void Paint_DrawCircle(UWORD X_Center, UWORD Y_Center, UWORD Radius, UWORD Color, DOT_PIXEL Line_width, DRAW_FILL Draw_Fill);
func Paint_DrawCircle(X_Center, Y_Center, Radius uint16, Color, Line_width, Draw_Fill uint8) {
	C.Paint_DrawCircle(
		C.UWORD(X_Center),
		C.UWORD(Y_Center),
		C.UWORD(Radius),
		C.UWORD(Color),
		C.DOT_PIXEL(Line_width),
		C.DRAW_FILL(Draw_Fill),
	)
}

// //Display string
// void Paint_DrawChar(UWORD Xstart, UWORD Ystart, const char Acsii_Char, sFONT* Font, UWORD Color_Foreground, UWORD Color_Background);
// void Paint_DrawString_EN(UWORD Xstart, UWORD Ystart, const char * pString, sFONT* Font, UWORD Color_Foreground, UWORD Color_Background);
// void Paint_DrawString_CN(UWORD Xstart, UWORD Ystart, const char * pString, cFONT* font, UWORD Color_Foreground, UWORD Color_Background);
// void Paint_DrawNum(UWORD Xpoint, UWORD Ypoint, int32_t Nummber, sFONT* Font, UWORD Color_Foreground, UWORD Color_Background);
func Paint_DrawNum(Xpoint, Ypoint uint16, Number int, FontName string, Color_Foreground, Color_Background uint8) {
	font := getFontByName(FontName)
	C.Paint_DrawNum(
		C.UWORD(Xpoint),
		C.UWORD(Ypoint),
		C.int32_t(Number),
		C.PsFONT(font),
		C.UWORD(Color_Foreground),
		C.UWORD(Color_Background),
	)
}

//void Paint_DrawTime(UWORD Xstart, UWORD Ystart, PAINT_TIME *pTime, sFONT* Font, UWORD Color_Foreground, UWORD Color_Background);
//
//void Paint_SetColor(UWORD x, UWORD y, UWORD color);
//void Paint_GetColor(UWORD color, UBYTE* arr_color);

func getFontByName(name string) unsafe.Pointer {
	switch name {
	case "Font24":
		return unsafe.Pointer(&C.Font24)
	case "Font20":
		return unsafe.Pointer(&C.Font20)
	case "Font16":
		return unsafe.Pointer(&C.Font16)
	case "Font12":
		return unsafe.Pointer(&C.Font12)
	case "Font8":
		return unsafe.Pointer(&C.Font8)
	case "Font12CN":
		return unsafe.Pointer(&C.Font12CN)
	case "Font24CN":
		return unsafe.Pointer(&C.Font24CN)
	default:
		panic(fmt.Errorf("Font %s is not found", name))
	}
}
