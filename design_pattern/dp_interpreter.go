package main

import (
	"strconv"
	"strings"
)

func exampleInterpreter() {
	println("exampleInterpreter")

	// 客户端
	// -- 第一种：手动构建字符串 "3 * (5 + 12)-1" 的语法树
	// -- 这种方式需要手动构建一个语法树，比较麻烦，而且不需要解析器参与，仅用于演示解释器内部原理
	add := &AddExpression{
		left:  &NumberExpression{value: 5},
		right: &NumberExpression{value: 12},
	}
	multiply := &MultiplyExpression{
		left:  &NumberExpression{value: 3},
		right: &ParenthesesExpression{Expression: add},
	}
	minus := &MinusExpression{
		left:  multiply,
		right: &NumberExpression{value: 1},
	}

	// 解释语法树
	result := minus.Interpret()
	if result != 50 {
		panic("The result of 3 * (5 + 12)-1 is not 50")
	}

	// -- 第二种：直接解释一个有意义的字符串表达式
	parser := NewParser("3 * (5 + 12)-1")
	v := parser.Parse().Interpret()
	if v != 50 {
		panic("Parser.Interpret(): The result of 3 * (5 + 12)-1 is not 50")
	}

	//NewParser("1+x").Parse() // panic: Invalid number: x
}

// 定义解释器接口
// -- 相当于语法树中的一个节点

// Expression 接口定义了任何表达式都应该具有的解释方法
type Expression interface {
	Interpret() int
}

// 实现解释器接口：加法表达式

// AddExpression 表示加法表达式
type AddExpression struct {
	left  Expression
	right Expression
}

func (ae *AddExpression) Interpret() int {
	return ae.left.Interpret() + ae.right.Interpret()
}

// 实现解释器接口：减法表达式

type MinusExpression struct {
	left  Expression
	right Expression
}

func (ae *MinusExpression) Interpret() int {
	return ae.left.Interpret() - ae.right.Interpret()
}

// 实现解释器接口：乘法表达式

// MultiplyExpression 表示乘法表达式
type MultiplyExpression struct {
	left  Expression
	right Expression
}

func (me *MultiplyExpression) Interpret() int {
	return me.left.Interpret() * me.right.Interpret()
}

// 实现解释器接口：数字表达式

// NumberExpression 表示数字表达式
type NumberExpression struct {
	value int
}

// Interpret 实现了 Expression 接口的 Interpret 方法
func (ne *NumberExpression) Interpret() int {
	return ne.value
}

// 定义非终结符表达式（上面的表达式都是终结符表达式）

// ParenthesesExpression 括号表达式
type ParenthesesExpression struct {
	Expression Expression
}

func (p *ParenthesesExpression) Interpret() int {
	return p.Expression.Interpret()
}

// 解析器：将输入的字符串解析为抽象语法树（难点）
// -- 语法树由多个Expression节点组成
// -- 难点是如何将一个可解释的字符串解析为语法树

type Parser struct {
	input string
	index int
}

func NewParser(input string) *Parser {
	return &Parser{
		input: strings.ReplaceAll(input, " ", ""),
	}
}

func (p *Parser) NextToken() string {
	index := p.index
	p.index++
	if index >= len(p.input) {
		return ""
	}
	return string(p.input[index])
}

func (p *Parser) Parse() Expression {
	return p.parseTerm()
}

// parseTerm 负责解析加减乘表达式以及因子（数字和括号内的表达式）
// -- 因子是指二元操作符两边的操作元素，如3+5，则3和5是这个表达式的两个因子
func (p *Parser) parseTerm() Expression {
	node := p.parseFactor()
	//fmt.Printf("parseTerm: %#v\n", node)
	for {
		switch p.NextToken() {
		case "*":
			//println(111)
			node = &MultiplyExpression{
				left:  node,
				right: p.parseFactor(),
			}
		case "-":
			//println(222)
			node = &MinusExpression{
				left:  node,
				right: p.parseFactor(),
			}
		case "+":
			//println(333)
			node = &AddExpression{
				left:  node,
				right: p.parseFactor(),
			}
		default:
			p.index-- // 回退非加减乘因子的token
			return node
		}
	}
}

// parseFactor 负责解析操作因子，即数字和括号内的表达式
func (p *Parser) parseFactor() Expression {
	tok := p.NextToken()
	if tok == "(" {
		node := p.parseTerm()
		p.Expect(")")
		return &ParenthesesExpression{Expression: node}
	} else {
		p.index-- // 回退到数字的开始位置
		return &NumberExpression{value: p.parseNumber()}
	}
}

func (p *Parser) Expect(token string) {
	tok := p.NextToken()
	if tok != token {
		panic("Expected " + token + " but got " + tok)
	}
}

// parseNumber 解析数字
func (p *Parser) parseNumber() int {
	start := p.index
	for p.index < len(p.input) && (p.input[p.index] >= '0' && p.input[p.index] <= '9') {
		p.index++
	}
	if start == p.index {
		panic("Invalid number: " + p.input[p.index:p.index+1])
	}
	numberStr := p.input[start:p.index]
	i, err := strconv.Atoi(numberStr)
	if err != nil {
		// may be overflow
		panic("failed to parse number: " + err.Error())
	}
	return i
}
