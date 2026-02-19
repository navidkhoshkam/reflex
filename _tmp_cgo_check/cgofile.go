package main

/*
#include <stdlib.h>
*/
import "C"

func cgoOK() *C.char {
return C.CString("ok")
}
