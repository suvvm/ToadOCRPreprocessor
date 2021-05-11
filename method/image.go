package method

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"sort"
	// "strconv"
	"time"

	"github.com/nfnt/resize"
	"gocv.io/x/gocv"

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

/// 识别图像
func RecgnoizeImage(rimg image.Image) ([][]float64, error) {
    // log.Printf("start RecgnoizeImage")
    // resize to width 1000 using Lanczos resampling
    // and preserve aspect ratio
    m := resize.Resize(1000, 0, rimg, resize.Lanczos3)
    resizedImageName := time.Now().Format("2006010215040510")+".jpg"
    out, err := os.Create(resizedImageName)
    if err != nil {
	log.Fatal(err)
    }
    defer out.Close()
    // write new image to file
    jpeg.Encode(out, m, nil)
    img := gocv.IMRead(resizedImageName, gocv.IMReadColor)
    // log.Printf("grayImage")
    gray := grayImage(img)
    // 形态变换的预处理，得到可以查找矩形的图片
    // log.Printf("dilation")
    dilation := preprocess(gray)
    defer dilation.Close()
    // 查找和筛选文字区域
    // log.Printf("findTextRegion")
    rects := findTextRegion(dilation)

    // 再转回白底黑字
    // gocv.Threshold(img, &dilation, 0, 255, gocv.ThresholdBinaryInv)

    imageSet := make([][]float64, 0)
    // 裁剪
    // log.Printf("range rects")
    for _, rect := range rects {
	img_region := img.Region(rect)
	// 转为灰度图
	img_region = grayImage(img_region)
	// 二值化
	gocv.AdaptiveThreshold(img_region, &img_region, 225, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 31, 16)

	// 将图像压缩成28*28
	gocv.Resize(img_region, &img_region, image.Point{28, 28}, 0, 0, gocv.InterpolationLinear)

	// 输出所有图像
	// gocv.IMWrite(strconv.Itoa(i)+".jpg", img_region)

	imgFloat := make([]float64, 0)
	// dataSlice, err := binary.DataPtrUint8()
	imgBytes := img_region.ToBytes()
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
    if pixelVal == 1.0 { // 如果缩放后的值为1.0时，为了数学性能的表现稳定，将其记为0.999
	return 0.999
    }
    return pixelVal
}

/// 转换图像为灰度图
func grayImage(img gocv.Mat) gocv.Mat {
    // 创建一个空的opencv mat 用于保存灰度图
    gray := gocv.NewMat()
    // defer gray.Close()
    // 转化图像为灰度图
    gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
    return gray
}

/// 对图像进行膨胀和腐蚀处理，好让文字识别更加准确
/// 两次闭运算，一次开运算
func preprocess(gray gocv.Mat) gocv.Mat {

    // 二值化
    binary := gocv.NewMat()
    defer binary.Close()
    gocv.AdaptiveThreshold(gray, &binary, 225, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinary, 31, 16)
    // 膨胀，去掉细节，前景是黑色的话，腐蚀操作会膨胀，膨胀操作会腐蚀
    dilation := gocv.NewMat()
    defer dilation.Close()
    gocv.Dilate(binary, &dilation, gocv.GetStructuringElement(gocv.MorphRect, image.Point{4, 4}))
    // 腐蚀，加粗轮廓，如表格线等。
    erosion := gocv.NewMat()
    defer erosion.Close()
    gocv.Erode(dilation, &erosion, gocv.GetStructuringElement(gocv.MorphRect, image.Point{4, 4}))
    // 再次膨胀，腐蚀细节
    dilation2 := gocv.NewMat()
    // defer dilation2.Close()
    gocv.Dilate(erosion, &dilation2, gocv.GetStructuringElement(gocv.MorphRect, image.Point{4, 4}))
    // 腐蚀加粗轮廓
    erosion2 := gocv.NewMat()
    gocv.Erode(dilation2, &erosion2, gocv.GetStructuringElement(gocv.MorphRect, image.Point{5, 5}))

    // 开始第二次处理

    // 第二次二值化，黑白反转
    binary2 := gocv.NewMat()
    // gocv.Threshold(erosion2, &binary2, 0, 255, gocv.ThresholdBinaryInv)
    gocv.AdaptiveThreshold(erosion2, &binary2, 225, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 31, 16)
    // 再次膨胀
    dilation3 := gocv.NewMat()
    gocv.Dilate(binary2, &dilation3, gocv.GetStructuringElement(gocv.MorphRect, image.Point{12,12}))
    // gocv.Dilate(binary2, &dilation3, gocv.GetStructuringElement(gocv.MorphRect, image.Point{8,8}))

    // 再次腐蚀
    // erosion3 := gocv.NewMat()
    // gocv.Erode(dilation3, &erosion3, gocv.GetStructuringElement(gocv.MorphRect, image.Point{5, 5}))

    // 存储中间图片
    // gocv.IMWrite("binary.png", binary)
    // gocv.IMWrite("dilation.png", dilation)
    // gocv.IMWrite("erosion.png", erosion)
    // gocv.IMWrite("dilation2.png", dilation2)

    // gocv.IMWrite("erosion2.png", erosion2)
    // gocv.IMWrite("binary2.png", binary2)
    // gocv.IMWrite("erosion3.png", erosion3)
    // gocv.IMWrite("dilation3.png", dilation3)

    return dilation3
}

/// 识别文字区域
///
/// 入参
///	img gocv.Mat // 待识别的图像，要求黑底白字
/// 返回
///	[]image.Rectangle // 文字矩形区域的数组
func findTextRegion(img gocv.Mat) []image.Rectangle {
    // 查找轮廓
    rects := make([]image.Rectangle, 0)
    contours := gocv.FindContours(img, gocv.RetrievalTree, gocv.ChainApproxSimple)
    contoursSlice := make([]gocv.PointVector, 0)
    for i := 0; i < contours.Size(); i++ {
	cnt := contours.At(i)

	contoursSlice = append(contoursSlice, cnt)
    }

    rects = removeOverlappingRect(img, contoursSlice)
    return rects
}

/// 查看rect是否在数组中有重叠的部分
func findOverlapping(rect image.Rectangle, rects []image.Rectangle) bool {
    for _, item := range rects {
	intersect := item.Intersect(rect)
	if intersect.Eq(rect) || intersect.Eq(item) {
	    return true
	}
    }
    return false
}

/// 移除重叠的且不完全的rect
func removeOverlappingRect(img gocv.Mat, rects []gocv.PointVector) []image.Rectangle {

    sort.Slice(rects, func(i, j int) bool {return gocv.ContourArea(rects[i]) > gocv.ContourArea(rects[j])})

    blankMat := gocv.Zeros(img.Rows(), img.Cols(), gocv.MatTypeCV8U)
    newRects := make([]image.Rectangle, 0)
    // println(len(rects))
    for i := 0; i < len(rects); i++ {
	polygon := gocv.MinAreaRect(rects[i])
	region := blankMat.Region(polygon.BoundingRect)

	// println(region.Mean().Val1)
	if region.Mean().Val1 > 0 {
	    // println(region.Mean().Val1)
	    continue
	}

	// gocv.IMWrite(strconv.Itoa(i)+"a.jpg", region)
	gocv.Rectangle(&blankMat, polygon.BoundingRect, color.RGBA{0, 255, 255, 255}, int(gocv.Filled))
	newRects = append(newRects, polygon.BoundingRect)
    }
    // println(blankMat.Cols())
    // println(blankMat.Rows())

    // gocv.IMWrite("blank.jpg", blankMat)
    return newRects
}
