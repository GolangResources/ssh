package gssh

import (
	"log"
	"net"
	"fmt"
	"os"
	"io"

	"context"

	"strings"
	"encoding/binary"

	"golang.org/x/crypto/ssh"

	"github.com/armon/go-socks5"

)

var serverConn *ssh.Client
var ipcache map[string]uint32
var ipcount uint32

type SSHClient struct {
	Debug		bool
	//ipcount		uint32
	//ipcache		map[string]uint32
	fakeDNS		socks5.NameResolver
	Client		*ssh.Client
}

func Init(sp *SSHClient) SSHClient {
	var s SSHClient
	//s.ipcount = 1
	//s.ipcache = make(map[string]uint32)

	if sp != nil {
		s = *sp
	} else {
		s = SSHClient{
			Debug: false,
		}
	}
	return s
}
//-L 
func (s *SSHClient) TCPTunnel(local string, remote string) error {
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.forward(conn, remote)
	}
}

func (s *SSHClient) forward(localConn net.Conn, remote string) {
	serverConn = s.Client
	remoteConn, err := serverConn.Dial("tcp", remote)
	if err != nil {
		fmt.Printf("Remote dial error: %s\n", err)
		return
	}

	copyConn:=func(writer, reader net.Conn) {
		_, err:= io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
			return
		}
	}
	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}
//-R 
func (s *SSHClient) RTCPTunnel(remote string, local string) error {
	serverConn = s.Client
	listener, err := serverConn.Listen("tcp", remote)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.rforward(conn, local)
	}
}

func (s *SSHClient) rforward(remoteConn net.Conn, local string) {
	localConn, err := net.Dial("tcp", local)
	if err != nil {
		fmt.Printf("Local dial error: %s\n", err)
		return
	}

	copyConn:=func(writer, reader net.Conn) {
		_, err:= io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
			return
		}
	}
	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

//-D
func sshDial(ctx context.Context, a string, b string) (net.Conn, error) {
	bb := strings.Split(b, ":")	
	ip := bb[0]
	port := bb[1]
	dns := "127.0.0.1"
	for k, v := range ipcache {
		ipstr := int2ip(v).String()
		if (ipstr == ip) {
			dns = k
		}
	}
	b = dns + ":" + port
	return serverConn.Dial(a, b)
}

type DNSResolver struct{}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func (d DNSResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	ip := int2ip(0)
	if (ipcache[name] == 0) {
		ipcache[name] = ipcount
		ip = int2ip(ipcount)
	} else {
		ip = int2ip(ipcache[name])
	}
	return ctx, ip, nil
}

func (s *SSHClient) Dynamic(listen string) { 
	s.fakeDNS = DNSResolver{} 
	serverConn = s.Client
	ipcache = make(map[string]uint32)
	ipcount = 1

	conf := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Dial: sshDial, 
		Resolver: s.fakeDNS,
	}

	server, err := socks5.New(conf)
	if err != nil {
	  panic(err)
	}

	if err := server.ListenAndServe("tcp", listen); err != nil {
	  panic(err)
	}
}
