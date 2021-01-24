package converters

type Node struct {
	Url           string
	ObjectId      string
	Repeat        int
	BulletNesting int64
	ListNumber    int
	Token         Token
	parent        *Node
	Children      []*Node
	Content       string
}

func (n *Node) append(node *Node) *Node {
	node.parent = n
	n.Children = append(n.Children, node)

	return node
}
