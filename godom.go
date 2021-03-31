package godom

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func startWith(str, needle string) bool {
	if len(str) < len(needle) {
		return false
	}
	return strings.ToLower(str[:len(needle)]) == needle
}
func findTokenNoQuote(str, token string) int {
	if len(str) < len(token) {
		return -1
	}
	for i := 0; i < len(str)-len(token); i++ {
		if startWith(str[i:], token) {
			return i
		}
		switch c := str[i]; c {
		case '"', '\'', '`':
			for i++; i < len(str); i++ {
				if str[i] == c && str[i-1] != '\\' {
					break
				}
			}
		}

	}
	return -1

}

type Node struct {
	TagContent string
	Nodes      []*Node
	Attr       map[string]string
	Index      map[string][]*Node
	Closed     bool
}

func NewNode(s string) *Node {
	n := &Node{
		TagContent: s,
		Nodes:      []*Node{},
		Attr:       map[string]string{},
		Index:      map[string][]*Node{},
	}
		n.updateAttribute()

	return n
}

func (n *Node) updateIndex() {
    n.Index=map[string][]*Node{}
    for _,child:=range n.Nodes {
		for ciKey,ciChild := range child.Index {
			n.Index[ciKey]=append(n.Index[ciKey],ciChild...)
		}
		if val,ok:=child.Attr["id"];ok {
			n.Index["#"+val]=[]*Node{child}
		}
		n.Index[child.GetTagName()] =append(n.Index[child.GetTagName()] ,child)
		if val,ok:=child.Attr["class"];ok {
			for _,class := range(strings.Split(val," ")) {
				if class !="" {
					n.Index["."+class] =append(n.Index["."+class],child)
				}
			}
		}
	}
}
func (n *Node) isTag() bool {
	return startWith(n.TagContent, "<") && !startWith(n.TagContent, "<!--")
}
func (n *Node) updateAttribute() {
	if n.isTag() {
		work:=n.TagContent[1:len(n.TagContent)-1]
		for i:=findTokenNoQuote(strings.ReplaceAll(work,"\n"," ")," ");i!=-1;i=findTokenNoQuote(strings.ReplaceAll(work,"\n"," ")," ") {
			j:=findTokenNoQuote(work[i+1:]+"  "," ")
			k:=findTokenNoQuote(work[i+1:i+j+1],"=")
			key := strings.TrimSpace(work[i+1:i+j+1])
			value := ""
			if k!=-1 {
				value=key[k+1:]
				if value[0] == '\'' || value[0] == '"' {
					value=value[1:len(value)-1]
				}
				key=key[:k]
			}
			if strings.TrimSpace(key) != "" {
				n.Attr[key]=value
			}
			work = work[i+1:]
		}
	}
}
func (n *Node) GetTagName() string {
	if !n.isTag() {
		return ""
	}
	var i int
	for i = 1; i < len(n.TagContent); i++ {
		if n.TagContent[i] == ' ' || n.TagContent[i] == '>' {
			break
		}
	}
	return strings.ToLower(n.TagContent[1:i])

}
func (n *Node) Selector(sel string) []*Node {
	return n.Index[sel]
}
func (n *Node) InnerContent() string {
	s:=""
	for _, an := range n.Nodes {
		s += an.GetContent()
	}
	return s
}

func (n *Node) GetContent() string {

	s := n.TagContent
	for _, an := range n.Nodes {
		s += an.GetContent()
	}
	if n.Closed {
		return s + fmt.Sprintf("</%s>", n.GetTagName())
	}
	return s
}
func (n *Node) GetStruct() string {
	bs, _ := json.MarshalIndent(n, "", " ")
	return string(bs)
}

func DocumentParser(source string) *Node {
	t := time.Now()
	defer func() {
		fmt.Println(time.Now().Sub(t))
	}()
	fmt.Println("OK")

	start := 0
	doc := NewNode("")
	lst := doc
	addToken := func(str string) {
		if str != "" {
			if startWith(str, "</") {
				tagName := str[2 : len(str)-1]
				for i := len(lst.Nodes) - 1; i >= 0; i-- {
					if !lst.Nodes[i].Closed && lst.Nodes[i].GetTagName() == tagName {
						lst.Nodes[i].Closed = true
						lst.Nodes[i].Nodes = append(lst.Nodes[i].Nodes, lst.Nodes[i+1:]...)
						lst.Nodes = append([]*Node{}, lst.Nodes[:i+1]...)
						lst.Nodes[i].updateIndex()
						break
					}
				}
			} else {
				lst.Nodes = append(lst.Nodes, NewNode(str))
			}
		}
	}
	for i := 0; i < len(source); i++ {
		switch c := source[i]; c {
		case '<':
			addToken(source[start:i])
			start = i
			if startWith(source[start:], "<!--") {
				ni := strings.Index(source[start:], "-->")
				if ni == -1 {
					log.Fatal("bug")
				}
				i += ni + 4
				addToken(source[start:i])
				start = i
			}
		case '>':
			addToken(source[start : i+1])
			ls := start
			start = i + 1
			if startWith(strings.ToLower(source[ls:start]), "<script") {
				iscript := findTokenNoQuote(source[start:], "</script>")
				i += iscript
			}

		}
	}
	addToken(source[start:])
	lst.updateAttribute()
	lst.updateIndex()
	return lst

}
