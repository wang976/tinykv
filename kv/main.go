package main

import (
	"flag"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/server"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/storage/raft_storage"
	"github.com/pingcap-incubator/tinykv/kv/storage/standalone_storage"
	"github.com/pingcap-incubator/tinykv/log"
	"github.com/pingcap-incubator/tinykv/proto/pkg/tinykvpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// 声明命令行参数
// 三个参数: 1.命令行选项名称; 2.默认值(空字符串); 3.帮助说明

// 注: 仅声明参数, 真正读取启动命令中的值, 要等到 flag.Parse()
var (
	schedulerAddr = flag.String("scheduler", "", "scheduler address") // 调度服务地址
	storeAddr     = flag.String("addr", "", "store address")          // 当前 TinyKV 节点地址
	dbPath        = flag.String("path", "", "directory path of db")   // 数据存储目录
	logLevel      = flag.String("loglevel", "", "the level of log")   // 日志级别
)

func main() {
	// 先解析命令行参数, 再生成一份默认配置.
	flag.Parse()
	conf := config.NewDefaultConfig()

	// 命令行参数非空, 就覆盖对应的默认配置, 否则保留默认值
	if *schedulerAddr != "" {
		conf.SchedulerAddr = *schedulerAddr
	}
	if *storeAddr != "" {
		conf.StoreAddr = *storeAddr
	}
	if *dbPath != "" {
		conf.DBPath = *dbPath
	}
	if *logLevel != "" {
		conf.LogLevel = *logLevel
	}

	// 日志配置.  设置日志级别, 控制输出哪些日志.
	log.SetLevelByString(conf.LogLevel)
	// 设置日志中包括 日期、时间、微秒精度 以及 简短的源码文件名和行号.
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	// 打印配置. %+v 对结构体会显示字段名, 便于检查.
	log.Infof("Server started with conf %+v", conf)

	// 选择并启动存储引擎
	var storage storage.Storage
	if conf.Raft {
		storage = raft_storage.NewRaftStorage(conf)
	} else {
		storage = standalone_storage.NewStandAloneStorage(conf)
	}

	// 调用实际存储对象的 Start()
	if err := storage.Start(); err != nil {
		log.Fatal(err)
	}

	// 创建业务 Server.
	server := server.NewServer(storage)

	// 配置连接保活策略.  限制客户端过于频繁地发送保活探测.
	var alivePolicy = keepalive.EnforcementPolicy{
		MinTime:             2 * time.Second, // If a client pings more than once every 2 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	// 创建 gRPC Server
	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(alivePolicy), // 使用前面的保活约束
		grpc.InitialWindowSize(1<<30),                // 设置 HTTP/2 单个流的初始流控窗口
		grpc.InitialConnWindowSize(1<<30),            // 设置 HTTP/2 整个连接的初始流控窗口
		grpc.MaxRecvMsgSize(10*1024*1024),            // 将单条接收消息的大小上限设为 10MiB
	)

	// 注册 TinyKV 服务.
	// 建立网络层和业务层之间的关联: 当 gRPC 收到针对 TinyKV 服务的请求, 由业务 Server 处理.
	// tinykvpb 中包含根据 Protocol Buffers 定义生成的代码, 负责描述服务方法, 和将请求交给对应的方法.
	tinykvpb.RegisterTinyKvServer(grpcServer, server)

	// 监听 TCP 端口
	listenAddr := conf.StoreAddr[strings.IndexByte(conf.StoreAddr, ':'):]
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	// 注册信号处理, 并提供正式服务
	handleSignal(grpcServer)

	// 使用监听器接收连接并提供 RPC 服务. 正常运行期间阻塞 main 的后续执行.
	err = grpcServer.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Server stopped.")
}

// 收到操作系统信号后停止
func handleSignal(grpcServer *grpc.Server) {
	sigCh := make(chan os.Signal, 1)

	// 让操作系统信号通过这个通道通知程序
	signal.Notify(sigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sigCh
		log.Infof("Got signal [%s] to exit.", sig)
		grpcServer.Stop() // 立即停止 gRPC 服务，不会等待所有正在处理的 RPC 完成.
	}()
}
