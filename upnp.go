package nat

import (
	"net"
	"time"

	"github.com/scottjg/upnp"
)

var (
	_ NAT = (*upnp_NAT)(nil)
)

func discoverUPNP() <-chan NAT {
	res := make(chan NAT, 1)
	go func() {
		client := new(upnp.Upnp)

		err := client.SearchGateway()
		if err != nil {
			return
		}

		res <- &upnp_NAT{client, "UPNP"}
	}()

	return res
}

type upnp_NAT_Client interface {
	GetExternalIPAddress() (string, error)
	AddPortMapping(string, uint16, string, uint16, string, bool, string, uint32) error
	DeletePortMapping(string, uint16, string) error
}

type upnp_NAT struct {
	c   *upnp.Upnp
	typ string
}

func (u *upnp_NAT) GetExternalAddress() (addr net.IP, err error) {
	err = u.c.ExternalIPAddr()
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(u.c.GatewayOutsideIP)
	if ip == nil {
		return nil, ErrNoExternalAddress
	}

	return ip, nil
}

func mapProtocol(s string) string {
	switch s {
	case "udp":
		return "UDP"
	case "tcp":
		return "TCP"
	default:
		panic("invalid protocol: " + s)
	}
}

func (u *upnp_NAT) AddPortMapping(protocol string, internalPort int, externalPort int, description string, timeout time.Duration) error {
	timeoutInSeconds := int(timeout / time.Second)
	return u.c.AddPortMapping(internalPort, externalPort, timeoutInSeconds, mapProtocol(protocol), description)
}

func (u *upnp_NAT) DeletePortMapping(protocol string, internalPort int, externalPort int) error {
	u.c.DelPortMapping(externalPort, mapProtocol(protocol))
	return nil
}

func (u *upnp_NAT) GetDeviceAddress() (net.IP, error) {
	ip := net.ParseIP(u.c.Gateway.Host)
	if ip == nil {
		return nil, ErrNoInternalAddress
	}

	return ip, nil
}

func (u *upnp_NAT) GetInternalAddress() (net.IP, error) {
	ip := net.ParseIP(u.c.LocalHost)
	if ip == nil {
		return nil, ErrNoInternalAddress
	}

	return ip, nil
}

func (n *upnp_NAT) Type() string { return n.typ }
