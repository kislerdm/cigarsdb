package htmlfilter

import (
	"iter"
	"slices"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Node struct {
	*html.Node
}

// Find returns an iterator over the children of n the html search selector.
// The result will contain only the elements from the level of the first Node which matches the selection criteria.
//
// The search follows a combination of depth-first and breadth-first preorder: depth-first until firth occurrence is
// found, afterward, only its siblings are scanned.
//
// Class selector: div.class0.class1.
// Id selector: div#id.
func (n Node) Find(selector string) iter.Seq[Node] {
	atomFilter, attrKeyRef, attrValFilter := readSelector(selector)
	var parentOfFirstFound *Node
	for nn := range n.Descendants() {
		if parentOfFirstFound != nil {
			break
		}
		if nn.DataAtom == atomFilter {
			switch attrKeyRef != "" {
			case true:
				for _, att := range nn.Attr {
					if att.Key == attrKeyRef && attrValFilter(att.Val) {
						parentOfFirstFound = &Node{nn.Parent}
					}
				}

			default:
				parentOfFirstFound = &Node{nn.Parent}
			}
		}
	}
	return func(yield func(Node) bool) {
		if parentOfFirstFound != nil {
			for nn := range parentOfFirstFound.ChildNodes() {
				if nn.DataAtom == atomFilter {
					switch attrKeyRef != "" {
					case true:
						for _, att := range nn.Attr {
							if att.Key == attrKeyRef && attrValFilter(att.Val) && !yield(Node{nn}) {
								return
							}
						}

					default:
						if !yield(Node{nn}) {
							return
						}
					}
				}
			}
		}
	}
}

type selectorFn func(s string) bool

func classSelector(v []string) selectorFn {
	return func(s string) bool {
		var cnt int
		classVal := strings.Split(s, " ")
		for _, vv := range v {
			greedy := strings.HasSuffix(vv, "*")
			switch greedy {
			case true:
				vv = strings.TrimSuffix(vv, "*")
				if strings.Contains(s, vv) {
					cnt++
				}

			case false:
				if slices.Contains(classVal, vv) {
					cnt++
				}
			}
		}
		return len(v) == cnt
	}
}

func idSelector(v string) selectorFn {
	return func(s string) bool {
		return v == s
	}
}

func readSelector(s string) (elementAtom atom.Atom, attrKeyRef string, attrValFilter selectorFn) {
	idSplit := strings.SplitN(s, "#", 2)
	classSplit := strings.Split(s, ".")
	switch {
	case len(idSplit) == 2:
		elementAtom = atom.Lookup([]byte(idSplit[0]))
		attrKeyRef = "id"
		attrValFilter = idSelector(idSplit[1])
	case len(classSplit) > 1:
		elementAtom = atom.Lookup([]byte(classSplit[0]))
		attrKeyRef = "class"
		attrValFilter = classSelector(classSplit[1:])
	default:
		elementAtom = atom.Lookup([]byte(s))
	}
	if elementAtom == 0 {
		panic("unsupported selector provided")
	}
	return elementAtom, attrKeyRef, attrValFilter
}
