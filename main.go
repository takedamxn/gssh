package main

import (
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	com "github.com/takemxn/gssh/common"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	command         string
	passEnv         bool
	tFlag           bool
	vFlag           bool
	hFlag           bool
	username string
	hostname string
	configPath string
	port int
	password string
	timeout int
)
const(
	version = "1.0.1"
)
type Session struct {
	*ssh.Session
}

func main() {
	err := parseArg()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if password == "" {
		config := com.NewConfig(configPath)
		err = config.ReadPasswords()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		password = config.GetPassword(username, hostname, port)
		if password == "" {
			password, err = com.ReadPasswordFromTerminal(username, hostname)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
	ci := com.NewConnectInfo(username, hostname, port, password)
	// Create a session
	conn, err := ci.Connect()
	if err != nil {
		log.Printf("unable to create session: %s", err)
		return
	}
	defer conn.Close()
	ses, err := conn.NewSession()
	if err != nil {
		log.Printf("unable to create session: %s", err)
		return
	}
	defer ses.Close()
	s := Session{ses}
	exitStatus := 0
	if(timeout > 0){
		go func(){
			time.Sleep(time.Duration(timeout) * time.Second)
			exitStatus = 124
			s.Signal(ssh.SIGTERM)
			s.Close()
		}()
	}
	// Terminal file descpriter?
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		err = s.remoteShell()
	} else {
		err =s.remoteExec()
	}
	// check remote process exit status when exitStatus not set
	if err != nil && exitStatus == 0{
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		} else {
			log.Printf("ExitError:%v", err)
			os.Exit(1)
		}
	}
	os.Exit(exitStatus)
}
func parseArg() (err error) {
	args := os.Args
	f := flag.NewFlagSet(args[0], flag.ContinueOnError)
	f.StringVar(&password, "p", "", "password")
	f.StringVar(&configPath, "f", "", "password file path")
	f.BoolVar(&passEnv, "e", false, "passing to pty")
	f.BoolVar(&tFlag, "t", false, "Force pseudo-tty allocation")
	f.BoolVar(&vFlag, "v", false, "Display Version")
	f.IntVar(&timeout, "w", 0, "Timeout")
	f.BoolVar(&hFlag, "h", false, "help")
	if err = f.Parse(args[1:]); err != nil {
		return
	}
	if vFlag {
		fmt.Println(path.Base(os.Args[0]), version)
		os.Exit(0)
	}
	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-t] [-p password] [-f config_path] [-v] [user@]hostname[:port] [command]\n", path.Base(os.Args[0]));
		f.PrintDefaults()
	}
	if f.NArg() <= 0 {
		usage()
		return fmt.Errorf("too few argument")
	}

	// Get user name
	rest := f.Arg(0)
	if strings.Contains(f.Arg(0), "@") {
		s := strings.Split(f.Arg(0), "@")
		if len(s[0]) == 0 {
			return fmt.Errorf("user name error")
		}
		username = s[0]
		rest = s[1]
	} else if username == "" {
		u, _ := user.Current()
		username = u.Username
	}

	// Get hostname
	s := strings.Split(rest, ":")
	if len(s[0]) == 0 {
		return fmt.Errorf("hostname error")
	}
	hostname = s[0]

	// Get port number
	if len(s) >= 2 {
		port, err = strconv.Atoi(s[1])
	}
	// Connect to ssh server
	if port == 0 {
		port = 22
	}
	// command
	command = strings.Join(f.Args()[1:], " ")

	return
}

func (s *Session) requestWindowChange(w, h int) (err error) {
	// RFC 4254 Section 6.7.
	req := struct {
		Columns uint32
		Rows    uint32
		Width   uint32
		Height  uint32
	}{
		Columns: uint32(w),
		Rows:    uint32(h),
		Width:   uint32(w * 8),
		Height:  uint32(h * 8),
	}
	ok, err := s.SendRequest("window-change", false, ssh.Marshal(&req))
	if err == nil && !ok {
		err = errors.New("ssh: window-change failed")
	}
	return err
}
func (s *Session) requestEnv(name, value string) (err error) {
	// RFC 4254 Section 6.4.
	req := struct {
		name  string
		value string
	}{
		name:  name,
		value: value,
	}
	fmt.Printf("name:[%s], value[%s]", name, value)
	ok, err := s.SendRequest("env", false, ssh.Marshal(&req))
	if err == nil && !ok {
		err = errors.New("ssh: env failed")
	}
	return err
}

func (s *Session) passEnvPty() (err error) {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		err = s.requestEnv(pair[0], pair[1])
		if err != nil {
			break
		}
	}
	return
}

func (s *Session) remoteShell() (err error) {
	exit := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}

	stdin, _ := s.StdinPipe()
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()

	go io.Copy(stdin, os.Stdin)

	go func() {
		wg.Add(1)
		defer wg.Done()
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
	}()

	if command != "" {
		if tFlag {
			// Forced request pseudo terminal
			w, h, err := terminal.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				log.Printf("request for terminal window size failed: %s", err)
				return err
			}
			// Set up terminal modes
			modes := ssh.TerminalModes{
				ssh.ECHO: 0,
			}
			if err := s.RequestPty("xterm", h, w, modes); err != nil {
				log.Printf("request for pseudo terminal failed: %s", err)
				return err
			}
		}
		// Exec Command
		if err = s.Start(command); err != nil {
			log.Printf("failed to start shell: %s", err)
			return err
		}
	} else {
		// For pipe
		fd := int(os.Stdin.Fd())

		// Make terminal into raw mode
		oldState, _ := terminal.MakeRaw(int(os.Stdin.Fd()))
		defer terminal.Restore(fd, oldState)

		// Request pseudo terminal
		w, h, err := terminal.GetSize(fd)
		if err != nil {
			log.Printf("request for terminal window size failed: %s", err)
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)
		go func() {
			for {
				select {
				case v := <-sig:
					switch v {
					case syscall.SIGINT, syscall.SIGTERM:
						return
					case syscall.SIGWINCH:
						if w, h, err := terminal.GetSize(fd); err == nil {
							s.requestWindowChange(w, h)
						}
					}
				case <-exit:
					return
				}
			}
		}()

		// Set up terminal modes
		modes := ssh.TerminalModes{
			ssh.ECHO: 1,
		}
		if err := s.RequestPty("xterm", h, w, modes); err != nil {
			log.Printf("request for pseudo terminal failed: %s", err)
			return err
		}
		// Start Remote Shell
		if err = s.Shell(); err != nil {
			log.Printf("failed to start shell: %s", err)
			return err
		}
	}
	err = s.Wait()
	wg.Wait()
	close(exit)
	return
}

func (s *Session) remoteExec() (err error) {
	stdin, _ := s.StdinPipe()
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()

	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		io.Copy(stdin, os.Stdin)
		stdin.Close()
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
	}()

	if tFlag {
		// Forced request pseudo terminal

		// Set up terminal modes
		modes := ssh.TerminalModes{
			ssh.ECHO: 0,
		}
		if err := s.RequestPty("xterm", 80, 40, modes); err != nil {
			log.Printf("request for pseudo terminal failed: %s", err)
			return err
		}
	}
	if command != "" {
		// Exec Command
		if err = s.Start(command); err != nil {
			log.Printf("failed to start shell: %s", err)
			return err
		}
	} else {
		// Start Remote Shell
		if err = s.Shell(); err != nil {
			log.Printf("failed to start shell: %s", err)
			return err
		}
	}

	err = s.Wait()
	wg.Wait()
	return
}
