package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/dockerversion"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/reexec"
	"github.com/docker/docker/utils"
)

const (
	defaultCaFile   = "ca.pem"
	defaultKeyFile  = "key.pem"
	defaultCertFile = "cert.pem"
)

func main() {
	if reexec.Init() {
		return
	}
	flag.Parse()
	// FIXME: validate daemon flags here

	if *flVersion {
		showVersion()
		return
	}
	if *flDebug {
		os.Setenv("DEBUG", "1")
	}
	//flHosts 的作用是为 Docker Client 提供所要连接的 host 对象，也为 Docker Server 提供所要监听的对象。
	if len(flHosts) == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")
		if defaultHost == "" || *flDaemon {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", api.DEFAULTUNIXSOCKET)
		}
		if _, err := api.ValidateHost(defaultHost); err != nil {
			log.Fatal(err)
		}
		flHosts = append(flHosts, defaultHost)
	}

	if *flDaemon {
		mainDaemon()
		return
	}

	if len(flHosts) > 1 {
		log.Fatal("Please specify only one -H")
	}
	protoAddrParts := strings.SplitN(flHosts[0], "://", 2)

	var (
		cli       *client.DockerCli
		//TlsConfig 对象的创建是为了保障 cli 在传输数据的时候，遵循安全传输层协议 (TLS)。安全传输层协议 (TLS) 用于两个通信应用程序之间保密性与数据完整性。
		tlsConfig tls.Config
	)
	tlsConfig.InsecureSkipVerify = true
	//若 flTlsVerify 这个 flag 参数为真的话，则说明需要验证 server 端的安全性，tlsConfig 对象需要加载一个受信的 ca 文件。
	//该 ca 文件的路径为 *flCA 参数的值，最终完成 tlsConfig 对象中 RootCAs 属性的赋值，并将 InsecureSkipVerify 属性置为假。
	// If we should verify the server, we need to load a trusted ca
	if *flTlsVerify {
		*flTls = true
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(*flCa)
		if err != nil {
			log.Fatalf("Couldn't read ca cert %s: %s", *flCa, err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.RootCAs = certPool
		tlsConfig.InsecureSkipVerify = false
	}

	// If tls is enabled, try to load and send client certificates
	if *flTls || *flTlsVerify {
		_, errCert := os.Stat(*flCert)
		_, errKey := os.Stat(*flKey)
		if errCert == nil && errKey == nil {
			*flTls = true
			cert, err := tls.LoadX509KeyPair(*flCert, *flKey)
			if err != nil {
				log.Fatalf("Couldn't load X509 key pair: %s. Key encrypted?", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}
	/*
	如果 flTls 和 flTlsVerify 两个 flag 参数中有一个为真，则说明需要加载以及发送 client 端的证书。最终将证书内容交给 tlsConfig 的 Certificates 属性。
	至此，flag 参数已经全部处理，并已经收集完毕 Docker Client 所需的配置信息。之后的内容为 Docker Client 如何实现创建并执行。	
	*/
	if *flTls || *flTlsVerify {
		cli = client.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, protoAddrParts[0], protoAddrParts[1], &tlsConfig)
	} else {
		cli = client.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, protoAddrParts[0], protoAddrParts[1], nil)
	}

	if err := cli.Cmd(flag.Args()...); err != nil {
		if sterr, ok := err.(*utils.StatusError); ok {
			if sterr.Status != "" {
				log.Println(sterr.Status)
			}
			os.Exit(sterr.StatusCode)
		}
		log.Fatal(err)
	}
}

func showVersion() {
	fmt.Printf("Docker version %s, build %s\n", dockerversion.VERSION, dockerversion.GITCOMMIT)
}
