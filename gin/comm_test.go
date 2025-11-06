package gin

import (
	"testing"
)

type NestedStruct struct {
	City    string `gorm:"column:city;default:1" json:"City"`
	Country string `gorm:"column:country;default:1" json:"Country"`
}

type SourceStruct struct {
	Name     string        `gorm:"column:name1;default:1" comment:"0 禁用，1 启用" json:"Name"`
	Age      int           `gorm:"column:age1;default:1" comment:"0 禁用，1 启用" json:"Age"`
	Address  *string       `gorm:"column:address1;default:1" comment:"0 禁用，1 启用" json:"Address"`
	Height   *int          `gorm:"column:height1;default:1" comment:"0 禁用，1 启用" json:"Height"`
	Location NestedStruct  `gorm:"column:location;default:1" json:"Location"`
	Profile  *NestedStruct `gorm:"column:profile;default:1" json:"Profile"`
}

func TestDeepCopyToMap(t *testing.T) {
	// 测试用例1：包含所有字段（包括嵌套结构体）
	src1 := SourceStruct{
		Name: "xiaoming",
		Age:  18,
		Address: func() *string {
			s := "beijing"
			return &s
		}(),
		Height: func() *int {
			i := 180
			return &i
		}(),
		Location: NestedStruct{
			City:    "Beijing",
			Country: "China",
		},
		Profile: &NestedStruct{
			City:    "Shanghai",
			Country: "China",
		},
	}
	
	// 测试用例2：部分字段为nil（包括嵌套结构体指针为nil）
	src2 := SourceStruct{}
	src2.Name = "xiaohong"
	src2.Age = 20
	src2.Height = func() *int {
		i := 181
		return &i
	}()
	src2.Location = NestedStruct{
		City:    "Guangzhou",
		Country: "China",
	}
	// src2.Profile 为 nil，应该被跳过
	
	dst := make(map[string]interface{})
	dst2 := make(map[string]interface{})
	
	err := GormStructToMap(&src1, &dst, "column")
	if err != nil {
		t.Error(err)
	}
	t.Logf("src1 result: %+v", dst)
	
	err = GormStructToMap(&src2, &dst2, "column")
	if err != nil {
		t.Error(err)
	}
	t.Logf("src2 result: %+v", dst2)
}

func TestNewGormTagParser(t *testing.T) {
	parser := NewGormTagParser(`gorm:"type:datetime(3);column:updateTime;default:CURRENT_TIMESTAMP(3);autoUpdateTime" json:"updateTime"`)
	column, _ := parser.Get("column")
	t.Logf("%+v", column)
}
