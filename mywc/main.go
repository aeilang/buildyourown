package main

import (
	"os"
)

func main() {
	wc, close := NewMyWc()
	defer close()

	wc.SetDefaultCmd()
	wc.WriteTo(os.Stdout)
}





