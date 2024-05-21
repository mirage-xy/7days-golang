//配置文件
//定义模块名字
//每一个go都应该有模块，这个模块名字是项目的唯一标识
module geecache

//指定了构建此模块所需的Go语言版本
go 1.21

//替换指令 尝试获取“gee”模块时，不要从模块代理或其他源获取，而是从当前项目的./gee子目录中获取

require (
	github.com/golang/protobuf v1.5.4 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
