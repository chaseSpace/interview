package main

func exampleFacade() {
	println("exampleFacade")
	// 假设example是转换文档的上下文函数

	// 其他上下文
	// ...

	// 调用外观类完成转换功能(当前上下文不关心转换细节，以后只需要修改外观类的实现，不会影响到当前上下文)
	// 实际情况中，也可以只提供一个外观函数，而不需要定义一个接口（类）
	convert := NewConvertDocAPI()
	convert.Word2PDF("./src/wordX.docx", "./dst/pdfX.pdf")

	// 其他上下文
	// ...
}

// 定义外观类（接口）

type ConvertDocAPI interface {
	Word2PDF(wordSrc, pdfDest string)
}

var _ ConvertDocAPI = (*ConverterImpl)(nil)

// 实现外观类

type ConverterImpl struct{}

func (c ConverterImpl) Word2PDF(wordSrc, pdfDest string) {
	// do convert ...
}

func NewConvertDocAPI() ConvertDocAPI {
	return &ConverterImpl{}
}
