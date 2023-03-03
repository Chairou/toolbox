package cal

import "testing"

// go test -v calDate_test.go calDate.go calIP_test.go calIP.go

func TestSubNetMaskToLen(t *testing.T) {
	mask := "255.255.255.0"
	len, err := SubNetMaskToLen(mask)
	if err != nil {
		t.Error("SubNetMaskToLen err:", err)
	}
	t.Log(len)
}

func TestGetCidrIpRange(t *testing.T) {
	ipmask := "192.168.1.2/24"
	first, broadcast := GetCidrIpRange(ipmask)
	t.Log(first, broadcast)
	if first != "192.168.1.1" && broadcast != "192.168.1.255" {
		t.Error("GetCidrIpRange err:", first, broadcast)
	}
}
