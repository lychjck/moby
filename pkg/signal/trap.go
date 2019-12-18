package signal

import (
	"log"
	"os"
	gosignal "os/signal"
	"sync/atomic"
	"syscall"
)

// Trap sets up a simplified signal "trap", appropriate for common
// behavior expected from a vanilla unix command-line tool in general
// (and the Docker engine in particular).
//
// * If SIGINT or SIGTERM are received, `cleanup` is called, then the process is terminated.
// * If SIGINT or SIGTERM are repeated 3 times before cleanup is complete, then cleanup is
// skipped and the process terminated directly.
// * If "DEBUG" is set in the environment, SIGQUIT causes an exit without cleanup.
//


/* 创建并设置一个 channel，用于发送信号通知；
定义 signals 数组变量，初始值为 os.SIGINT, os.SIGTERM; 若环境变量 DEBUG 为空的话，则添加 os.SIGQUIT 至 signals 数组；
通过 gosignal.Notify(c, signals...) 中 Notify 函数来实现将接收到的 signal 信号传递给 c。
需要注意的是只有 signals 中被罗列出的信号才会被传递给 c，其余信号会被直接忽略；
创建一个 goroutine 来处理具体的 signal 信号，当信号类型为 os.Interrupt 或者 syscall.SIGTERM 时，执行传入 Trap 函数的具体执行方法，
形参为 cleanup(), 实参为 eng.Shutdown。 */
func Trap(cleanup func()) {
	c := make(chan os.Signal, 1)
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	if os.Getenv("DEBUG") == "" {
		signals = append(signals, syscall.SIGQUIT)
	}
	gosignal.Notify(c, signals...)
	go func() {
		interruptCount := uint32(0)
		for sig := range c {
			go func(sig os.Signal) {
				log.Printf("Received signal '%v', starting shutdown of docker...\n", sig)
				switch sig {
				case os.Interrupt, syscall.SIGTERM:
					// If the user really wants to interrupt, let him do so.
					if atomic.LoadUint32(&interruptCount) < 3 {
						atomic.AddUint32(&interruptCount, 1)
						// Initiate the cleanup only once
						if atomic.LoadUint32(&interruptCount) == 1 {
							// Call cleanup handler
							cleanup()
							os.Exit(0)
						} else {
							return
						}
					} else {
						log.Printf("Force shutdown of docker, interrupting cleanup\n")
					}
				case syscall.SIGQUIT:
				}
				os.Exit(128 + int(sig.(syscall.Signal)))
			}(sig)
		}
	}()
}
