package main

import (
	"bitbucket.org/airenas/listgo/internal/app/zoom"
	"github.com/labstack/gommon/color"
)

func main() {
	printBanner()
	zoom.Execute()
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
                          ____             
 ____           ____     / __ \   ____ ___ 
/_  /          / __ \   / / / /  / __ `+ "`" +`__ \
 / /_    _    / /_/ /  / /_/ /  / / / / / /
/___/   (_)   \____/   \____/  /_/ /_/ /_/   v: %s

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("bitbucket.org/airenas/listgo"))
}

