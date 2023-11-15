package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/toga4/ppstern/ppstern"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		in := scanner.Bytes()
		s, err := ppstern.ParseAndFormat(in)
		if err != nil {
			panic(err)
		}
		fmt.Println(s)
	}
}
