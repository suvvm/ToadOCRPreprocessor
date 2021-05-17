package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"log"
	"strconv"
	pb "suvvm.work/ToadOCRPreprocessor/client/toad_ocr_engine/idl"
	"suvvm.work/ToadOCRPreprocessor/dal/cluster"
	"suvvm.work/ToadOCRPreprocessor/dal/db"
	"suvvm.work/ToadOCRPreprocessor/model"
	"time"
)

var (
	successCode = flag.Int("success code", 0, "rpc reply code")
	serv = flag.String("service", "toad_ocr_service", "service name")
	reg  = flag.String("reg", "http://localhost:2379", "register etcd address")
	toadOCREngineClient pb.ToadOcrClient

)

func init() {
	flag.Parse()
	r := NewResolver(*reg, *serv)
	resolver.Register(r)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*serv,
		grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	toadOCREngineClient = pb.NewToadOcrClient(*conn)
}

func Predict(ctx context.Context, appID, netFlag string, image []byte) (string, error) {
	req := &pb.PredictRequest{
		AppId: appID,
		NetFlag: netFlag,
		Image: image,
	}
	token, err :=  getBasicToken(ctx, req.AppId, req.NetFlag + pixelHashStr(req.Image))
	if err != nil {
		return "", err
	}
	req.BasicToken = token
	resp, err := toadOCREngineClient.Predict(context.Background(), req)
	if err != nil {
		return "", err
	}
	if resp.Code != int32(*successCode) {
		err = fmt.Errorf("resp code not success code:%v message:%v", resp.Code, resp.Message)
		return "", err
	}
	return resp.Label, nil
}

func getBasicToken(ctx context.Context, appID, verifyStr string) (string, error) {
	hasher := md5.New()
	secret, err := cluster.GetKV(ctx, appID)
	if err != nil {
		idInt, errInner := strconv.Atoi(appID)
		if errInner != nil {
			log.Printf("AppID:%v not int", appID)
			return "", err
		}
		appInfo, errInner := db.GetAppInfo(&model.AppInfo{ID: idInt})
		if errInner != nil {
			log.Printf("db get app err:%v", errInner)
			return "", err
		}
		secret = appInfo.Secret
		cluster.PutKV(ctx, appID, secret)
	}
	//imglen := strconv.Itoa(len(req.Image))
	//hasher.Write([]byte(secret + req.NetFlag + imglen))
	hasher.Write([]byte(secret + verifyStr))
	md5Token := hex.EncodeToString(hasher.Sum(nil))
	return md5Token, nil
}

func pixelHashStr(pxs []byte) string {
	var resp string
	for i := 0; i < len(pxs); i++ {
		pixelVal := int(pxs[i])
		resp += strconv.Itoa(pixelVal)
	}
	return resp
}
