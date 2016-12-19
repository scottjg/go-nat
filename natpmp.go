package nat

import (
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/jackpal/gateway"
	"github.com/jackpal/go-nat-pmp"
)

var (
	_ NAT = (*natpmpNAT)(nil)
)

func discoverNATPMP() <-chan NAT {
	res := make(chan NAT, 1)

	ip, err := gateway.DiscoverGateway()
	if err == nil {
		go discoverNATPMPWithAddr(res, ip)
	}

	return res
}

func discoverNATPMPWithAddr(c chan NAT, ip net.IP) {
	client := natpmp.NewClient(ip)
	_, err := client.GetExternalAddress()
	if err != nil {
		return
	}

	c <- &natpmpNAT{client, ip}
}

type natpmpNAT struct {
	c       *natpmp.Client
	gateway net.IP
}

func (n *natpmpNAT) GetDeviceAddress() (addr net.IP, err error) {
	return n.gateway, nil
}

func (n *natpmpNAT) GetInternalAddress() (addr net.IP, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			switch x := addr.(type) {
			case *net.IPNet:
				if x.Contains(n.gateway) {
					return x.IP, nil
				}
			}
		}
	}

	return nil, ErrNoInternalAddress
}

func (n *natpmpNAT) GetExternalAddress() (addr net.IP, err error) {
	res, err := n.c.GetExternalAddress()
	if err != nil {
		return nil, err
	}

	d := res.ExternalIPAddress
	return net.IPv4(d[0], d[1], d[2], d[3]), nil
}

func (n *natpmpNAT) AddPortMapping(protocol string, internalPort int, externalPort int, description string, timeout time.Duration) error {
	var (
		err error
	)

	timeoutInSeconds := int(timeout / time.Second)

	result, err := n.c.AddPortMapping(protocol, internalPort, externalPort, timeoutInSeconds)
	if result.MappedExternalPort != uint16(externalPort) {
		n.c.AddPortMapping(protocol, internalPort, 0, 0)
		return errors.New("external port " + strconv.Itoa(externalPort) + " already in use")
	}
	return err
}

func (n *natpmpNAT) DeletePortMapping(protocol string, internalPort int, externalPort int) (err error) {
	_, err = n.c.AddPortMapping(protocol, internalPort, 0, 0)
	return err
}

func (n *natpmpNAT) Type() string {
	return "NAT-PMP"
}
