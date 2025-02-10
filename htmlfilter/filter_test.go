package htmlfilter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestNode_Find(t *testing.T) {
	//<ul><li class="foo"><ul><li class="foo qux"></li></ul></li><li class="foo bar"></li></ul>
	fragmentHtml := &html.Node{DataAtom: atom.Ul, Type: html.ElementNode}
	l00 := &html.Node{
		Type:     html.ElementNode,
		Parent:   fragmentHtml,
		DataAtom: atom.Li,
		Attr:     []html.Attribute{{Key: "class", Val: "foo"}},
	}
	l01 := &html.Node{
		Type:     html.ElementNode,
		Parent:   fragmentHtml,
		DataAtom: atom.Li,
		Attr:     []html.Attribute{{Key: "class", Val: "foo bar"}},
	}
	l00.NextSibling = l01
	l01.PrevSibling = l00
	fragmentHtml.FirstChild = l00
	fragmentHtml.LastChild = l01

	r1 := &html.Node{DataAtom: atom.Ul, Parent: l00, Type: html.ElementNode}
	l10 := &html.Node{
		Type:     html.ElementNode,
		Parent:   r1,
		DataAtom: atom.Li,
		Attr:     []html.Attribute{{Key: "class", Val: "foo qux"}},
	}
	r1.FirstChild = l10
	r1.LastChild = l10
	l00.FirstChild = r1
	l00.LastChild = r1

	fragment := Node{fragmentHtml}

	tests := map[string]struct {
		fragment Node
		selector string
		want     []Node
	}{
		"found two nodes at zero layer": {
			fragment: fragment,
			selector: "li.foo",
			want:     []Node{{l00}, {l01}},
		},
		"found single node found at zero layer": {
			fragment: fragment,
			selector: "li.foo.bar",
			want:     []Node{{l01}},
		},
		"found single node at second layer": {
			fragment: fragment,
			selector: "li.qux",
			want:     []Node{{l10}},
		},
		"no nodes found": {
			fragment: fragment,
			selector: "div.foo",
		},
		"found two nodes by tag only": {
			fragment: fragment,
			selector: "li",
			want:     []Node{{l00}, {l01}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []Node
			for v := range test.fragment.Find(test.selector) {
				got = append(got, v)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_readSelector(t *testing.T) {
	tests := []struct {
		name                   string
		selector               string
		wantElementAtom        atom.Atom
		wantAttrKeyRef         string
		attrValMiss            []string
		attrValHitCombinations []string
	}{
		{
			name:            "class: two vals",
			selector:        "div.foo.bar",
			wantElementAtom: atom.Div,
			wantAttrKeyRef:  "class",
			attrValMiss:     []string{"foo", "bar", "foo qux", "qux"},
			attrValHitCombinations: []string{
				"foo bar", "bar foo",
				"foo bar qux", "qux foo bar", "bar qux foo",
			},
		},
		{
			name:            "class: single val",
			selector:        "div.foo",
			wantElementAtom: atom.Div,
			wantAttrKeyRef:  "class",
			attrValMiss:     []string{"1", "qux"},
			attrValHitCombinations: []string{
				"foo",
				"foo bar", "bar foo",
				"foo bar qux", "qux foo bar", "bar qux foo",
				"foo qux",
			},
		},
		{
			name:                   "id",
			selector:               "div#foo",
			wantElementAtom:        atom.Div,
			wantAttrKeyRef:         "id",
			attrValMiss:            []string{"1", "qux"},
			attrValHitCombinations: []string{"foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotElementAtom, gotAttrKeyRef, gotAttrValFilter := readSelector(tt.selector)
			assert.Equal(t, tt.wantElementAtom, gotElementAtom)
			assert.Equal(t, tt.wantAttrKeyRef, gotAttrKeyRef)
			for _, v := range tt.attrValMiss {
				assert.False(t, gotAttrValFilter(v))
			}
			for _, v := range tt.attrValHitCombinations {
				assert.True(t, gotAttrValFilter(v))
			}
		})
	}
}
