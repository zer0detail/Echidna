package echidna

var opts struct {
	Web     bool `short:"w" long:"web" description:"Start Local web server to interact with Echidna"`
	Plugins bool `short:"p" long:"plugins" description:"Scan WordPress Plugins"`
	Themes  bool `short:"t" long:"themes" description:"Scan WordPress Themes"`
	Help    bool `short:"h" long:"help" description:"Display help"`
}