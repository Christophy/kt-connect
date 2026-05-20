package cluster

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HeartBeatStatus struct {
	status map[string]bool
	sync.RWMutex
}

func (h *HeartBeatStatus) Get(key string) (value bool, exists bool) {
	h.RLock()
	value, exists = h.status[key]
	h.RUnlock()
	return
}

func (h *HeartBeatStatus) Set(key string, value bool) {
	h.Lock()
	h.status[key] = value
	h.Unlock()
	return
}

// LastHeartBeatStatus record last heart beat status to avoid verbose log
var LastHeartBeatStatus = &HeartBeatStatus{
	status: map[string]bool{},
}

// SetupTimeDifference get time difference between cluster and local
func SetupTimeDifference() error {
	rectifierPodName := fmt.Sprintf("%s%s", util.RectifierPodPrefix, strings.ToLower(util.RandomString(5)))
	_, err := Ins().CreateRectifierPod(rectifierPodName)
	if err != nil {
		return err
	}
	stdout, stderr, err := Ins().ExecInPod(util.DefaultContainer, rectifierPodName, opt.Get().Global.Namespace, "date", "+%s")
	if err != nil {
		return err
	}
	go func() {
		if err2 := Ins().RemovePod(rectifierPodName, opt.Get().Global.Namespace); err2 != nil {
			log.Debug().Err(err).Msgf("Failed to remove pod %s", rectifierPodName)
		}
	}()
	remoteTime, err := strconv.ParseInt(stdout, 10, 0)
	if err != nil {
		log.Warn().Msgf("Invalid cluster time: '%s' %s", stdout, stderr)
		return err
	}
	timeDifference := remoteTime - time.Now().Unix()
	if timeDifference >= -1 && timeDifference <= 1 {
		log.Debug().Msgf("No time difference")
	} else {
		log.Debug().Msgf("Time difference is %d", timeDifference)
	}
	util.TimeDifference = timeDifference
	return nil
}

// SetupHeartBeat setup heartbeat watcher
func SetupHeartBeat(name, namespace string, updater func(string, string)) {
	ticker := time.NewTicker(time.Minute*util.ResourceHeartBeatIntervalMinus - util.RandomSeconds(0, 10))
	go func() {
		for range ticker.C {
			updater(name, namespace)
		}
	}()
}

// SetupPortForwardHeartBeat setup heartbeat watcher for port forward.
// For SSH port (remotePort=22), reads banner to verify SPDY tunnel is alive.
// For other ports, falls back to TCP connect only.
func SetupPortForwardHeartBeat(port int, remotePort int, stop chan struct{}) *time.Ticker {
	ticker := time.NewTicker(util.PortForwardHeartBeatIntervalSec*time.Second - util.RandomSeconds(0, 5))
	isSSH := remotePort == 22
	go func() {
		consecutiveFailures := 0
		for range ticker.C {
			ok := false
			if isSSH {
				ok = sshDataProbe(port)
			} else {
				if conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), 5*time.Second); err == nil {
					_ = conn.Close()
					ok = true
				}
			}
			if ok {
				consecutiveFailures = 0
				log.Debug().Msgf("Heartbeat port forward %d ticked at %s", port, util.FormattedTime())
			} else {
				consecutiveFailures++
				log.Warn().Msgf("Heartbeat port forward %d probe failed (%d/3)", port, consecutiveFailures)
				if isSSH && consecutiveFailures >= 3 {
					log.Warn().Msgf("Port forward %d SPDY dead, forcing reconnect", port)
					close(stop)
					ticker.Stop()
					return
				}
			}
		}
	}()
	return ticker
}

// sshDataProbe connects to the local port-forward port and reads the SSH banner.
// If data arrives, the SPDY tunnel is alive. If timeout, SPDY is dead.
func sshDataProbe(port int) bool {
	result := make(chan bool, 1)
	go func() {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), 5*time.Second)
		if err != nil {
			result <- false
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 32)
		n, _ := conn.Read(buf)
		result <- n > 0
	}()
	select {
	case r := <-result:
		return r
	case <-time.After(12 * time.Second):
		return false
	}
}

func resourceHeartbeatPatch() string {
	return fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		util.KtLastHeartBeat, util.GetTimestamp())
}
