package gee

import "strings"

// 树节点上应该存储的信息
// 定义树节点结构体
type node struct {
	pattern  string  //待匹配路由（其实也是一种路径），例如/p/:lang
	part     string  //路由中的一部分，例如：:lang
	children []*node //子结点，即下一级路径
	isWild   bool    //是否精准匹配，part含有：或者*的时候为true
}

// 当我们匹配 /p/go/doc/这个路由时，第一层节点，p精准匹配到了p，第二层节点，go模糊匹配到:lang，
// 那么将会把lang这个参数赋值为go，继续下一层匹配。我们将匹配的逻辑，包装为下列辅助函数。
// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild { //找到的第一个匹配成功的节点
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0) //node切片
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 插入节点，如果没有匹配到当前part的节点，则新建一个
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height { //如果路径长度等于树高就说明最后查找的路由刚刚好在叶子结点
		n.pattern = pattern
		return
	}

	part := parts[height]       //最下面的树节点
	child := n.matchChild(part) //查找符合第一个符合该路径的路由节点
	if child == nil {           //没有找到符合要求的节点
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1) //递归继续往下找，直到所有的路径都被走完，之后创建节点
}

// 查询
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil //未查询到
}

//对于路由来说，最重要的当然是注册与匹配了。开发服务时，注册路由规则，映射handler；
//访问时，匹配路由规则，查找到对应的handler。因此，Trie 树需要支持节点的插入与查询。
//插入功能很简单，递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个，
//有一点需要注意，/p/:lang/doc只有在第三层节点，即doc节点，pattern才会设置为/p/:lang/doc。
//p和:lang节点的pattern属性皆为空。因此，当匹配结束时，我们可以使用n.pattern == ""来判断路由规则是否匹配成功。
//例如，/p/python虽能成功匹配到:lang，但:lang的pattern值为空，因此匹配失败。
//查询功能，同样也是递归查询每一层的节点，退出规则是，匹配到了*，匹配失败，或者匹配到了第len(parts)层节点。
