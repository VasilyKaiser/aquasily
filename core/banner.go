package core

import "fmt"

// Basic info
var (
	Name = fmt.Sprintf(`
                                             d8b 888          
                                             Y8P 888          
                                                 888          
 8888b.   .d88888 888  888  8888b.  .d8888b  888 888 888  888 
    "88b d88" 888 888  888     "88b 88K      888 888 888  888 
.d888888 888  888 888  888 .d888888 "Y8888b. 888 888 888  888 
888  888 Y88b 888 Y88b 888 888  888      X88 888 888 Y88b 888 
"Y888888  "Y88888  "Y88888 "Y888888  88888P' 888 888  "Y88888 
              888                                         888 
              888                                    Y8b d88P 
              888              %s                  "Y88P"
`, Version)
	Version = "v1.0.2"
	Author  = "Vasily Kaiser"
	Website = "https://github.com/VasilyKaiser/aquasily"
)
