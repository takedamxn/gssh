package shared
import (
	"golang.org/x/crypto/ssh"
	"fmt"
	"log"
	"net"
)
func Connect(c *Config) (client *ssh.Client, err error){
	// Create client config
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: func(Hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := fmt.Sprintf("%s:%d", c.Hostname, c.Port)
	client, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf("ssh.Dial : %s", err)
		return
	}
	return
}
