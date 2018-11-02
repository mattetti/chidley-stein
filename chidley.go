package chidleystein

// Copyright 2014,2015,2016 Glen Newton
// glen.newton@gmail.com

import (
	"bufio"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

var DEBUG = false
var progress = false
var attributePrefix = "Attr_"
var structsToStdout = false
var nameSpaceInJsonName = false
var prettyPrint = false
var codeGenConvert = false
var readFromStandardIn = false
var sortByXmlOrder = false

var codeGenDir = "codegen"
var codeGenFilename = "CodeGenStructs.go"

// Java out
const javaBasePackage = "ca.gnewton.chidley"
const mavenJavaBase = "src/main/java"

var javaBasePackagePath = strings.Replace(javaBasePackage, ".", "/", -1)
var javaAppName = "jaxb"
var writeJava = false
var baseJavaDir = "java"
var userJavaPackageName = ""

var nameSuffix = ""
var xmlName = false
var url = false
var useType = false
var addDbMetadata = false

type structSortFunc func(v *PrintGoStructVisitor)

var structSort = printStructsAlphabetical

type Writer interface {
	Open(s string, lineChannel chan string) error
	Close()
}

var outputs = []*bool{
	&codeGenConvert,
	&structsToStdout,
	&writeJava,
}

func printPackageInfo(node *Node, javaDir string, javaPackage string, globalTagAttributes map[string][]*FQN, nameSpaceTagMap map[string]string) {

	//log.Printf("%+v\n", node)

	if node.Space != "" {
		_ = findNameSpaces(globalTagAttributes[nk(node)])
		//attributes := findNameSpaces(globalTagAttributes[nk(node)])

		t := template.Must(template.New("package-info").Parse(jaxbPackageInfoTemplage))
		packageInfoPath := javaDir + "/xml/package-info.java"
		fi, err := os.Create(packageInfoPath)
		if err != nil {
			log.Print("Problem creating file: " + packageInfoPath)
			panic(err)
		}
		defer fi.Close()

		writer := bufio.NewWriter(fi)
		packageInfo := JaxbPackageInfo{
			BaseNameSpace: node.Space,
			//AdditionalNameSpace []*FQN
			PackageName: javaPackage + ".xml",
		}
		err = t.Execute(writer, packageInfo)
		if err != nil {
			log.Println("executing template:", err)
		}
		bufio.NewWriter(writer).Flush()
	}

}

const XMLNS = "xmlns"

func findNameSpaces(attributes []*FQN) []*FQN {
	if attributes == nil || len(attributes) == 0 {
		return nil
	}
	xmlns := make([]*FQN, 0)
	return xmlns
}

func printMavenPom(pomPath string, javaAppName string) {
	t := template.Must(template.New("mavenPom").Parse(mavenPomTemplate))
	fi, err := os.Create(pomPath)
	if err != nil {
		log.Print("Problem creating file: " + pomPath)
		panic(err)
	}
	defer fi.Close()

	writer := bufio.NewWriter(fi)
	maven := JaxbMavenPomInfo{
		AppName: javaAppName,
	}
	err = t.Execute(writer, maven)
	if err != nil {
		log.Println("executing template:", err)
	}
	bufio.NewWriter(writer).Flush()
}

func printJavaJaxbMain(rootElementName string, javaDir string, javaPackage string, sourceXMLFilename string, date time.Time) {
	t := template.Must(template.New("chidleyJaxbGenClass").Parse(jaxbMainTemplate))
	writer, f, err := javaClassWriter(javaDir, javaPackage, "Main")
	defer f.Close()

	classInfo := JaxbMainClassInfo{
		PackageName:       javaPackage,
		BaseXMLClassName:  rootElementName,
		SourceXMLFilename: sourceXMLFilename,
		Date:              date,
	}
	err = t.Execute(writer, classInfo)
	if err != nil {
		log.Println("executing template:", err)
	}
	bufio.NewWriter(writer).Flush()

}

func makeSourceReader(sourceName string, url bool, standardIn bool) (Source, error) {
	var err error

	var source Source
	if url {
		source = new(UrlSource)
		if DEBUG {
			log.Print("Making UrlSource")
		}
	} else {
		if standardIn {
			source = new(StdinSource)
			if DEBUG {
				log.Print("Making StdinSource")
			}
		} else {
			source = new(FileSource)
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

func capitalizeFirstLetter(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
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

// func makeOneLevelDown(node *Node) []*XMLType {
// 	var children []*XMLType

// 	for _, np := range node.Children {
// 		if np == nil {
// 			continue
// 		}
// 		for _, n := range np.Children {
// 			if n == nil {
// 				continue
// 			}
// 			x := XMLType{NameType: n.MakeType(namePrefix, nameSuffix),
// 				XMLName:      n.Name,
// 				XMLNameUpper: capitalizeFirstLetter(n.Name),
// 				XMLSpace:     n.Space}
// 			children = append(children, &x)
// 		}
// 	}
// 	return children
// }

func printChildrenChildren(node *Node) {
	for k, v := range node.Children {
		log.Print(k)
		log.Printf("children: %+v\n", v.Children)
	}
}

// Order Xml is encountered
func printStructsByXml(v *PrintGoStructVisitor) {
	orderNodes := make(map[int]*Node)
	var order []int

	for k := range v.AlreadyVisitedNodes {
		nodeOrder := v.AlreadyVisitedNodes[k].DiscoveredOrder
		orderNodes[nodeOrder] = v.AlreadyVisitedNodes[k]
		order = append(order, nodeOrder)
	}
	sort.Ints(order)

	for o := range order {
		print(v, orderNodes[o])
	}
}

// Alphabetical order
func printStructsAlphabetical(v *PrintGoStructVisitor) {
	var keys []string
	for k := range v.AlreadyVisitedNodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		print(v, v.AlreadyVisitedNodes[k])
	}

}
