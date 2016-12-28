package shell

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

type Credentials struct {
	IP      string
	User    string
	KeyName string
}

func NewClient(keyDirectory string, cred *Credentials) (*ssh.Client, error) {
	keyPath := path.Join(keyDirectory, cred.KeyName)
	privateKey, err := ioutil.ReadFile(keyPath)
	if os.IsNotExist(err) {
		privateKey, err = ioutil.ReadFile(keyPath + ".pem")
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: cred.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	return ssh.Dial("tcp", cred.IP+":22", config)
}

func InteractiveTerminal(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stderr, stderr)

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}