package shared

import (
	"fmt"
	"os"
	"golang.org/x/crypto/ssh/terminal"
)
func ReadPasswordFromTerminal(c *Config)(passwd string, err error){
	fmt.Printf("%s@%s's password: ", c.Username, c.Hostname)
	p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return
	}
	passwd = string(p)
	fmt.Println()
	return
}


