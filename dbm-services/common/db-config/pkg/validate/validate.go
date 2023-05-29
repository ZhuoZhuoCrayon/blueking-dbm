// Package validate TODO
package validate

import (
	"bk-dbconfig/pkg/util"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/pkg/errors"
)

// ValidateEnums TODO
// make validate tag work with enums tag
// 避免 validate oneof 和 swagger enums 写 2 份重复的校验和文档
// example: Method string `validate:"required,enums" enums:"post,get" json:"method"`
func ValidateEnums(f validator.FieldLevel) bool {
	fieldValue := f.Field().String()
	fieldName := f.StructFieldName()
	// get StructField
	sf, _ := f.Parent().Type().FieldByName(fieldName)
	// get tag value from tag_field enums
	tagValue := sf.Tag.Get(TagEnum)
	enumsValues := strings.Split(tagValue, ",")
	if util.StringsHas(enumsValues, fieldValue) {
		return true
	} else {
		return false
	}
}

// GoValidateStructSimple TODO
// 简单校验 struct，不涉及逻辑
// 如果 struct 上有 tag validate:"enums"，必须启用enum=true校验
func GoValidateStructSimple(v interface{}, enum bool) error {
	validate := validator.New()
	if enum {
		_ = validate.RegisterValidation("enums", ValidateEnums)
	}
	if err := validate.Struct(v); err != nil {
		return err
	}
	return nil
}

// TagEnum TODO
const TagEnum = "enums"

// GoValidateStruct v 不能是Ptr
func GoValidateStruct(v interface{}, enum bool) error {
	validate := validator.New()
	uni := ut.New(en.New())
	trans, _ := uni.GetTranslator("en")
	// 提示时显示 json 字段的名字
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// name := fld.Tag.Get("json")
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return err
	}

	if enum {
		_ = validate.RegisterValidation(TagEnum, ValidateEnums)
	}
	if err := validate.Struct(v); err != nil {
		return translateErr2Msg(v, trans, err)
	}
	return nil
}

// translateErr2Msg v 不能是Ptr
func translateErr2Msg(v interface{}, trans ut.Translator, err error) error {
	var errStr []string
	for _, vErr := range err.(validator.ValidationErrors) {
		if vErr.Tag() == TagEnum {
			errmsg := ""
			// errmsg := customEnumTransFunc(vErr, v)
			if vErr.Param() == "" {
				sf, _ := reflect.TypeOf(v).FieldByName(vErr.StructField())
				tagValue := sf.Tag.Get(TagEnum)
				errmsg = fmt.Sprintf("%s must be one of [%s]", vErr.Field(), tagValue)
			} else {
				errmsg = vErr.Param()
			}
			errStr = append(errStr, errmsg)
			continue
		}
		errStr = append(errStr, vErr.Translate(trans))
	}
	return errors.New(strings.Join(errStr, " || "))
}
func customEnumTransFunc(fe validator.FieldError, v interface{}) string {
	if fe.Param() == "" {
		sf, _ := reflect.TypeOf(v).FieldByName(fe.StructField())
		tagValue := sf.Tag.Get(TagEnum)
		errmsg := fmt.Sprintf("%s must be one of [%s]", fe.Field(), tagValue)
		return errmsg
	} else {
		return fe.Param()
	}
}

// registerTranslator 为自定义字段添加翻译功能
func registerTranslator(tag string, msg string) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		if err := trans.Add(tag, msg, false); err != nil {
			return err
		}
		return nil
	}
}

// customTransFunc TODO
// translate 自定义字段的翻译方法
func customTransFunc(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field())
	if err != nil {
		panic(fe.(error).Error())
	}
	return msg
}

func translate(ut ut.Translator, fe validator.FieldError) string {
	s, err := ut.T(fe.Tag(), fe.Field(), "fe.Param()")
	if err != nil {
		log.Printf("warning: error translating FieldError: %#v", fe)
		return fe.(error).Error()
	}
	return s
}
