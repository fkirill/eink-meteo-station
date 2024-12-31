package clib

// #cgo LDFLAGS: -lexample -L.
// #include "DEV_Config.h"
import "C"

// void DEV_Digital_Write(UWORD Pin, UBYTE Value);
// UBYTE DEV_Digital_Read(UWORD Pin);
//
// void DEV_SPI_WriteByte(UBYTE Value);
// UBYTE DEV_SPI_ReadByte();
//
// void DEV_Delay_ms(UDOUBLE xms);
// void DEV_Delay_us(UDOUBLE xus);
//
// UBYTE DEV_Module_Init(void);
func DEV_Module_Init() uint8 {
	return uint8(C.DEV_Module_Init())
}

//void DEV_Module_Exit(void);
