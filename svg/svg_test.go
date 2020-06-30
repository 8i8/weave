package svg

import "testing"

func TestViewBoxToBytes(t *testing.T) {
	vb := ViewBox{minX: 234, minY: 232345092459, width: 3456, height: 3456}
	byt := vb.viewBoxToBytes()
	str1 := string(byt)
	str2 := "234 232345092459 3456 3456"
	if str1 != str2 {
		t.Errorf("error: viewBoxToByte: expected %s recieved %s", str2, str1)
	}
}
