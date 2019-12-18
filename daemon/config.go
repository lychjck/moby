package daemon

import (
	"net"

	"github.com/docker/docker/daemon/networkdriver"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
)

const (
	defaultNetworkMtu    = 1500
	DisableNetworkBridge = "none"
)

// Config define the configuration of a docker daemon
// These are the configuration settings that you pass
// to the docker daemon when you launch it with say: `docker -d -e lxc`
// FIXME: separate runtime configuration from http api configuration
type Config struct {
	Pidfile                     string   ////Docker Daemon 所属进程的 PID 文件
	Root                        string  //Docker 运行时所使用的 root 路径
	AutoRestart                 bool	// 已被启用，转而支持 docker run 时的重启
	Dns                         []string	//Docker 使用的 DNS Server 地址
	DnsSearch                   []string	//Docker 使用的指定的 DNS 查找域名
	EnableIptables              bool	// 启用 Docker 的 iptables 功能
	EnableIpForward             bool	// 启用 net.ipv4.ip_forward 功能
	DefaultIp                   net.IP	// 绑定容器端口时使用的默认 IP
	BridgeIface                 string	// 添加容器网络至已有的网桥 
	BridgeIP                    string	// 添加容器网络至已有的网桥 
	InterContainerCommunication bool	// 是否允许相同 host 上容器间的通信
	GraphDriver                 string	 //Docker 运行时使用的特定存储驱动 
	GraphOptions                []string	 // 可设置的存储驱动选项
	ExecDriver                  string	 // Docker 运行时使用的特定 exec 驱动
	Mtu                         int		// 设置容器网络的 MTU
	DisableNetwork              bool	// 有定义，之后未初始化
	EnableSelinuxSupport        bool	// 启用 SELinux 功能的支持
	Context                     map[string][]string	// 有定义，之后未初始化
}

// InstallFlags adds command-line options to the top-level flag parser for
// the current process.
// Subsequent calls to `flag.Parse` will populate config with values parsed
// from the command-line.
/* 定义一个为 String 类型的 flag 参数；
该 flag 的名称为”p”或者”-pidfile”;
该 flag 的值为” /var/run/docker.pid”, 并将该值绑定在变量 config.Pidfile 上；
该 flag 的描述信息为"Path to use for daemon PID file"。 */
func (config *Config) InstallFlags() {
	flag.StringVar(&config.Pidfile, []string{"p", "-pidfile"}, "/var/run/docker.pid", "Path to use for daemon PID file")
	flag.StringVar(&config.Root, []string{"g", "-graph"}, "/var/lib/docker", "Path to use as the root of the Docker runtime")
	flag.BoolVar(&config.AutoRestart, []string{"#r", "#-restart"}, true, "--restart on the daemon has been deprecated infavor of --restart policies on docker run")
	flag.BoolVar(&config.EnableIptables, []string{"#iptables", "-iptables"}, true, "Enable Docker's addition of iptables rules")
	flag.BoolVar(&config.EnableIpForward, []string{"#ip-forward", "-ip-forward"}, true, "Enable net.ipv4.ip_forward")
	flag.StringVar(&config.BridgeIP, []string{"#bip", "-bip"}, "", "Use this CIDR notation address for the network bridge's IP, not compatible with -b")
	flag.StringVar(&config.BridgeIface, []string{"b", "-bridge"}, "", "Attach containers to a pre-existing network bridge\nuse 'none' to disable container networking")
	flag.BoolVar(&config.InterContainerCommunication, []string{"#icc", "-icc"}, true, "Enable inter-container communication")
	flag.StringVar(&config.GraphDriver, []string{"s", "-storage-driver"}, "", "Force the Docker runtime to use a specific storage driver")
	flag.StringVar(&config.ExecDriver, []string{"e", "-exec-driver"}, "native", "Force the Docker runtime to use a specific exec driver")
	flag.BoolVar(&config.EnableSelinuxSupport, []string{"-selinux-enabled"}, false, "Enable selinux support. SELinux does not presently support the BTRFS storage driver")
	flag.IntVar(&config.Mtu, []string{"#mtu", "-mtu"}, 0, "Set the containers network MTU\nif no value is provided: default to the default route MTU or 1500 if no default route is available")
	opts.IPVar(&config.DefaultIp, []string{"#ip", "-ip"}, "0.0.0.0", "Default IP address to use when binding container ports")
	opts.ListVar(&config.GraphOptions, []string{"-storage-opt"}, "Set storage driver options")
	// FIXME: why the inconsistency between "hosts" and "sockets"?
	opts.IPListVar(&config.Dns, []string{"#dns", "-dns"}, "Force Docker to use specific DNS servers")
	opts.DnsSearchListVar(&config.DnsSearch, []string{"-dns-search"}, "Force Docker to use specific DNS search domains")
}

func GetDefaultNetworkMtu() int {
	if iface, err := networkdriver.GetDefaultRouteIface(); err == nil {
		return iface.MTU
	}
	return defaultNetworkMtu
}
