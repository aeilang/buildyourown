package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// MyWc保存了输入的指令，和要读取的io.Reader
type MyWc struct {
	cmds    map[Cmd]bool
	readers []*rAndErr
	isStdin bool // 是否标准输入
}

// Cmd定义了输入命令的类型
type Cmd string

var (
	IsByte Cmd = "bytes" // -c
	IsLine Cmd = "lines" // -l
	IsWord Cmd = "words" // -w
	IsChar Cmd = "chars" // -m
)

// rAndErr定义了打开的文件和打开过程发生的错误。
type rAndErr struct {
	r   io.ReadSeeker // 执行多个命令，需要读取和重置游标
	err error
}

// NewMyWc 新建MyWc结构体，期间解析命令行参数
// 打开文件，并记录过程中的错误。
func NewMyWc() (wc *MyWc, close func()) {

	c := flag.Bool("c", false, "print the byte counts")
	l := flag.Bool("l", false, "print the newline counts")
	w := flag.Bool("w", false, "print the word counts")
	m := flag.Bool("m", false, "print the character counts")

	flag.Parse()

	cmds := make(map[Cmd]bool)
	cmds[IsByte] = *c
	cmds[IsLine] = *l
	cmds[IsWord] = *w
	cmds[IsChar] = *m

	fs := flag.Args()

	wc = &MyWc{
		cmds: cmds,
	}

	readers := make([]*rAndErr, 0, len(fs))

	for _, fname := range fs {
		f, err := os.Open(fname)

		fe := &rAndErr{
			r:   f,
			err: err,
		}

		readers = append(readers, fe)
	}

	if len(readers) == 0 {
		input, err := io.ReadAll(os.Stdin)
		r := bytes.NewReader(input)
		fe := &rAndErr{
			r:   r,
			err: err,
		}
		readers = append(readers, fe)
		wc.isStdin = true
	}

	wc.readers = readers

	// 关闭文件资源
	close = func() {
		for _, re := range wc.readers {
			if re.err == nil {
				// 看re.r底层没有实现Close方法
				if c, ok := re.r.(io.Closer); ok {
					c.Close()
				}
			}
		}
	}

	return wc, close
}

// 当不指定命令时，默认设置-c -w -l三条
func (wc *MyWc) SetDefaultCmd() {
	var has bool

	for _, ok := range wc.cmds {
		if ok {
			has = true
		}
	}

	if !has {
		wc.cmds[IsByte] = true
		wc.cmds[IsWord] = true
		wc.cmds[IsLine] = true
	}
}

// WriteTo将结果打印到w
func (wc *MyWc) WriteTo(w io.Writer) (int, error) {
	// 因为map循环顺序是随机的，我们需要固定的顺序
	keys := make([]Cmd, 0, len(wc.cmds))
	for cmd, ok := range wc.cmds {
		if ok {
			keys = append(keys, cmd)
		}
	}
	slices.Sort(keys)

	// 设置打印的表头标题
	header := make([]string, len(keys))
	for i, v := range keys {
		header[i] = string(v)
	}

	// 如果是标准输入即命令行，不打印file名称
	// 因为标准输入没有文件名
	if !wc.isStdin {
		header = append(header, "file")
	}

	// results是一个二维切片，每一行存储该文件的-l -w -c等结果
	results := make([][]string, 0, len(wc.readers)+1)

	// 第一行先添加表头
	results = append(results, header)

	// 需要确定每一列最小宽度，打印才对齐
	minWidth := make([]int, len(header))

	// 把表头也进行比较
	for i, v := range header {
		minWidth[i] = len(v)
	}

	// 循环每一个reader，获取对应命令-r -w -c 等结果，并报错到results中
	for _, re := range wc.readers {
		// 如果文件打开错误，就直接存储错误，进行下一个迭代
		if re.err != nil {
			results = append(results, []string{re.err.Error()})
			continue
		}

		counts := wc.Count(re.r, keys)
		result := make([]string, 0, len(counts)+1)

		for i, c := range counts {
			str := strconv.Itoa(c)
			if len(str) > minWidth[i] {
				minWidth[i] = len(str)
			}

			result = append(result, str)
		}

		if !wc.isStdin {
			if f, ok := re.r.(*os.File); ok {
				name := f.Name()
				result = append(result, name)

				if len(name) > minWidth[len(minWidth)-1] {
					minWidth[len(minWidth)-1] = len(name)
				}
			}
		}

		results = append(results, result)
	}

	format := formatResult(results, minWidth)
	return w.Write([]byte(format))
}

// 对应数据源r, 执行对应cmd获取结果后返回
func (wc *MyWc) Count(r io.ReadSeeker, keys []Cmd) []int {
	cs := make([]int, 0, len(keys))

	for _, cmd := range keys {
		// 如果cmd对应是false就跳出，执行下一迭代
		if !wc.cmds[cmd] {
			continue
		}

		result := count(r, which(cmd))
		cs = append(cs, result)
	}

	return cs
}

// count 根据分隔的依据by进行读取，记录读取次数并返回。
func count(r io.ReadSeeker, by bufio.SplitFunc) int {
	scanner := bufio.NewScanner(r)
	scanner.Split(by)

	var total int

	for scanner.Scan() {
		total++
	}

	// 重置游标到开始文字，下一次使用就直接从头开始读了
	r.Seek(0, io.SeekStart)

	return total
}

// which 存储了Cmd类型到bufio.SplitFunc的对应关系
func which(by Cmd) bufio.SplitFunc {
	switch by {
	case IsByte:
		return bufio.ScanBytes
	case IsChar:
		return bufio.ScanRunes
	case IsLine:
		return bufio.ScanLines
	case IsWord:
		return bufio.ScanWords
	default:
		return bufio.ScanLines
	}
}

// formatResult格式化输出结果，每一列都是等宽的。
func formatResult(results [][]string, minWidth []int) string {
	var builder strings.Builder

	for _, row := range results {
		if len(row) == 1 {
			builder.WriteString(row[0])
			builder.WriteString("\n")
			continue
		}

		for i, v := range row {
			// 每一列之间用“ ”分隔
			if i > 0 {
				builder.WriteString(" ")
			}
			// *表示动态的宽度, -表示左对齐，默认右对齐。
			builder.WriteString(fmt.Sprintf("%-*s", minWidth[i], v))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
