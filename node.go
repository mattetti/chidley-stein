package chidleystein

//"log"

type Node struct {
	Name            string
	Space           string
	spaceTag        string
	parent          *Node
	parents         []*Node
	Children        map[string]*Node
	childCount      map[string]int
	repeats         bool
	nodeTypeInfo    *NodeTypeInfo
	hasCharData     bool
	tempCharData    string
	DiscoveredOrder int
}

type NodeVisitor interface {
	Visit(n *Node) bool
	AlreadyVisited(n *Node) bool
	SetAlreadyVisited(n *Node)
}

func (n *Node) initialize(name string, space string, spaceTag string, parent *Node) {
	n.parent = parent
	n.parents = make([]*Node, 0, 0)
	n.pushParent(parent)
	n.Name = name
	n.Space = space
	n.spaceTag = spaceTag
	n.Children = make(map[string]*Node)
	n.childCount = make(map[string]int)
	n.nodeTypeInfo = new(NodeTypeInfo)
	n.nodeTypeInfo.initialize()
	n.hasCharData = false
}

func (n *Node) makeName() string {
	spaceTag := ""
	if n.spaceTag != "" {
		spaceTag = "_" + n.spaceTag
	}
	return capitalizeFirstLetter(cleanName(n.Name)) + spaceTag
}

func (n *Node) MakeType(prefix string, suffix string) string {
	return capitalizeFirstLetter(makeTypeGeneric(n.Name, n.spaceTag, prefix, suffix, false))
}

func (n *Node) makeJavaType(prefix string, suffix string) string {
	return capitalizeFirstLetter(makeTypeGeneric(n.Name, n.spaceTag, prefix, suffix, true))
}

func (n *Node) peekParent() *Node {
	if len(n.parents) == 0 {
		return nil
	}
	a := n.parents
	return a[len(a)-1]
}

func (n *Node) pushParent(parent *Node) {
	n.parents = append(n.parents, parent)
}

func (n *Node) popParent() *Node {
	if len(n.parents) == 0 {
		return nil
	}
	var poppedNode *Node
	a := n.parents
	poppedNode, n.parents = a[len(a)-1], a[:len(a)-1]
	return poppedNode
}

func makeTypeGeneric(name string, space string, prefix string, suffix string, capitalizeName bool) string {
	spaceTag := ""
	if space != "" {
		spaceTag = space + "_"
	}
	if capitalizeName {
		name = capitalizeFirstLetter(name)
	}
	return prefix + spaceTag + cleanName(name) + suffix

}
