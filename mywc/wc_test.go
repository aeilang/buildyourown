package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestText(t *testing.T) {
	os.Args = []string{"./mywc", "-c", "-l", "-m", "-w", "test.txt"}

	wc, close := NewMyWc()
	defer close()

	wc.SetDefaultCmd()
	var buf bytes.Buffer
	wc.WriteTo(&buf)
	output := buf.String()

	expected := "bytes  chars  lines words file    \n342190 339292 7145  58164 test.txt\n"

	if output != expected {
		t.Errorf("输出不匹配\n期望:\n%s\n实际:\n%s", expected, output)
	}
}

// TestNewMyWc 测试 NewMyWc 函数是否正确初始化
func TestNewMyWc(t *testing.T) {
	// 创建一个临时文件，写入测试数据
	file, err := os.CreateTemp("", "test_wc")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(file.Name()) // 测试结束后删除文件

	content := "Hello World\nThis is a test\n"
	file.WriteString(content)
	file.Seek(0, io.SeekStart) // 重置文件指针

	// 模拟命令行参数
	os.Args = []string{"wc", "-c", file.Name()}

	wc, close := NewMyWc()
	defer close()

	if len(wc.readers) != 1 {
		t.Errorf("预期1个reader，实际有%d个", len(wc.readers))
	}
}

// TestWriteTo 测试 WriteTo 函数的输出
func TestWriteTo(t *testing.T) {
	content := "Hello World\nThis is a test\n"

	// 创建一个内存 reader，模拟输入
	reader := &rAndErr{
		r:   bytes.NewReader([]byte(content)),
		err: nil,
	}

	wc := &MyWc{
		cmds: map[Cmd]bool{
			IsByte: true,
			IsLine: true,
			IsWord: true,
		},
		readers: []*rAndErr{reader},
		isStdin: true,
	}

	var buf bytes.Buffer
	_, err := wc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo 出现错误: %v", err)
	}

	output := buf.String()
	expected := "bytes lines words\n27    2     6    \n"
	if output != expected {
		t.Errorf("输出不匹配\n期望:\n%s\n实际:\n%s", expected, output)
	}
}

// TestCount 测试 Count 函数
func TestCount(t *testing.T) {
	content := "Hello World\nThis is a test\n"
	reader := bytes.NewReader([]byte(content))

	wc := &MyWc{
		cmds: map[Cmd]bool{
			IsWord: true,
		},
	}

	result := wc.Count(reader, []Cmd{IsWord})
	if len(result) != 1 || result[0] != 6 {
		t.Errorf("预期5个单词，实际得到%d个", result[0])
	}
}

// TestFormatResult 测试格式化输出
func TestFormatResult(t *testing.T) {
	results := [][]string{
		{"bytes", "lines", "words"},
		{"26", "2", "5"},
	}
	minWidth := []int{5, 5, 5}
	expected := "bytes lines words\n26    2     5    \n"

	output := formatResult(results, minWidth)
	if output != expected {
		t.Errorf("输出格式化错误\n期望:\n%s\n实际:\n%s", expected, output)
	}
}
