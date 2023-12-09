package main

import (
	"github.com/airenas/listgo/internal/app/manager"
	"github.com/labstack/gommon/color"
)

func main() {
	printBanner()
	manager.Execute()
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
    __  ___            ___              ____ 
   /  |/  /___ _____  /   | ____ ____  / __ \
  / /|_/ / __ ` + "`" + `/ __ \/ /| |/ __ ` + "`" + `/ _ \/ /_/ /
 / /  / / /_/ / / / / ___ / /_/ /  __/ _, _/ 
/_/  /_/\__,_/_/ /_/_/  |_\__, /\___/_/ |_|  v: %s
                         /____/ 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("github.com/airenas/listgo"))
}
