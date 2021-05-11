package rpc

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"io/ioutil"
	"log"
	"suvvm.work/ToadOCRPreprocessor/common"
	pb "suvvm.work/ToadOCRPreprocessor/rpc/idl"
	"time"
)

func RunRpcClient() {
	flag.Parse()
	r := NewResolver(*reg, *serv)
	resolver.Register(r)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*serv, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
	defer cancel()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewToadOcrPreprocessorClient(*conn)
	filename := "5.jpg"
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("image size:%v", len(file))
	resp, err := client.Process(ctx, &pb.ProcessRequest{NetFlag: common.SnnName, Image: file})
	if err == nil {
		log.Printf("\nSNN\nMsg is %s\nCode is %d\nLab is %s", resp.Message, resp.Code, resp.Labels)
	}
	log.Printf("error %v", err)

	log.Printf("\nSNN\nMsg is %s\nCode is %d\nLab is %s", resp.Message, resp.Code, resp.Labels)
}
