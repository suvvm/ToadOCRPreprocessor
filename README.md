# ToadOCRPreprocessor

## 概述

​	ToadOCRPreprocessor是OCR项目ToadOCR的预处理中间件，使用OpenCV对图片进行预处理并得到适合预测的格式。主要实现图像灰度化、文字区域提取、分词、图像二值化以及产物图像压缩。最终得以将预处理后的图像交付给下层ToadOCREngine

## 核心逻辑

- 接受请求
  - ToadOCRPreprocessor RPC Client 将请求交付给ToadOCRPreprocessor RPC Server
- 样式统一
  - 将接收到的字节流恢复为图像
  - 保持图像长宽比将图像宽度压缩至1000px
  - 对图像进行灰度化处理并保存至opencv Mat
  - 形态变换的预处理，得到可以查找矩形的图片
    - 二值化
    - 膨胀使得文字轮廓突出
    - 腐蚀消去图片多余细节
    - 再次膨胀使文字轮廓更加明显
- OpenCV查找文字轮廓与分词
  - FindContours查找二进制图像中的轮廓
  - 筛选掉面积小的轮廓
  - 将筛选后的图像二值化并反转黑白像素
  - 将单个文字图像压缩为28 * 28
  - 对图像进行像素灰度缩放并读取至fload64数组中
- 调用下游
  - 将保存有待预测图像数据的数组交给 ToadOCRPreprocessor RPC Server handler
  - handler 根据字符数量启动至多5个协程并发调用 ToadOCREngine RPC Client
  - ToadOCREngine RPC Client 访问负载均衡模块 Etdc Load Balance Center
  - Etdc Load Balance Center 分配一个在线的ToadOCREngine RPC Server
  - ToadOCREngine RPC Client建立与ToadOCREngine RPC Server 将请求交付给RPC服务端
  - 得到RPC 服务端的预测响应数据
- 响应
  - handler等待所有并行预测请求处理完毕
  - 根据ToadOCREngine预测结果构造响应并返回给调用Client

## 运行指令

- 构建：运行 ``make help``查看支持的构建命令详细信息
- 运行：构建后运行``./toad_ocr_preprocessor help``查看支持的命令

## 产物结构

```
├── output
│   └── bin
│       └── toad_ocr_preprocessor           # 二进制产物
├── bootstrap.sh                        		# 启动运行脚本
└── toad_ocr_preprocessor                   # 二进制产物
```

