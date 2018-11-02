package chidleystein

import (
	"strconv"
	"strings"
)

type NodeTypeInfo struct {
	alwaysBool    bool
	alwaysFloat32 bool
	alwaysFloat64 bool

	alwaysInt0  bool
	alwaysInt08 bool
	alwaysInt16 bool
	alwaysInt32 bool
	alwaysInt64 bool

	alwaysUint08 bool
	alwaysUint16 bool
	alwaysUint32 bool
	alwaysUint64 bool
}

func (nti *NodeTypeInfo) initialize() {
	nti.alwaysBool = true
	nti.alwaysFloat32 = true
	nti.alwaysFloat64 = true

	nti.alwaysInt0 = true
	nti.alwaysInt08 = true
	nti.alwaysInt16 = true
	nti.alwaysInt32 = true
	nti.alwaysInt64 = true

	nti.alwaysUint08 = true
	nti.alwaysUint16 = true
	nti.alwaysUint32 = true
	nti.alwaysUint64 = true
}

func (n *NodeTypeInfo) checkFieldType(v string) {
	v = strings.TrimSpace(v)

	if _, err := strconv.ParseBool(v); err != nil {
		n.alwaysBool = false
	}

	if _, err := strconv.ParseFloat(v, 32); err != nil {
		n.alwaysFloat32 = false
	}

	if _, err := strconv.ParseFloat(v, 64); err != nil {
		n.alwaysFloat64 = false
	}

	if _, err := strconv.ParseInt(v, 10, 0); err != nil {
		n.alwaysInt0 = false
	}

	if _, err := strconv.ParseInt(v, 10, 8); err != nil {
		n.alwaysInt08 = false
	}

	if _, err := strconv.ParseInt(v, 10, 16); err != nil {
		n.alwaysInt16 = false
	}

	if _, err := strconv.ParseInt(v, 10, 32); err != nil {
		n.alwaysInt32 = false
	}

	if _, err := strconv.ParseInt(v, 10, 64); err != nil {
		n.alwaysInt64 = false
	}

	if _, err := strconv.ParseUint(v, 10, 8); err != nil {
		n.alwaysUint08 = false
	}

	if _, err := strconv.ParseUint(v, 10, 16); err != nil {
		n.alwaysUint16 = false
	}

	if _, err := strconv.ParseUint(v, 10, 32); err != nil {
		n.alwaysUint32 = false
	}

	if _, err := strconv.ParseUint(v, 10, 64); err != nil {
		n.alwaysUint64 = false
	}

}
