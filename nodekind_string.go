// Code generated by "stringer --type nodeKind"; DO NOT EDIT.

package effe

import "strconv"

const _nodeKind_name = "NodeKindFunctionNodeKindLiteralNodeKindOperatorNodeKindHole"

var _nodeKind_index = [...]uint8{0, 16, 31, 47, 59}

func (i nodeKind) String() string {
	if i < 0 || i >= nodeKind(len(_nodeKind_index)-1) {
		return "nodeKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _nodeKind_name[_nodeKind_index[i]:_nodeKind_index[i+1]]
}
