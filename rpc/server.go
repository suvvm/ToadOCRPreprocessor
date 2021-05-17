package rpc

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"suvvm.work/ToadOCRPreprocessor/config"
	"suvvm.work/ToadOCRPreprocessor/dal/db"
	pb "suvvm.work/ToadOCRPreprocessor/rpc/idl"
	"syscall"
	"time"
)

var (
	dbConfig = "./conf/db_config.yaml"
	successMsg = flag.String("pre success msg", "success", "rpc reply message")
	successCode = flag.Int("pre success code", 0, "rpc reply code")
	errorCode = flag.Int("pre error code", 1, "rpc reply code")
	errorLab = flag.String("pre error label", "-1", "rpc reply label")
	serv = flag.String("pre service", "toad_ocr_preprocessor", "service name")
	host = flag.String("pre host", "localhost", "listening host")
	port = flag.String("pre port", "18887", "listening port")
	reg  = flag.String("pre reg", "http://localhost:2379", "register etcd address")
)

// InitConfig 初始化配置信息
func InitConfig() {
	str, err := os.Getwd() // 获取相对路径
	if err != nil {
		panic(fmt.Sprintf("filepath failed, err=%v", err))
	}
	dbFileName, err := filepath.Abs(filepath.Join(str, dbConfig)) // 获取db配置文件路径
	if err != nil {
		panic(fmt.Sprintf("filepath failed, err=%v", err))
	}
	conf := config.Init(dbFileName)                    // 读取db配置文件
	if err = db.InitDB(&conf.DBConfig); err != nil { // 初始化db链接
		panic(fmt.Sprintf("init db conn err=%v", err))
	}
}

func RunRPCServer() {
	log.Printf("service init config")
	InitConfig()
	log.Printf("service listen port:%v", *port)
	lis, err := net.Listen("tcp", ":" + *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("register rpc server to control center...")
	err = Register(*reg, *serv, *host, *port, time.Second*10, 15)
	if err != nil {
		panic(err)
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("receive signal '%v'", s)
		UnRegister()
		os.Exit(1)
	}()
	log.Printf("create new toad ocr preprocessor rpc server...")
	s := grpc.NewServer()
	log.Printf("register handler...")
	pb.RegisterToadOcrPreprocessorServer(s, &Server{})
	log.Printf("run toad ocr rpc server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

