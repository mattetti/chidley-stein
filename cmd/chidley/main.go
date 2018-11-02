package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/mattetti/chidley-stein"
)

var (
	DEBUG               bool
	progress            bool
	attributePrefix     = "Attr_"
	structsToStdout     bool
	nameSpaceInJsonName bool
	prettyPrint         bool
	codeGenConvert      bool
	readFromStandardIn  bool
	sortByXmlOrder      bool
	codeGenDir          = "codegen"
	codeGenFilename     = "CodeGenStructs.go"
)

// Java out
const javaBasePackage = "ca.gnewton.chidley"
const mavenJavaBase = "src/main/java"

var (
	javaBasePackagePath = strings.Replace(javaBasePackage, ".", "/", -1)
	javaAppName         = "jaxb"
	// writeJava           = false
	baseJavaDir         = "java"
	userJavaPackageName = ""
	namePrefix          = "Chi"
	nameSuffix          = ""
	xmlName             = false
	url                 = false
	useType             = false
	addDbMetadata       = false
)

type structSortFunc func(v *chidleystein.PrintGoStructVisitor)

var structSort = printStructsAlphabetical

var outputs = []*bool{
	&codeGenConvert,
	&structsToStdout,
	// &writeJava,
}

func init() {

	flag.BoolVar(&DEBUG, "d", DEBUG, "Debug; prints out much information")
	flag.BoolVar(&addDbMetadata, "B", addDbMetadata, "Add database metadata to created Go structs")
	flag.BoolVar(&sortByXmlOrder, "X", sortByXmlOrder, "Sort output of structs in Go code by order encounered in source XML (default is alphabetical order)")
	flag.BoolVar(&codeGenConvert, "W", codeGenConvert, "Generate Go code to convert XML to JSON or XML (latter useful for validation) and write it to stdout")
	flag.BoolVar(&nameSpaceInJsonName, "n", nameSpaceInJsonName, "Use the XML namespace prefix as prefix to JSON name; prefix followed by 2 underscores (__)")
	flag.BoolVar(&prettyPrint, "p", prettyPrint, "Pretty-print json in generated code (if applicable)")
	flag.BoolVar(&progress, "r", progress, "Progress: every 50000 input tags (elements)")
	flag.BoolVar(&readFromStandardIn, "c", readFromStandardIn, "Read XML from standard input")
	flag.BoolVar(&structsToStdout, "G", structsToStdout, "Only write generated Go structs to stdout")
	// flag.BoolVar(&url, "u", url, "Filename interpreted as an URL")
	flag.BoolVar(&useType, "t", useType, "Use type info obtained from XML (int, bool, etc); default is to assume everything is a string; better chance at working if XMl sample is not complete")
	// flag.BoolVar(&writeJava, "J", writeJava, "Generated Java code for Java/JAXB")
	flag.BoolVar(&xmlName, "x", xmlName, "Add XMLName (Space, Local) for each XML element, to JSON")
	flag.StringVar(&attributePrefix, "a", attributePrefix, "Prefix to attribute names")
	// flag.StringVar(&baseJavaDir, "D", baseJavaDir, "Base directory for generated Java code (root of maven project)")
	// flag.StringVar(&javaAppName, "k", javaAppName, "App name for Java code (appended to ca.gnewton.chidley Java package name))")
	flag.StringVar(&namePrefix, "e", namePrefix, "Prefix to struct (element) names; must start with a capital")
	// flag.StringVar(&userJavaPackageName, "P", userJavaPackageName, "Java package name (rightmost in full package name")
}

func handleParameters() error {
	flag.Parse()

	numBoolsSet := countNumberOfBoolsSet(outputs)
	if numBoolsSet > 1 {
		log.Print("  ERROR: Only one of -W -J -X -V -c can be set")
	} else if numBoolsSet == 0 {
		log.Print("  ERROR: At least one of -W -J -X -V -c must be set")
	}
	if sortByXmlOrder {
		structSort = printStructsByXml
	}
	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := handleParameters()
	// chidleystein.DEBUG = true

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err != nil {
		flag.Usage()
		return
	}

	if len(flag.Args()) != 1 && !readFromStandardIn {
		fmt.Println("chidley <flags> xmlFileName|url")
		fmt.Println("xmlFileName can be .gz or .bz2: uncompressed transparently")
		flag.Usage()
		return
	}

	var sourceName string

	if !readFromStandardIn {
		sourceName = flag.Args()[0]
	}
	if !url && !readFromStandardIn {
		sourceName, err = filepath.Abs(sourceName)
		if err != nil {
			log.Fatal("FATAL ERROR: " + err.Error())
		}
	}

	source, err := makeSourceReader(sourceName, url, readFromStandardIn)
	if err != nil {
		log.Fatal("FATAL ERROR: " + err.Error())
	}

	ex := chidleystein.Extractor{
		NamePrefix: namePrefix,
		// NameSuffix: nameSuffix,
		Reader: source.GetReader(),
		// useType:    useType,
		// progress:   progress,
	}

	// if DEBUG {
	log.Print("extracting")
	// }
	err = ex.Extract()
	if err != nil {
		log.Fatal("FATAL ERROR: " + err.Error())
	}

	var writer chidleystein.Writer
	lineChannel := make(chan string, 100)
	switch {
	case codeGenConvert:
		sWriter := new(chidleystein.StringWriter)
		writer = sWriter
		writer.Open("", lineChannel)

		printGoStructVisitor := new(chidleystein.PrintGoStructVisitor)
		printGoStructVisitor.Init(lineChannel, 9999,
			ex.GlobalTagAttributes,
			ex.NameSpaceTagMap,
			useType,
			nameSpaceInJsonName)
		printGoStructVisitor.NamePrefix = namePrefix
		printGoStructVisitor.NameSuffix = nameSuffix
		printGoStructVisitor.AttributePrefix = attributePrefix

		printGoStructVisitor.Visit(ex.Root)

		structSort(printGoStructVisitor)

		close(lineChannel)
		sWriter.Close()

		xt := chidleystein.XMLType{NameType: ex.FirstNode.MakeType(namePrefix, nameSuffix),
			XMLName:      ex.FirstNode.Name,
			XMLNameUpper: chidleystein.CapitalizeFirstLetter(ex.FirstNode.Name),
			XMLSpace:     ex.FirstNode.Space,
		}

		x := chidleystein.XmlInfo{
			BaseXML:         &xt,
			OneLevelDownXML: makeOneLevelDown(ex.Root),
			Filename:        chidleystein.GetFullPath(sourceName),
			Structs:         sWriter.S,
		}
		t := template.Must(template.New("chidleyGen").Parse(chidleystein.CodeTemplate))

		err := t.Execute(os.Stdout, x)
		if err != nil {
			log.Println("executing template:", err)
		}

	case structsToStdout:
		writer = new(chidleystein.StdoutWriter)
		writer.Open("", lineChannel)
		printGoStructVisitor := new(chidleystein.PrintGoStructVisitor)
		printGoStructVisitor.Init(lineChannel, 999, ex.GlobalTagAttributes, ex.NameSpaceTagMap, useType, nameSpaceInJsonName)
		printGoStructVisitor.Visit(ex.Root)
		structSort(printGoStructVisitor)
		close(lineChannel)
		writer.Close()
	}

}

const XMLNS = "xmlns"

func findNameSpaces(attributes []*chidleystein.FQN) []*chidleystein.FQN {
	if attributes == nil || len(attributes) == 0 {
		return nil
	}
	xmlns := make([]*chidleystein.FQN, 0)
	return xmlns
}

func makeSourceReader(sourceName string, url bool, standardIn bool) (chidleystein.Source, error) {
	var err error

	var source chidleystein.Source
	if url {
		source = new(chidleystein.UrlSource)
		if DEBUG {
			log.Print("Making UrlSource")
		}
	} else {
		if standardIn {
			source = new(chidleystein.StdinSource)
			if DEBUG {
				log.Print("Making StdinSource")
			}
		} else {
			source = new(chidleystein.FileSource)
			if DEBUG {
				log.Print("Making FileSource")
			}
		}
	}
	if DEBUG {
		log.Print("Making Source:[" + sourceName + "]")
	}
	err = source.NewSource(sourceName)
	return source, err
}

func attributes(atts map[string]bool) string {
	ret := ": "
	for k, _ := range atts {
		ret = ret + k + ", "
	}
	return ret
}

func indent(d int) string {
	indent := ""
	for i := 0; i < d; i++ {
		indent = indent + "\t"
	}
	return indent
}

func lowerFirstLetter(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}

func countNumberOfBoolsSet(a []*bool) int {
	counter := 0
	for i := 0; i < len(a); i++ {
		if *a[i] {
			counter += 1
		}
	}
	return counter
}

func makeOneLevelDown(node *chidleystein.Node) []*chidleystein.XMLType {
	var children []*chidleystein.XMLType

	for _, np := range node.Children {
		if np == nil {
			continue
		}
		for _, n := range np.Children {
			if n == nil {
				continue
			}
			x := chidleystein.XMLType{NameType: n.MakeType(namePrefix, nameSuffix),
				XMLName:      n.Name,
				XMLNameUpper: chidleystein.CapitalizeFirstLetter(n.Name),
				XMLSpace:     n.Space}
			children = append(children, &x)
		}
	}
	return children
}
func printChildrenChildren(node *chidleystein.Node) {
	for k, v := range node.Children {
		log.Print(k)
		log.Printf("children: %+v\n", v.Children)
	}
}

// Order Xml is encountered
func printStructsByXml(v *chidleystein.PrintGoStructVisitor) {
	orderNodes := make(map[int]*chidleystein.Node)
	var order []int

	for k := range v.AlreadyVisitedNodes {
		nodeOrder := v.AlreadyVisitedNodes[k].DiscoveredOrder
		orderNodes[nodeOrder] = v.AlreadyVisitedNodes[k]
		order = append(order, nodeOrder)
	}
	sort.Ints(order)

	for o := range order {
		v.Print(orderNodes[o])
	}
}

// Alphabetical order
func printStructsAlphabetical(v *chidleystein.PrintGoStructVisitor) {
	var keys []string
	for k := range v.AlreadyVisitedNodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v.Print(v.AlreadyVisitedNodes[k])
	}

}
