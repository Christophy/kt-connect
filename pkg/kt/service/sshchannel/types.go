package sshchannel

import (
	"sync"

	"github.com/wzshiming/sshproxy"
)

// Channel network channel
type Channel interface {
	StartSocks5Proxy(privateKey, sshAddress, socks5Address string) error
	ForwardRemoteToLocal(privateKey, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(privateKey, sshAddress, script string) (string, error)
}

// Cli the singleton type
type Cli struct {
	currentDialer *sshproxy.Dialer
	sshAddr       string
	mu            sync.Mutex
}
var instance *Cli

// Ins get singleton instance
func Ins() Channel {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}
