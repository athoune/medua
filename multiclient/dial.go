package multiclient

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	})
	fmt.Println("addr", addr)
	fmt.Println("ip", ctx.Value("ip"))
	fmt.Println("host", ctx.Value("host"))

	slug := strings.Split(addr, ":")
	ip := ctx.Value("ip")
	if ip != nil {
		port, _ := strconv.Atoi(slug[1])
		return net.DialTCP(network, nil, &net.TCPAddr{
			IP:   ctx.Value("ip").(net.IP),
			Port: port,
		})
	}
	return dialer.Dial(network, addr)
}
