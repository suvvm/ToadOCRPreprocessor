package rpc

import (
	"bytes"
	"context"
	"image"
	"log"
	"suvvm.work/ToadOCRPreprocessor/method"
	pb "suvvm.work/ToadOCRPreprocessor/rpc/idl"
	"sync"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	pb.UnimplementedToadOcrPreprocessorServer
}

// SayHello implements helloworld.GreeterServer
func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.PongReply{Message: "Hello " + in.GetName()}, nil
}

func (s *Server)Process(ctx context.Context, in *pb.ProcessRequest) (*pb.ProcessReply, error) {
	log.Printf("Received Process")
	labels := make([]string, 0)
	log.Printf("in.Image size:%v", len(in.Image))
	inImage, _, err := image.Decode(bytes.NewReader(in.Image))
	if err != nil {
		log.Printf("rpc handler Process err %v", err)
		return &pb.ProcessReply{Code: int32(*errorCode), Message: err.Error(), Labels: labels}, err
	}
	imageSet, err := method.RecgnoizeImage(inImage)
	log.Printf("Process floatImageSet size:%v", len(imageSet))
	if err != nil {
		return &pb.ProcessReply{Code: int32(*errorCode), Message: err.Error(), Labels: labels}, err
	}
	var lock sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan int, 5)
	for i := 0; i < len(imageSet); i++ {
		tempImage := imageSet[i]
		wg.Add(1)
		ch <- 1
		go method.OCRGetLabels(in.NetFlag, tempImage, &labels, &lock, &wg, ch)
	}
	wg.Wait()
	return &pb.ProcessReply{Code: int32(*successCode), Message: *successMsg, Labels: labels}, nil
}

