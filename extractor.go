package chidleystein

import (
	"encoding/xml"
	"io"
	"log"
	"strconv"
	"strings"
)

var nameMapper = map[string]string{
	"-": "_",
	".": "_dot_",
}

var DiscoveredOrder = 0

type Extractor struct {
	GlobalTagAttributes    map[string]([]*FQN)
	GlobalTagAttributesMap map[string]bool
	GlobalNodeMap          map[string]*Node
	NamePrefix             string
	NameSpaceTagMap        map[string]string
	nameSuffix             string
	Reader                 io.Reader
	Root                   *Node
	FirstNode              *Node
	hasStartElements       bool
	useType                bool
	progress               bool
}

func (ex *Extractor) Extract() error {
	ex.GlobalTagAttributes = make(map[string]([]*FQN))
	ex.GlobalTagAttributesMap = make(map[string]bool)
	ex.NameSpaceTagMap = make(map[string]string)
	ex.GlobalNodeMap = make(map[string]*Node)

	decoder := xml.NewDecoder(ex.Reader)

	ex.Root = new(Node)
	ex.Root.initialize("root", "", "", nil)

	ex.hasStartElements = false

	tokenChannel := make(chan xml.Token, 100)
	handleTokensDoneChannel := make(chan bool)

	go handleTokens(tokenChannel, ex, handleTokensDoneChannel)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				// OK
				break
			}
			log.Println(err)
			return err
		}
		if token == nil {
			log.Println("Empty token")
			break
		}
		tokenChannel <- xml.CopyToken(token)
	}
	close(tokenChannel)
	_ = <-handleTokensDoneChannel
	return nil
}

func handleTokens(tChannel chan xml.Token, ex *Extractor, handleTokensDoneChannel chan bool) {
	depth := 0
	thisNode := ex.Root
	first := true
	var progressCounter int64 = 0

	for token := range tChannel {
		switch element := token.(type) {
		case xml.Comment:
			if DEBUG {
				log.Print(thisNode.Name)
				log.Printf("Comment: %+v\n", string(element))
			}

		case xml.ProcInst:
			if DEBUG {
				log.Println("ProcInst: Target=" + element.Target + "  Inst=[" + string(element.Inst) + "]")
			}

		case xml.Directive:
			if DEBUG {
				log.Printf("Directive: %+v\n", string(element))
			}

		case xml.StartElement:
			progressCounter += 1
			if DEBUG {
				log.Printf("StartElement: %+v\n", element)
			}
			ex.hasStartElements = true

			if element.Name.Local == "" {
				continue
			}
			thisNode = ex.handleStartElement(element, thisNode)
			thisNode.tempCharData = ""
			if first {
				first = false
				ex.FirstNode = thisNode
			}
			depth += 1
			if progress {
				if progressCounter%50000 == 0 {
					log.Print(progressCounter)
				}
			}

		case xml.CharData:
			if DEBUG {
				log.Print(thisNode.Name)
				log.Printf("CharData: [%+v]\n", string(element))
			}

			//if !thisNode.hasCharData {
			thisNode.tempCharData += strings.TrimSpace(string(element))
		//}

		case xml.EndElement:
			thisNode.nodeTypeInfo.checkFieldType(thisNode.tempCharData)

			if DEBUG {
				log.Printf("EndElement: %+v\n", element)
				log.Printf("[[" + thisNode.tempCharData + "]]")
				log.Printf("Char is empty: ", isJustSpacesAndLinefeeds(thisNode.tempCharData))
			}
			if !thisNode.hasCharData && !isJustSpacesAndLinefeeds(thisNode.tempCharData) {
				thisNode.hasCharData = true

			} else {

			}
			thisNode.tempCharData = ""
			depth -= 1

			for key, c := range thisNode.childCount {
				if c > 1 {
					thisNode.Children[key].repeats = true
				}
				thisNode.childCount[key] = 0
			}
			if thisNode.peekParent() != nil {
				thisNode = thisNode.popParent()
			}
		}
	}
	handleTokensDoneChannel <- true
	close(handleTokensDoneChannel)
}

func space(n int) string {
	s := strconv.Itoa(n) + ":"
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}

func (ex *Extractor) findNewNameSpaces(attrs []xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" {
			ex.NameSpaceTagMap[attr.Value] = attr.Name.Local
		}
	}
}

var full struct{}

func (ex *Extractor) handleStartElement(startElement xml.StartElement, thisNode *Node) *Node {
	name := startElement.Name.Local
	space := startElement.Name.Space

	ex.findNewNameSpaces(startElement.Attr)

	var child *Node
	var attributes []*FQN
	key := nks(space, name)

	child, ok := thisNode.Children[key]
	// Does thisNode node already exist as child
	//fmt.Println(space, name)
	if ok {
		thisNode.childCount[key] += 1
		attributes, ok = ex.GlobalTagAttributes[key]
	} else {
		// if thisNode node does not already exist as child, it may still exist as child on other node:
		child, ok = ex.GlobalNodeMap[key]
		if !ok {
			child = new(Node)
			DiscoveredOrder += 1
			child.DiscoveredOrder = DiscoveredOrder
			ex.GlobalNodeMap[key] = child
			spaceTag, _ := ex.NameSpaceTagMap[space]
			child.initialize(name, space, spaceTag, thisNode)
			thisNode.childCount[key] = 1

			attributes = make([]*FQN, 0, 2)
			ex.GlobalTagAttributes[key] = attributes
		} else {
			attributes = ex.GlobalTagAttributes[key]
		}
		thisNode.Children[key] = child
	}
	child.pushParent(thisNode)

	for _, attr := range startElement.Attr {
		bigKey := key + "_" + attr.Name.Space + "_" + attr.Name.Local
		_, ok := ex.GlobalTagAttributesMap[bigKey]
		if !ok {
			fqn := new(FQN)
			fqn.name = attr.Name.Local
			fqn.space = attr.Name.Space
			attributes = append(attributes, fqn)
			ex.GlobalTagAttributesMap[bigKey] = true
		}
	}
	ex.GlobalTagAttributes[key] = attributes
	return child
}

func isJustSpacesAndLinefeeds(s string) bool {
	s = strings.Replace(s, "\\n", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return len(strings.TrimSpace(s)) == 0
}
