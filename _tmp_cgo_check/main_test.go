package main

import "testing"

func TestCGO(t *testing.T) {
if cgoOK() == nil {
t.Fatal("nil")
}
}
