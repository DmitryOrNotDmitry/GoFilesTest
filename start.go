package main

import (
	"example_ftest/ftest"
	"fmt"
	"io"
)

func Sum(a, b int) int {
	return a + b
}

func main() {
	ftest.RunTests("sumTests", "input*.txt", Process, true, true)
	//fmt.Print(Process(os.Stdin))
}

func Process(readerIn io.Reader) string {
	var a, b int
	fmt.Fscan(readerIn, &a)
	fmt.Fscan(readerIn, &b)

	actual := fmt.Sprintln(Sum(a, b))
	return actual
}
