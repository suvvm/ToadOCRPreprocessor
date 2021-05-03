package main

import (
	"log"
	"os"
	"suvvm.work/ToadOCRPreprocessor/common"
	"suvvm.work/ToadOCRPreprocessor/rpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("Please provide command parameters\n Running with " +
			"`help` to show currently supported commands")
		return
	}

	//filename := os.Args[1]
	//file, err := os.Open(filename)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//// decode jpeg into image.Image
	//rimg, err := jpeg.Decode(file)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//file.Close()
	cmd := os.Args[1]
	if _, ok := common.CMDMap[cmd]; !ok {
		log.Printf("Unknow command!\n")
		return
	}
	if cmd == common.CmdServer {
		rpc.RunRPCServer()
		return
	} else if cmd == common.CmdClient {
		rpc.RunRpcClient()
		return
	} else if cmd == common.CmdHelp {
		log.Printf("\nToad OCR Preprocessor Help:\n" +
			"server: use command `%s` to run rpc server(etdc load blance control center must be online)\n" +
			"client: use command `%s` to run rpc client to sent one snn predict request and one cnn predict request" +
			"(etdc load blance control center must be online and at least one" +
			" Preprocessor server and one Ocr Engine server registered)",
			common.CmdServer, common.CmdClient)
		return
	} else {
		log.Printf("Unknow command!\n")
	}
}

//func main() {
//	file, err := os.Open("testImg1.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// decode jpeg into image.Image
//	rimg, err := jpeg.Decode(file)
//	if err != nil {
//		log.Fatal(err)
//	}
//	file.Close()
//
//	// 识别字符并输出28*28灰度图
//	method.RecgnoizeImage(rimg)
//}
