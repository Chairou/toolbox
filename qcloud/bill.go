package qcloud

import (
	"fmt"
	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

type BillArg struct {
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region"`
	BeginTime string `json:"BeginTime"`
	EndTime   string `json:"endTime"`
	Offset    uint64 `json:"offset"`
	Limit     uint64 `json:"limit"`
}

func DescribeBillDetail(args BillArg) string {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		args.SecretId,
		args.SecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "billing.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := billing.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := billing.NewDescribeBillDetailRequest()
	request.Offset = common.Uint64Ptr(args.Offset)
	request.Limit = common.Uint64Ptr(args.Limit)
	request.BeginTime = common.StringPtr(args.BeginTime)
	request.EndTime = common.StringPtr(args.EndTime)
	// 返回的resp是一个DescribeBillDetailResponse的实例，与请求对象对应
	response, err := client.DescribeBillDetail(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return err.Error()
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	return response.ToJsonString()
}
