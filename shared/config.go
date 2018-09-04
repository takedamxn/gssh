package shared

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"os"
	"os/user"
	"regexp"
	"strings"
)
type Config struct {
	Username   string
	Hostname   string
	Port       int
	ConfigPath string
	Password   string
	passwords map[string]string
	sec *ini.Section
}

func NewConfig(user, host string,port int, path, passwd string) *Config{
	return &Config{Username:user, Hostname:host, Port:port, ConfigPath:path, Password:passwd}
}
func (c *Config) ReadPasswords() (err error) {
	if len(c.ConfigPath) == 0 {
		c.ConfigPath = os.Getenv("GSSH_PASSWORDFILE")
		if len(c.ConfigPath) == 0 {
			usr, err := user.Current()
			if err == nil {
				f := usr.HomeDir + "/.gssh"
				_, err = os.Stat(f)
				if os.IsNotExist(err) == false {
					c.ConfigPath = f
				}
			}
		}
	}
	if len(c.ConfigPath) != 0 {
		cfg, err := ini.InsensitiveLoad(c.ConfigPath)
		if err != nil {
			return err
		}
		c.sec, err = cfg.GetSection("passwords")
		if err != nil {
			return err
		}
		c.passwords = c.sec.KeysHash()
		return err
	}
	env := os.Getenv("GSSH_PASSWORDS")
	if len(env) != 0 {
		re := regexp.MustCompile("(.+)=(.+)")
		for _, v := range strings.Split(env, " ") {
			group := re.FindStringSubmatch(v)
			if group == nil {
				return errors.New("$GSSH_PASSWORDS illeagal format")
			}
			c.passwords[group[1]] = group[2]
		}
	}
	return
}
func (c *Config) GetPassword(u, h string, port int) string {
	target := ""
	// search password with user@hostname[:port]
	if port != 22 {
		target = fmt.Sprintf("%s@%s:%d", u, h, port)
	} else {
		target = fmt.Sprintf("%s@%s", u, h)
	}
	if password, ok := c.passwords[target]; ok {
		return password
	}
	// search password with user
	if password, ok := c.passwords[u]; ok {
		return password
	}
	return ""
}
