package qcloud

import (
	"testing"
)

func TestQcloudBillList(t *testing.T) {
	args := BillArg{}
	args.SecretId = "XXX-Secret-"
	args.SecretKey = "XXX-Secret"
	args.BeginTime = "2023-01-01 00:00:00"
	args.EndTime = "2023-01-20 00:00:00"
	args.Region = "ap-hongkong"
	args.Offset = 0
	args.Limit = 100
	t.Log(DescribeBillDetail(args))
}
