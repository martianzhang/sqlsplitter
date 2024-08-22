package main

import (
	"fmt"
	"os"

	"github.com/martianzhang/sqlsplitter"
)

func main() {
	s, err := sqlsplitter.New(os.Args[1], sqlsplitter.Default)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var i uint64
	for s.Next() {
		sql := s.Scan()
		i++
		fmt.Println(i, sql)
	}
	if s.Error != nil {
		fmt.Println(s.Error)
	}
}
