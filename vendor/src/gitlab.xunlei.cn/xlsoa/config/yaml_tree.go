package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"strconv"
	"strings"
)

const (
	ROOT = ""
)

// Node Type
type nodeType int

const (
	valueNode nodeType = iota
	objectNode
	arrayNode
)

func (t nodeType) String() string {
	v, ok := nodeTypeString[t]
	if ok {
		return v
	} else {
		return "Unknown"
	}
}

var nodeTypeString = map[nodeType]string{
	valueNode:  "Value",
	objectNode: "Object",
	arrayNode:  "Array",
}

func getNodeType(v interface{}) nodeType {
	switch v.(type) {
	case map[interface{}]interface{}:
		return objectNode
	case []interface{}:
		return arrayNode
	default:
		return valueNode
	}
}

// Yaml tree
type yamlTree struct {
	root *yamlNode
	data []byte
}

func newYamlTreeFromByte(data []byte) *yamlTree {
	return &yamlTree{
		data: data,
	}
}
func newYamlTreeFromYamlNode(node *yamlNode) *yamlTree {
	return &yamlTree{
		root: node,
	}
}

func (t *yamlTree) String() string {
	return fmt.Sprintf("YamlTree { root: %v }", t.root)
}

func (t *yamlTree) init() error {
	if t.data != nil {
		var err error

		var out interface{}
		err = yaml.Unmarshal(t.data, &out)
		if err != nil {
			return errors.Wrap(err, "yamlTree.Init(): yaml.Unmarshal fail")
		}

		t.root = newYamlNode("root", out)

		t.data = nil // Release for gc
	}
	return nil
}

func (t *yamlTree) populate(out interface{}) error {
	return t.root.populate(out)
}

func (t *yamlTree) toYamlByte() ([]byte, error) {
	var err error

	var out interface{}
	err = t.root.populate(&out)
	if err != nil {
		return nil, errors.Wrap(err, "Populate from root fail")
	}

	var data []byte
	data, err = yaml.Marshal(out)
	if err != nil {
		return nil, errors.Wrap(err, "yaml.Marshal out error")
	}

	return data, nil
}

// *Value
//    nil: Not exists, when error==nil
//    !nil: Success, when error==nil
//
// error:
//    nil: ok
//    !nil: fail
func (t *yamlTree) get(key string) (*Value, error) {

	n := t.root.get(key)
	if n == nil {
		return nil, nil
	}

	// Copy a new sub-tree to avoid race condition.
	var err error
	var copyNode *yamlNode
	copyNode, err = n.copy()
	if err != nil {
		return nil, errors.Wrap(err, "yamlTree.Get(): Copy node fail")
	}

	subTree := newYamlTreeFromYamlNode(copyNode)
	subTree.init()
	return newValue(subTree), nil
}

// Yaml Node
type yamlNode struct {
	key      string
	value    interface{}
	nodeType nodeType
}

func newYamlNode(key string, value interface{}) *yamlNode {
	return &yamlNode{
		key:      key,
		value:    value,
		nodeType: getNodeType(value),
	}
}

// Deep copy
func (n *yamlNode) copy() (*yamlNode, error) {

	var out interface{}

	err := n.populate(&out)
	if err != nil {
		return nil, errors.Wrap(err, "yamlNode.Copy(): node.Unmarshal() fail")
	}

	return newYamlNode(n.key, out), nil
}

func (n *yamlNode) String() string {
	return fmt.Sprintf("YamlNode{ Key: '%v', NodeType: '%v', Value: '%v' }", n.key, n.nodeType, n.value)
}

// Get(ROOT): Return root node 'n'.
// Get("server.name"): Get "server.name".
func (n *yamlNode) get(key string) *yamlNode {

	if key == ROOT {
		return n
	}

	var node = n
	parts := strings.Split(key, ".")
	for {
		if len(parts) == 0 {
			return node
		}
		part := parts[0]

		children := node.children()
		if len(children) == 0 {
			break
		}

		found := false
		for _, child := range children {
			if child.key == part {
				node = child
				found = true
				break
			}
		}

		if !found {
			break
		}

		parts = parts[1:]
	}

	return nil
}

func (n *yamlNode) populate(out interface{}) error {
	var err error
	var data []byte
	data, err = yaml.Marshal(n.value)
	if err != nil {
		return errors.Wrap(err, "yamlNode.Unmarshal(): yaml.Marshal n.value fail")
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return errors.Wrap(err, "yamlNode.Unmarshal(): yaml.Unmarshal fail")
	}
	return nil
}

func (n *yamlNode) children() []*yamlNode {
	nodes := make([]*yamlNode, 0)

	// Case valueNode, there is no children.
	switch n.nodeType {
	case objectNode:
		for k, v := range n.value.(map[interface{}]interface{}) {
			nn := newYamlNode(k.(string), v)
			nodes = append(nodes, nn)
		}
	case arrayNode:
		for i, v := range n.value.([]interface{}) {
			nn := newYamlNode(strconv.Itoa(i), v)
			nodes = append(nodes, nn)
		}

	}

	return nodes
}
