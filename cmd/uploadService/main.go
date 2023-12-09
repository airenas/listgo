package main

import (
	"github.com/airenas/listgo/internal/app/upload"
	"github.com/labstack/gommon/color"
)

func main() {
	printBanner()
	upload.Execute()
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
               __                __
  __  ______  / /___  ____ _____/ /
 / / / / __ \/ / __ \/ __ ` + "`" + `/ __  / 
/ /_/ / /_/ / / /_/ / /_/ / /_/ /  
\__,_/ .___/_/\____/\__,_/\__,_/ v: %s  
    /_/ 	
%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("github.com/airenas/listgo"))
}
