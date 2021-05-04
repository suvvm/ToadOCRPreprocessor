package method

import (
	"github.com/nfnt/resize"
	"gocv.io/x/gocv"
	"image"
	"image/jpeg"
	"log"
	"os"
	client "suvvm.work/ToadOCRPreprocessor/client/toad_ocr_engine"
	"sync"
)

func OCRGetLabels(netFlag string, floatImage []float64, labels *[]string, lock *sync.Mutex, wg *sync.WaitGroup, ch chan int) {
	defer func() {
		<-ch
		wg.Done()
	}()
	resp, err := client.Predict(netFlag, floatImage)
	if err != nil {
		log.Printf("method:OCRGetLabels err %v", err)
		return
	}
	lock.Lock()
	*labels = append(*labels, resp)
	lock.Unlock()
}

func RecgnoizeImage(rimg image.Image) ([][]float64, error) {
	// log.Printf("start RecgnoizeImage")
	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(1000, 0, rimg, resize.Lanczos3)
	// log.Printf("create test_resized")
	out, err := os.Create("test_resized.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	// write new image to file
	jpeg.Encode(out, m, nil)
	// log.Printf("read test_resized")
	img := gocv.IMRead("test_resized.jpg", gocv.IMReadColor)
	// log.Printf("grayImage")
	gray := grayImage(img)
	// 形态变换的预处理，得到可以查找矩形的图片
	// log.Printf("dilation")
	dilation := preprocess(gray)
	defer dilation.Close()
	// 查找和筛选文字区域
	// log.Printf("findTextRegion")
	rects := findTextRegion(dilation)
	imageSet := make([][]float64, 0)
	// 裁剪
	// log.Printf("range rects")
	for _, rect := range rects {
		img_region := img.Region(rect)
		// 转为灰度图
		img_region = grayImage(img_region)
		// 将图像压缩成28*28
		gocv.Resize(img_region, &img_region, image.Point{28, 28}, 0, 0, gocv.InterpolationLinear)
		// 二值化mat
		binary := gocv.NewMat()
		gocv.Threshold(img_region, &binary, 0, 255, gocv.ThresholdOtsu+gocv.ThresholdBinary)
		gocv.BitwiseNot(binary, &binary)
		// 输出所有图像
		//gocv.IMWrite(strconv.Itoa(i)+".jpg", binary)
		imgFloat := make([]float64, 0)
		// dataSlice, err := binary.DataPtrUint8()
		imgBytes := binary.ToBytes()
		// 像素缩放
		for _, b := range imgBytes {
			imgFloat = append(imgFloat, PixelWeight(b))
		}
		if err != nil {
			log.Printf("fail to DataPtrFloat32 %v", err)
			return nil, err
		}

		log.Printf("append imageSet")
		imageSet = append(imageSet, imgFloat)
		// 用绿线画出这些找到的轮廓
		// gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 255}, 2)
	}
	// 显示带轮廓的图像
	// gocv.IMWrite("imgDrawRect.jpg", img)
	log.Printf("inner imageSet size:%v", len(imageSet))
	return imageSet, nil
}

// PixelWeight 像素灰度缩放
// mnist给出的图像作为灰度图，单个像素点通过8位的灰度值(0~255)来表示。
// PixelWeight 函数将字节类型的像素灰度值转换为float64类型，并将范围缩放至(0.0~1.0)
//
// 入参
//	px byte	// 字节型像素灰度值
//
// 返回
//	float64 // 缩放后的浮点灰度值
func PixelWeight(px byte) float64 {
	pixelVal := (float64(px) / 255 * 0.999) + 0.001
	if pixelVal == 1.0 {	// 如果缩放后的值为1.0时，为了数学性能的表现稳定，将其记为0.999
		return 0.999
	}
	return pixelVal
}


func grayImage(img gocv.Mat) gocv.Mat {
	// 创建一个空的opencv mat 用于保存灰度图
	gray := gocv.NewMat()
	// defer gray.Close()
	// 转化图像为灰度图
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	return gray
}

func preprocess(gray gocv.Mat) gocv.Mat {
	// Sobel算子，x,y方向分别求梯度
	sobelX := gocv.NewMat()
	sobelY := gocv.NewMat()
	gocv.Sobel(gray, &sobelX, gocv.MatTypeCV64F, 1, 0, 3, 1, 0, gocv.BorderDefault)
	gocv.Sobel(gray, &sobelY, gocv.MatTypeCV64F, 0, 1, 3, 1, 0, gocv.BorderDefault)

	// 错误的计算方法，但是起码最终结果是好的……待改正
	absSobelX := gocv.NewMat()
	absSobelY := gocv.NewMat()
	gocv.ConvertScaleAbs(sobelX, &absSobelX, 1, 0)
	gocv.ConvertScaleAbs(sobelY, &absSobelY, 1, 0)

	sobel := gocv.NewMat()
	gocv.AddWeighted(absSobelX, 0.5, absSobelY, 0.5, 0, &sobel)

	// 二值化
	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(sobel, &binary, 0, 255, gocv.ThresholdOtsu+gocv.ThresholdBinary)
	// 膨胀和腐蚀操作的核函数
	element1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{10, 5})
	element2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{10, 5})
	// 膨胀，让轮廓突出
	dilation := gocv.NewMat()
	defer dilation.Close()
	gocv.Dilate(binary, &dilation, element2)
	// 腐蚀，去掉细节，如表格线等。注意这里去掉的是竖直的线
	erosion := gocv.NewMat()
	defer erosion.Close()
	gocv.Erode(dilation, &erosion, element1)
	// 再次膨胀，使轮廓明显
	dilation2 := gocv.NewMat()
	// defer dilation2.Close()
	gocv.Dilate(erosion, &dilation2, element2)
	// 存储中间图片
	// gocv.IMWrite("binary.png", binary)
	// gocv.IMWrite("dilation.png", dilation)
	// gocv.IMWrite("erosion.png", erosion)
	// gocv.IMWrite("dilation2.png", dilation2)
	return dilation2
}

func findTextRegion(img gocv.Mat) []image.Rectangle {
	// 查找轮廓
	rects := make([]image.Rectangle, 0)
	contours := gocv.FindContours(img, gocv.RetrievalTree, gocv.ChainApproxSimple)
	for i := 0; i < contours.Size(); i++ {
		cnt := contours.At(i)
		// 计算该轮廓的面积
		area := gocv.ContourArea(cnt)
		// 面积小的都筛选掉
		// 可以调节 1000
		if area < 500 {
			continue
		}
		// 轮廓近似，作用很小
		epsilon := 0.001 * gocv.ArcLength(cnt, true)
		// approx := gocv.ApproxPolyDP(cnt, epsilon, true)
		_ = gocv.ApproxPolyDP(cnt, epsilon, true)
		// 找到最小矩形，该矩形可能有方向
		rect := gocv.MinAreaRect(cnt)
		// fmt.Println(rect.Points)
		// fmt.Println(rect.Points)
		// fmt.Println(rect.BoundingRect.Max)
		// fmt.Println(rect.BoundingRect.Min)
		// 计算高和宽
		// mWidth := float64(rect.BoundingRect.Max.X - rect.BoundingRect.Min.X)
		// mHeight := float64(rect.BoundingRect.Max.Y - rect.BoundingRect.Min.Y)
		// 筛选那些太细的矩形，留下扁的
		// 可以调节 mHeight > (mWidth * 1.2)
		// if mHeight > (mWidth * 0.8) {
		// 	continue
		// }
		// 符合条件的rect添加到rects集合中
		rects = append(rects, rect.BoundingRect)
	}
	return rects
}

