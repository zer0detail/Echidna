package echidna

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	Web     *bool
	Plugins *bool
	Themes  *bool
	Help    *bool
}


// not used anymore. Unless I decide to expand to using a web server or add themes scanning too
func (o *options) Parse() {

	o.Web = flag.Bool("w", false, "Enable web server on port 8080 to display results")
	o.Plugins = flag.Bool("p", false, "Scan WordPress Plugins")
	o.Themes = flag.Bool("t", false, "Scan WordPress Themes")
	o.Help = flag.Bool("h", false, "Displays this help message")

	flag.Parse()

	if *o.Help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !(*o.Plugins || *o.Themes) {
		fmt.Println("No scan target was selected. Use either -p or -t to select a target")
		flag.PrintDefaults()
		os.Exit(0)
	}

}
