package main

import (
	"github.com/airenas/listgo/internal/app/cmdworker"
	"github.com/labstack/gommon/color"
)

func main() {
	printBanner()
	cmdworker.Execute()
}

var (
	version string
)

func printBanner() {
	banner := `
    ___      __ 
   / (_)____/ /_
  / / / ___/ __/
 / / (__  ) /_  
/_/_/____/\__/  
         ____        __ __             
        / __ \      / //_/    |
       / / / /     / ,<       | 
      / /_/ /     / /| |      | 
  w   \____/  r  /_/ |_|   er | v: %s

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("github.com/airenas/listgo"))
}
