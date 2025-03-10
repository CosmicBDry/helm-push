package tools

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

// 创建一个ssh的远程连接client
func CreateSshClient(username, passord, remoteAddr string) (*ssh.Client, error) {
	config := ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(passord),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", remoteAddr, &config)
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial() Error: %s", err.Error())
	}

	return client, nil
}

// 基于session发送文件至远程机器指定目录下
func SendFileToRemote(local_path, remote_path string, client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("client.NewSession() Error: %s", err.Error())
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("session.StdinPipe() Error: %s", err.Error())
	}

	defer stdin.Close()
	cmd := fmt.Sprintf("cat > %s", remote_path)
	err = session.Start(cmd)
	if err != nil {
		return fmt.Errorf("session.Start(cmd) Error: %s", err.Error())
	}
	osFile, err := os.Open(local_path)
	if err != nil {
		return fmt.Errorf("os.Open(local_path) Error: %s", err.Error())
	}
	_, err = io.Copy(stdin, osFile)

	if err != nil {
		return fmt.Errorf("io.Copy(stdin, osFile) Error: %s", err.Error())
	}
	// err = session.Wait()
	// if err != nil {
	// 	return fmt.Errorf("session.Wait() Error: %s", err.Error())
	// }
	defer session.Close()
	return nil
}

// 执行本地命令
func LocalCommand(localCmd string) (string, error) {
	cmd := exec.Command("bash", "-c", localCmd)

	outBytes, err := cmd.CombinedOutput()

	//err = cmd.Run()
	if err != nil {
		return "", err
	}

	return string(outBytes), nil
}

// 基于session执行远程命令
func RemoteCommand(remoteCmd string, client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("client.NewSession() Error: %s", err.Error())
	}
	err = session.Start(remoteCmd)
	if err != nil {
		return fmt.Errorf("session.Start(remoteCmd) Error: %s", err.Error())
	}
	// err = session.Wait()
	// if err != nil {
	// 	return fmt.Errorf("session.Wait() Error: %s", err.Error())
	// }
	defer session.Close()

	return err
}
