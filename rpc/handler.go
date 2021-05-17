package rpc

import (
	"bytes"
	"context"
	"image"
	"log"
	"strconv"
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
	err := method.VerifySecret(ctx, in.AppId, in.BasicToken, in.Name)
	if err != nil {
		return &pb.PongReply{Message: "permission denied"}, nil
	}
	log.Printf("Received: %v", in.GetName())
	return &pb.PongReply{Message: "Hello " + in.GetName()}, nil
}

func (s *Server)Process(ctx context.Context, in *pb.ProcessRequest) (*pb.ProcessReply, error) {
	log.Printf("app:%v Received Process from", in.AppId)
	err := method.VerifySecret(ctx, in.AppId, in.BasicToken, in.NetFlag + strconv.Itoa(len(in.Image)))
	if err != nil {
		log.Printf("app:%v permission denied", in.AppId)
		return &pb.ProcessReply{Code: int32(*errorCode), Message: "permission denied", Labels: make([]string, 0)}, nil
	}
	// log.Printf("Received Process")
	// log.Printf("in.Image size:%v", len(in.Image))
	inImage, _, err := image.Decode(bytes.NewReader(in.Image))
	if err != nil {
		log.Printf("app:%v Decode image err %v", in.AppId, err)
		return &pb.ProcessReply{Code: int32(*errorCode), Message: err.Error(), Labels: make([]string, 0)}, nil
	}
	imageSet, err := method.RecgnoizeImage(inImage)
	if err != nil {
		log.Printf("app:%v RecgnoizeImage err %v", in.AppId, err)
		return &pb.ProcessReply{Code: int32(*errorCode), Message: err.Error(), Labels:  make([]string, 0)}, nil
	}
	log.Printf("app:%v Process floatImageSet size:%v", in.AppId, len(imageSet))
	labels := make([]string, len(imageSet))
	var lock sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan int, 5)
	for i := 0; i < len(imageSet); i++ {
		tempImage := imageSet[i]
		tmpIndex := i
		wg.Add(1)
		ch <- 1
		go method.OCRGetLabels(in.NetFlag, tempImage, &labels, &lock, &wg, ch, tmpIndex)
	}
	wg.Wait()
	return &pb.ProcessReply{Code: int32(*successCode), Message: *successMsg, Labels: labels}, nil
}

