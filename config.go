package main

type Configuration struct {
	Soft      bool
	Verbose   bool
	NoConfirm bool

	ConfigFile   string
	DatabaseName string
	DatabaseDir  string
	DatabasePath string

	Packages []string
}
