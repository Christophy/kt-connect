package transmission

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// SetupPortForwardToLocal mapping local port to shadow pod ssh port
func SetupPortForwardToLocal(podName string, remotePort, localPort int) (chan int, error) {
	gone := make(chan int)
	return gone, setupPortForwardToLocal(podName, remotePort, localPort, gone, true)
}

func setupPortForwardToLocal(podName string, remotePort, localPort int, gone chan int, isInitConnect bool) error {
	initReady := make(chan struct{})
	stop := make(chan struct{})
	var ticker *time.Ticker
	go func() {
		first := true
		for {
			ready := initReady
			if !first {
				ready = make(chan struct{})
			}
			fw, err := createPortForwarder(podName, remotePort, localPort, stop, ready)
			if err != nil {
				log.Warn().Err(err).Msgf("Port forward create failed, retrying in 10s")
				time.Sleep(10 * time.Second)
				stop = make(chan struct{})
				continue
			}
			fwDone := make(chan error, 1)
			go func() {
				fwDone <- fw.ForwardPorts()
			}()
			if !first {
				select {
				case <-ready:
					ticker = cluster.SetupPortForwardHeartBeat(localPort, remotePort, stop)
					log.Info().Msgf("Port forward local:%d -> pod %s:%d re-established", localPort, podName, remotePort)
				case err := <-fwDone:
					log.Debug().Err(err).Msgf("Port forward exited before ready")
					time.Sleep(10 * time.Second)
					stop = make(chan struct{})
					continue
				case <-time.After(time.Duration(opt.Get().Global.PortForwardTimeout) * time.Second):
					log.Warn().Msgf("Port forward reconnect timeout, retrying in 10s")
					time.Sleep(10 * time.Second)
					stop = make(chan struct{})
					continue
				}
			}
			first = false
			err = <-fwDone
			if err != nil {
				log.Debug().Err(err).Msgf("Port forward local:%d -> pod %s:%d interrupted", localPort, podName, remotePort)
			}
			if ticker != nil {
				ticker.Stop()
			}
			time.Sleep(10 * time.Second)
			log.Debug().Msgf("Port forward reconnecting ...")
			stop = make(chan struct{})
		}
	}()

	select {
	case <-initReady:
		ticker = cluster.SetupPortForwardHeartBeat(localPort, remotePort, stop)
		log.Info().Msgf("Port forward local:%d -> pod %s:%d established", localPort, podName, remotePort)
		return nil
	case <-time.After(time.Duration(opt.Get().Global.PortForwardTimeout) * time.Second):
		return fmt.Errorf("connect to port-forward failed")
	}
}

// createPortForwarder fetch a port forward handler
func createPortForwarder(podName string, remotePort, localPort int, stop, ready chan struct{}) (*portforward.PortForwarder, error) {
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", opt.Get().Global.Namespace, podName)
	log.Debug().Msgf("Request port forward pod:%d -> local:%d via %s", remotePort, localPort, opt.Store.RestConfig.Host)
	apiUrl, err := parseReqHost(opt.Store.RestConfig.Host, apiPath)
	if err != nil {
		return nil, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(opt.Store.RestConfig)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	return portforward.New(dialer, ports, stop, ready, util.BackgroundLogger, util.BackgroundLogger)
}

// parseReqHost get the final url to port forward api
func parseReqHost(host, apiPath string) (*url.URL, error) {
	pos := strings.Index(host, "://")
	if pos < 0 {
		return nil, fmt.Errorf("invalid host address: %s", host)
	}
	protocol := host[0:pos]
	hostIP := host[pos+3:]
	baseUrl := ""
	pos = strings.Index(hostIP, "/")
	if pos > 0 {
		baseUrl = hostIP[pos:]
		hostIP = hostIP[0:pos]
	}
	fullPath := path.Join(baseUrl, apiPath)
	return &url.URL{Scheme: protocol, Host: hostIP, Path: fullPath}, nil
}
