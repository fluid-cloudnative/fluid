# fluid单元测试说明

## 如何运行测试

在项目根路径下执行如下命令，将自动运行所有测试样例并输出测试结果

```shell
$ make test
```

## 单元测试方案

fluid包含三个层级的测试：

- 工具util测试
- Engine测试
- controller测试

针对三个层级的测试会采用不同的测试框架和测试方案。

### 通用测试规则

为了让`go test`正确定位测试代码，首先应满足以下规则：

- 测试代码文件以`xxx_test.go`的格式结尾，其中`xxx`对应某个待测试代码文件名字
- 测试用例函数以`TestXxx()`的格式开头，其中`Xxx`是某个函数或者功能名字
- 尽量使用`Table Driven`的测试结构

### 工具util测试

#### 测试方案

直接使用go提供的`testing`单元测试框架。下面以`pkg/utils/crtl_utils.go`的单元测试为例说明。

1. 创建代码文件`crtl_utils.go`对应的测试代码文件`crtl_utils_test.go`
2. 创建`GetOrDefault`方法对应的测试函数`TestGetOrDefault`
3. 使用`Table Driven`风格，定义一系列测试样例，并循环测试

```go
package utils

import (
	"testing"
)

func TestGetOrDefault(t *testing.T) {
	var defaultStr = "default string"
	var nonnullStr = "non-null string"
	var tests = []struct {
		pstr        *string
		defaultStr  string
		expectedStr string
	} {
		{&nonnullStr, defaultStr, nonnullStr},
		{nil, defaultStr, defaultStr},
	}

	for _, test := range tests {
		if str := GetOrDefault(test.pstr, test.defaultStr); str != test.expectedStr {
			t.Errorf("expected %s, got %s", test.expectedStr, str)
		}
	}
}
```

### Engine测试

tbd

### controller测试

#### 测试方案

使用[Ginkgo](https://www.ginkgo.wiki/)测试框架。

安装Ginkgo

```shell
$ go get github.com/onsi/ginkgo/ginkgo
$ go get github.com/onsi/gomega/...
```



