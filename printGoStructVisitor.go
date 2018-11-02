package chidleystein

import (
	"sort"
)

type PrintGoStructVisitor struct {
	NamePrefix          string
	NameSuffix          string
	AttributePrefix     string
	AlreadyVisited      map[string]bool
	AlreadyVisitedNodes map[string]*Node
	globalTagAttributes map[string]([]*FQN)
	lineChannel         chan string
	maxDepth            int
	depth               int
	nameSpaceTagMap     map[string]string
	useType             bool
	nameSpaceInJsonName bool
}

func (v *PrintGoStructVisitor) Init(lineChannel chan string, maxDepth int, globalTagAttributes map[string]([]*FQN), nameSpaceTagMap map[string]string, useType bool, nameSpaceInJsonName bool) {
	v.AlreadyVisited = make(map[string]bool)
	v.AlreadyVisitedNodes = make(map[string]*Node)
	v.globalTagAttributes = make(map[string]([]*FQN))
	v.globalTagAttributes = globalTagAttributes
	v.lineChannel = lineChannel
	v.maxDepth = maxDepth
	v.depth = 0
	v.nameSpaceTagMap = nameSpaceTagMap
	v.useType = useType
	v.nameSpaceInJsonName = nameSpaceInJsonName
}

func (v *PrintGoStructVisitor) Visit(node *Node) bool {
	v.depth += 1

	if v.IsAlreadyVisited(node) {
		v.depth += 1
		return false
	}
	v.SetAlreadyVisited(node)

	for _, child := range node.Children {
		v.Visit(child)
	}
	v.depth += 1
	return true
}

func (v *PrintGoStructVisitor) Print(node *Node) {
	attributes := v.globalTagAttributes[nk(node)]
	v.lineChannel <- "type " + node.MakeType(v.NamePrefix, v.NameSuffix) + " struct {"
	makeAttributes(v.lineChannel, v.AttributePrefix, attributes, v.nameSpaceTagMap)
	v.printInternalFields(node)
	if node.Space != "" {
		v.lineChannel <- "\tXMLName  xml.Name `" + makeXmlAnnotation(node.Space, false, node.Name) + " " + makeJsonAnnotation(node.spaceTag, false, node.Name) + "`"
	}
	v.lineChannel <- "}\n"
}

func print(v *PrintGoStructVisitor, node *Node) {
	attributes := v.globalTagAttributes[nk(node)]
	v.lineChannel <- "type " + node.MakeType(v.NamePrefix, v.NameSuffix) + " struct {"
	makeAttributes(v.lineChannel, v.AttributePrefix, attributes, v.nameSpaceTagMap)
	v.printInternalFields(node)
	if node.Space != "" {
		v.lineChannel <- "\tXMLName  xml.Name `" + makeXmlAnnotation(node.Space, false, node.Name) + " " + makeJsonAnnotation(node.spaceTag, false, node.Name) + "`"
	}
	v.lineChannel <- "}\n"

}

func (v *PrintGoStructVisitor) IsAlreadyVisited(n *Node) bool {
	_, ok := v.AlreadyVisited[nk(n)]
	return ok
}

func (v *PrintGoStructVisitor) SetAlreadyVisited(n *Node) {
	v.AlreadyVisited[nk(n)] = true
	v.AlreadyVisitedNodes[nk(n)] = n
}

func (pn *PrintGoStructVisitor) printInternalFields(n *Node) {
	var fields []string

	var field string

	for i, _ := range n.Children {
		v := n.Children[i]
		field = "\t" + v.MakeType(pn.NamePrefix, pn.NameSuffix) + " "
		if v.repeats {
			field += "[]*"
		} else {
			field += "*"
		}
		field += v.MakeType(pn.NamePrefix, pn.NameSuffix)

		jsonAnnotation := makeJsonAnnotation(v.spaceTag, pn.nameSpaceInJsonName, v.Name)
		xmlAnnotation := makeXmlAnnotation(v.Space, false, v.Name)
		dbAnnotation := ""
		if addDbMetadata {
			dbAnnotation = " " + makeDbAnnotation(v.Space, false, v.Name)
		}

		annotation := " `" + xmlAnnotation + " " + jsonAnnotation + dbAnnotation + "`"

		field += annotation
		fields = append(fields, field)
	}

	if n.hasCharData {
		xmlString := " `xml:\",chardata\" " + makeJsonAnnotation("", false, "") + "`"
		charField := "\t" + "Text" + " " + findType(n.nodeTypeInfo, useType) + xmlString
		fields = append(fields, charField)
	}
	sort.Strings(fields)
	for i := 0; i < len(fields); i++ {
		pn.lineChannel <- fields[i]
	}
}

func makeJsonAnnotation(spaceTag string, useSpaceTagInName bool, name string) string {
	return makeAnnotation("json", spaceTag, false, useSpaceTagInName, name)
}

func makeXmlAnnotation(spaceTag string, useSpaceTag bool, name string) string {
	return makeAnnotation("xml", spaceTag, true, false, name)
}

func makeDbAnnotation(spaceTag string, useSpaceTag bool, name string) string {
	return makeAnnotation("db", spaceTag, true, false, name)
}

func makeAnnotation(annotationId string, spaceTag string, useSpaceTag bool, useSpaceTagInName bool, name string) (annotation string) {
	annotation = annotationId + ":\""

	if useSpaceTag {
		annotation = annotation + spaceTag
		annotation = annotation + " "
	}

	if useSpaceTagInName {
		if spaceTag != "" {
			annotation = annotation + spaceTag + "__"
		}
	}

	annotation = annotation + name + ",omitempty\""

	return annotation
}
