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
		in := scanner.Text()
		s, err := ppstern.ParseAndFormat([]byte(in))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", s)
	}
}
