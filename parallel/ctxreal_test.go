package parallel_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func server(ctx context.Context, l net.Listener, cancel context.CancelFunc) {
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Printf("全局上下文操作了Done命令进行关闭服务器操作 %v\n", ctx.Err())
			return
		}
	}(ctx)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("接收到一个错误的消息信息: %v\n", err)
			continue
		}
		go handler(ctx, conn, cancel)
	}
}

func handler(ctx context.Context, conn net.Conn, cancel context.CancelFunc) {
	defer conn.Close()
	msg, err := io.ReadAll(conn)
	if err != nil {
		fmt.Printf("[%s]消息接收失败: %v\n", conn.RemoteAddr().String(), err)
		return
	}
	fmt.Printf("[%s]接收到客户端消息\n%s\n", conn.RemoteAddr().String(), string(msg))
	if string(msg) == "close" {
		cancel()
	}
}

// socket server 基于context控制运行的启停使用
func TestSocketContext(t *testing.T) {
	l, err := net.Listen("tcp", ":8858")
	if err != nil {
		e := fmt.Sprintf("%v", err)
		panic("启动失败:" + e)
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go server(ctx, l, cancel)
	select {
	case <-ctx.Done():
		fmt.Println("程序退出", ctx.Err())
	}
}

// socket client 基于服务的调用操作
func TestSocketClient(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8858", 5*time.Second)
	if err != nil {
		e := fmt.Sprintf("%v", err)
		panic("启动失败:" + e)
	}
	defer conn.Close()

	conn.Write([]byte("close"))

}
