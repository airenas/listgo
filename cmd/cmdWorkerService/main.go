package main

import (
	"bitbucket.org/airenas/listgo/internal/app/cmdworker"
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
                
              ____             __ __             
 _      __   / __ \   _____   / //_/   ___  _____
| | /| / /  / / / /  / ___/  / ,<     / _ \/ ___/
| |/ |/ /  / /_/ /  / /     / /| |   /  __/ /    
|__/|__/   \____/  /_/     /_/ |_|   \___/_/    v: %s


%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("bitbucket.org/airenas/listgo"))
}
