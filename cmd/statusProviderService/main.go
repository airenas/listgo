package main

import (
	"github.com/airenas/listgo/internal/app/status"
	"github.com/labstack/gommon/color"
)

func main() {
	printBanner()
	status.Execute()
}

var version string

func printBanner() {
	banner := `
    ___      __ 
   / (_)____/ /_
  / / / ___/ __/
 / / (__  ) /_  
/_/_/____/\__/  
         __  ___  ______          
   _____/ /_/   |/_  __/_  _______
  / ___/ __/ /| | / / / / / / ___/
 (__  ) /_/ ___ |/ / / /_/ (__  ) 
/____/\__/_/  |_/_/  \__,_/____/  | v: %s

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("github.com/airenas/listgo"))
}
