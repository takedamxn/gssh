package main

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"os"
	"os/user"
	"regexp"
	"strings"
)

var sec *ini.Section
var passwords = make(map[string]string)

func readPasswords() (err error) {
	if len(configPath) == 0 {
		configPath = os.Getenv("GSSH_PASSWORDFILE")
		if len(configPath) == 0 {
			usr, err := user.Current()
			if err == nil {
				f := usr.HomeDir + "/.gssh"
				_, err = os.Stat(f)
				if os.IsNotExist(err) == false {
					configPath = f
				}
			}
		}
	}
	if len(configPath) != 0 {
		cfg, err := ini.InsensitiveLoad(configPath)
		if err != nil {
			return err
		}
		sec, err = cfg.GetSection("passwords")
		if err != nil {
			return err
		}
		passwords = sec.KeysHash()
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
			passwords[group[1]] = group[2]
		}
	}
	return
}
func getPassword(u, h string, port int) string {
	target := ""
	// search password with user@hostname[:port]
	if port != 22 {
		target = fmt.Sprintf("%s@%s:%d", u, h, port)
	} else {
		target = fmt.Sprintf("%s@%s", u, h)
	}
	if password, ok := passwords[target]; ok {
		return password
	}
	// search password with user
	if password, ok := passwords[u]; ok {
		return password
	}
	return ""
}
