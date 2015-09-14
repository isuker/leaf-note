package module

import (
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/console"
	"github.com/name5566/leaf/go" //包名实际为g
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"time"
)

//骨架类型定义
type Skeleton struct {
	GoLen              int               //Go管道长度
	TimerDispatcherLen int               //定时器分发器管道长度
	ChanRPCServer      *chanrpc.Server   //RPC服务器引用（外部传入）
	g                  *g.Go             //leaf的Go机制
	dispatcher         *timer.Dispatcher //定时器分发器
	server             *chanrpc.Server   //RPC服务器引用(内部引用)
	commandServer      *chanrpc.Server   //命令RPC服务器引用
}

//初始化
func (s *Skeleton) Init() {
	//检查Go管道长度
	if s.GoLen <= 0 {
		s.GoLen = 0
	}
	//检查定时器分发器管道长度
	if s.TimerDispatcherLen <= 0 {
		s.TimerDispatcherLen = 0
	}

	s.g = g.New(s.GoLen)                                     //创建Go
	s.dispatcher = timer.NewDispatcher(s.TimerDispatcherLen) //创建分发器
	s.server = s.ChanRPCServer                               //外部传入的，内部引用

	if s.server == nil { //外部传入的为空
		s.server = chanrpc.NewServer(0) //内部创建一个
	}
	s.commandServer = chanrpc.NewServer(0) //创建命令RPC服务器
}

//实现了Module接口的Run方法
func (s *Skeleton) Run(closeSig chan bool) {
	for { //死循环
		select { //
		case <-closeSig:
			s.commandServer.Close()
			s.server.Close()
			s.g.Close()
			return
		case ci := <-s.server.ChanCall:
			err := s.server.Exec(ci)
			if err != nil {
				log.Error("%v", err)
			}
		case ci := <-s.commandServer.ChanCall:
			err := s.commandServer.Exec(ci)
			if err != nil {
				log.Error("%v", err)
			}
		case cb := <-s.g.ChanCb:
			s.g.Cb(cb)
		case t := <-s.dispatcher.ChanTimer:
			t.Cb()
		}
	}
}

func (s *Skeleton) AfterFunc(d time.Duration, cb func()) *timer.Timer {
	if s.TimerDispatcherLen == 0 {
		panic("invalid TimerDispatcherLen")
	}

	return s.dispatcher.AfterFunc(d, cb)
}

func (s *Skeleton) CronFunc(expr string, cb func()) (*timer.Cron, error) {
	if s.TimerDispatcherLen == 0 {
		panic("invalid TimerDispatcherLen")
	}

	return s.dispatcher.CronFunc(expr, cb)
}

func (s *Skeleton) Go(f func(), cb func()) {
	if s.GoLen == 0 {
		panic("invalid GoLen")
	}

	s.g.Go(f, cb)
}

func (s *Skeleton) NewLinearContext() *g.LinearContext {
	if s.GoLen == 0 {
		panic("invalid GoLen")
	}

	return s.g.NewLinearContext()
}

func (s *Skeleton) RegisterChanRPC(id interface{}, f interface{}) {
	if s.ChanRPCServer == nil {
		panic("invalid ChanRPCServer")
	}

	s.server.Register(id, f)
}

func (s *Skeleton) RegisterCommand(name string, help string, f interface{}) {
	console.Register(name, help, f, s.commandServer)
}
