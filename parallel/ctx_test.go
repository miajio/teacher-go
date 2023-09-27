package parallel_test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ctx 是基于golang的context.Context使用实例
// 这个文件将告诉你golang的context.Context是什么,有什么用,该如何用

// context.Context是go语言并发编程的核心部分,该接口在go 1.7版本时引入标准库之中,其定义了4个需要实现的方法
// Dealine: 返回context.Context被取消的时间,即完成工作的截止时间
// Done: 返回一个Channel(通道),这个Channel会在当前工作完成或上下文被取消后关闭,多次调用Done方法会返回同一个Channel
// Err: 返回context.Context结束的原因,它只会在Done方法对应的Channel关闭时返回非空值
//
//	如果context.Context被取消,会返回Canceled错误
//	如果context.Context超时,会返回DealineExceeded错误
//
// Value: 从context.Context中获取键对应的值,对于同一个上下文而言,多次调用Value并传入相同的Key会返回相同的结果,该方法可以用于传递特定数据
func TestContextInterface(t *testing.T) {
	// 该方法说明了context.Context这个接口
	// 或者你也可以直接阅读源码去看context.Context
	type Context interface {
		Dealine() (deadline time.Time, ok bool)
		Done() <-chan struct{}
		Err() error
		Value(key any) any
	}
}

// context包中提供的context.Background,context.TODO,context.WithDeadline和context.WithValue函数会返回实现context.Context接口的私有结构体
// context.Context最大的作用时在goroutine构成的树形结构中同步信号以减少计算资源的浪费,go语言服务的每一个请求都是通过单独的goroutine处理的,HTTP/RPC请求处理器会启动新的goroutine访问数据库和其它服务

//                      --> Goroutine --> Goroutine
// Context -> Goroutine --> Goroutine
//                      --> Goroutine

// 如上方注释,我们可能会创建多个goroutine来处理一次请求,而context.Context的作用时在不同的goroutine之间同步请求特定数据,取消信号以及处理请求的截止时间
// 每一个context.Context都会从最顶层的goroutine逐层传递到最底层.context.Context可以在上层goroutine执行出现错误时将信号及时同步到下层
// 如果错误的使用context.Context去进行操作,goroutine操作会在当最上层的goroutine因为某些原因执行失败时,下层的goroutine由于没有接收到这个信号而继续执行,那么会消耗掉多余的无用资源
// 而正确的使用context.Context去进行操作,那么在上层goroutine因为某些原因执行失败时,下层的goroutine将接收到这个信号从而停止执行,就可以及时的停掉无用的工作而得以减少额外的资源消耗

func TestRunContext01(t *testing.T) {
	// 因为过期时间大于处理时间,所以我们有足够的时间处理该请求,运行该测试代码会打印出以下内容

	// 使用处理请求 500ms
	// 主程序 context deadline exceeded

	// goroutine函数没有进入超时的select分支,但是主程序的select会等待context.Context超时并打印 主程序 context deadline exceeded
	// 如果将goroutine的函数中时间变更为2s,整个程序都会因为上下文过期而终止并打印一下内容

	// 操作Done context deadline exceeded
	// 主程序 context deadline exceeded

	ctx, canel := context.WithTimeout(context.Background(), 1*time.Second)
	defer canel()

	go func(ctx context.Context, duration time.Duration) {
		select {
		case <-ctx.Done():
			fmt.Println("操作Done", ctx.Err())
		case <-time.After(duration):
			fmt.Println("使用处理请求", duration)
		}
	}(ctx, 500*time.Millisecond)
	// }(ctx, 2*time.Second)

	select {
	case <-ctx.Done():
		fmt.Println("主程序", ctx.Err())
	}
}

// 依据上方测试方法的例子,应该能够帮助各位理解context.Context的使用方法和设计原理,多个goroutine同时订阅ctx.Done() channel中的消息,一旦接收到取消信号,就立刻停止当前正在执行的工作

// 默认上下文
// context包中最常用的方法时context.Background和context.TODO,这两个方法都会返回预先初始化号的私有变量background和todo,它们会在同一个go程序中被复用

// Background和TODO都是通过new(emptyCtx)语句初始化的,它们是指向私有结构体context.emptyCtx的指针,这是最简单,最常用的上下文类型

/*
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}
*/

// 上述代码中不难发现,context.emptyCtx通过空方法实现了context.Context接口中的所有方法,它没有任何功能,从源代码来看,context.Background和context.TODO只是互为别名,没有太大的差别,只是使用语义上稍有不同
// background 是上下文的默认值,其它所有上下文都应该从它衍生出来
// todo 应该是仅在不确定应该使用哪种上下文时使用

// 在多数情况下,如果当前函数没有上下文做为入参,我们会使用context.Background做为起始上下文向下传递

// 取消信号
// context.WithCancel函数能够从context.Context中衍生出新的子上下文,并返回用于取消该上下文的函数.一旦我们执行返回的取消函数,当前上下文及其子上下文都会被取消,所有的goroutine都会同步收到这一取消信号
// 详见context源码中WithCancel函数

// context.withCancel将传入上下文封装成私有结构体context.cancelCtx
// context.cancelCtx.propagateCancel会构建父子上下文之间的关联,当父上下文被取消时,子上下文也会被取消

// 除context.WithCancel外,context包中的另外两个函数context.WithTimeout和context.WithDeadline也都能创建可以被取消的计时器上下文context.timerCtx

// context.WithDealine在创建context.timerCtx的过程中判断了父上下文的截止时间与当前时间,并通过time.AfterFunc创建定时器,当时间超过截止日期后会调用context.timerCtx.cancel同步取消信号
// context.timerCtx内部不仅通过嵌入context.cancelCtx结构体继承了相关变量和方法,还通过持有定时器timer和截止时间deadline实现了定时取消功能
// context.timerCtx.cancel方法不仅调用了context.cancelCtx.cancel,还会停止持有的定时器减少不必要的资源浪费

// 传值方法
// 最后需要了解如何使用上下文传值了,context包中的context.WithValue能从上下文中创建子上下文,传递的子上下文使用context.valueCtx类型
// context.valueCtx结构体将除Value外的Err,Deadline等方法代理到父上下文中,它只会影响context.valueCtx.Value方法
// 如果context.valueCtx中存储的键值对与context.valueCtx.Value方法中传入的参数不匹配,就会从父上下文中查找该键值对的值,直到某个父上下文中返回nil或者查到对应的值

// 总结
// Go语言中的context.Context的主要作用是在多个goroutine组成的树中同步取消信号以减少对资源的消耗和占用,同时其拥有传值的功能,能够在多个goroutine中进行消息传递
// 在真正使用传值功能时,我们应该非常谨慎,使用context.Context传递请求的所有参数是一种非常差的设计,比较常见的使用场景是传递请求对应用户的认证令牌以及用于进行分布式追踪的请求id
