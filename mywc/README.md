### 构建你自己的 wc 命令行工具。

参考[codingChalleges]("https://codingchallenges.fyi/challenges/challenge-wc/#step-one") 的任务要求:

主要功能打印文件的行数`-l`, 字节数`-c`, 字符数`-m`,和单词数`-w`

下列是`linux`环境，windows 或 macOS, 使用`go build`后再执行生成的可执行性文件即可。

```sh
Usage of ./mywc:
  -c    print the 字节数
  -l    print the 行数
  -m    print the 字符数
  -w    print the 单词数
```

例如,打印 test.txt 的行数

```sh
./mywc -l test.txt
```

输出:

```sh
lines file
7145  test.txt
```

接收标准输入，例如

```sh
cat test.txt | ./mywc -l -c -w -m
```

输出:

```
bytes  chars  lines words
342190 339292 7145  58164
```

### 特点

优雅实现，文件输入和标准输入（命令输入）采用统一的接口`io.ReadSeeker`。

当然，这个简单的需求谁都能写得出来，但优雅地实现还是需要一番力气的。
