package xstorage

import (
	"errors"
	"fmt"
	"github.com/intmian/mian_go_lib/tool/misc"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

/*
初始化时传入数据库地址和表名
列Key用来存储键
列ValueInt用来存储整数值
列ValueString用来存储字符串值
...
请注意假如value类型为slice，会被存储于key[0]、key[1]、key[2]...列中，key[0]、key[1]、key[2]...列的值为value的每个元素
*/

type SqliteCore struct {
	db *gorm.DB
	misc.InitTag
}

func NewSqliteCore(DbFileAddr string) (*SqliteCore, error) {
	db, err := gorm.Open(sqlite.Open(DbFileAddr), &gorm.Config{})
	if err != nil {
		return nil, errors.Join(errors.New("open sqlite error"), err)
	}
	sqliteCore := &SqliteCore{
		db: db,
	}
	sqliteCore.SetInitialized()
	return sqliteCore, nil
}

func (m *SqliteCore) Get(key string) (bool, *ValueUnit, error) {
	if !m.IsInitialized() {
		return false, nil, errors.New("sqlite core not init")
	}
	var keyValueModel KeyValueModel
	result := m.db.Where("key = ?", key).First(&keyValueModel)
	// 如果没有这个

	if result.Error != nil {
		return false, nil, errors.Join(errors.New("get value error"), result.Error)
	}

	sliceNum, valueUnit, err := sqliteModel2Data(keyValueModel)
	if err != nil {
		return false, valueUnit, err
	}

	if sliceNum == 0 {
		return false, valueUnit, nil
	}

	// slice 的内容存放在 key[0]、key[1]、key[2]...key[sliceNum-1] 列中
	switch ValueType(keyValueModel.valueType) {
	case VALUE_TYPE_SLICE_INT:
		valueUnit.Data = make([]int, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_INT
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.valueInt == nil {
				return false, nil, fmt.Errorf("slice but valueInt is nil, key: %s[%d]", key, i)
			}
			valueUnit.Data.([]int)[i] = *keyValueModel.valueInt
		}
	case VALUE_TYPE_SLICE_STRING:
		valueUnit.Data = make([]string, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_STRING
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.valueString == nil {
				return false, nil, fmt.Errorf("slice but valueString is nil, key: %s[%d]", key, i)
			}
			valueUnit.Data.([]string)[i] = *keyValueModel.valueString
		}
	case VALUE_TYPE_SLICE_FLOAT:
		valueUnit.Data = make([]float32, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_FLOAT
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.valueFloat == nil {
				return false, nil, fmt.Errorf("slice but valueFloat is nil, key: %s[%d]", key, i)
			}
			valueUnit.Data.([]float32)[i] = *keyValueModel.valueFloat
		}
	case VALUE_TYPE_SLICE_BOOL:
		valueUnit.Data = make([]bool, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_BOOL
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.valueInt == nil {
				return false, nil, fmt.Errorf("slice but valueInt is nil, key: %s[%d]", key, i)
			}
			if *keyValueModel.valueInt == 0 {
				valueUnit.Data.([]bool)[i] = false
			} else {
				valueUnit.Data.([]bool)[i] = true
			}
		}
	}

	return false, valueUnit, nil
}

// sqliteModel2Data 将从数据库取出来的model转化为ValueUnit，但是需要注意的是，如果是slice类型，只返回slice的长度，不返回具体的值
func sqliteModel2Data(keyValueModel KeyValueModel) (int, *ValueUnit, error) {
	var value *ValueUnit
	sliceNum := 0
	// 判断合法性
	switch ValueType(keyValueModel.valueType) {
	case VALUE_TYPE_INT, VALUE_TYPE_BOOL:
		if keyValueModel.valueInt == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_STRING:
		if keyValueModel.valueString == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_FLOAT:
		if keyValueModel.valueFloat == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_SLICE_INT, VALUE_TYPE_SLICE_STRING, VALUE_TYPE_SLICE_FLOAT, VALUE_TYPE_SLICE_BOOL:
		if keyValueModel.valueInt == nil {
			return 0, nil, errors.New("slice but valueInt is nil")
		}
	default:
		return 0, nil, errors.New("value type error")
	}

	// 读取值
	switch ValueType(keyValueModel.valueType) {
	case VALUE_TYPE_INT:
		value.Data = *keyValueModel.valueInt
		value.Type = VALUE_TYPE_INT
	case VALUE_TYPE_BOOL:
		if (*keyValueModel.valueInt) == 0 {
			value.Data = false
		} else {
			value.Data = true
		}
		value.Type = VALUE_TYPE_BOOL
	case VALUE_TYPE_STRING:
		value.Data = *keyValueModel.valueString
		value.Type = VALUE_TYPE_STRING
	case VALUE_TYPE_FLOAT:
		value.Data = *keyValueModel.valueFloat
		value.Type = VALUE_TYPE_FLOAT
	case VALUE_TYPE_SLICE_INT, VALUE_TYPE_SLICE_STRING, VALUE_TYPE_SLICE_FLOAT, VALUE_TYPE_SLICE_BOOL:
		sliceNum = *keyValueModel.valueInt
		if sliceNum <= 0 {
			return 0, nil, fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
	}
	return sliceNum, value, nil
}

func (m *SqliteCore) Set(key string, value *ValueUnit) error {
	if !m.IsInitialized() {
		return errors.New("sqlite core not init")
	}

	// 为避免GetAll时，取出slice的成员作为单独的主键，这里不允许key中包含[]
	if strings.Contains(key, "[") || strings.Contains(key, "]") {
		return errors.New("key can not contain []")
	}

	keyValueModel := &KeyValueModel{
		key:       &key,
		valueType: int(value.Type),
	}
	switch value.Type {
	case VALUE_TYPE_INT:
		valueInt := Get[int](value)
		keyValueModel.valueInt = &valueInt
	case VALUE_TYPE_BOOL:
		valueBool := Get[bool](value)
		if valueBool {
			valueInt := 1
			keyValueModel.valueInt = &valueInt
		} else {
			valueInt := 0
			keyValueModel.valueInt = &valueInt
		}
	case VALUE_TYPE_STRING:
		valueString := Get[string](value)
		keyValueModel.valueString = &valueString
	case VALUE_TYPE_FLOAT:
		valueFloat := Get[float32](value)
		keyValueModel.valueFloat = &valueFloat
	case VALUE_TYPE_SLICE_INT, VALUE_TYPE_SLICE_STRING, VALUE_TYPE_SLICE_FLOAT, VALUE_TYPE_SLICE_BOOL:
		sliceNum := len(value.Data.([]int))
		if sliceNum <= 0 {
			return fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
		valueInt := sliceNum
		keyValueModel.valueInt = &valueInt
	}
	result := m.db.Create(keyValueModel)
	if result.Error != nil {
		return errors.Join(errors.New("set value error"), result.Error)
	}

	// slice 的内容存放在 key[0]、key[1]、key[2]...key[sliceNum-1] 列中
	var sliceErr error
	switch value.Type {
	case VALUE_TYPE_SLICE_INT:
		for i, v := range Get[[]int](value) {
			sliceErr = m.Set(key+"["+strconv.Itoa(i)+"]", &ValueUnit{
				Type: VALUE_TYPE_INT,
				Data: v,
			})
		}
	case VALUE_TYPE_SLICE_STRING:
		for i, v := range Get[[]string](value) {
			sliceErr = m.Set(key+"["+strconv.Itoa(i)+"]", &ValueUnit{
				Type: VALUE_TYPE_STRING,
				Data: v,
			})
		}
	case VALUE_TYPE_SLICE_FLOAT:
		for i, v := range Get[[]float32](value) {
			sliceErr = m.Set(key+"["+strconv.Itoa(i)+"]", &ValueUnit{
				Type: VALUE_TYPE_FLOAT,
				Data: v,
			})
		}
	case VALUE_TYPE_SLICE_BOOL:
		for i, v := range Get[[]bool](value) {
			sliceErr = m.Set(key+"["+strconv.Itoa(i)+"]", &ValueUnit{
				Type: VALUE_TYPE_BOOL,
				Data: v,
			})
		}
	}
	if sliceErr != nil {
		// 删除前面的脏数据
		m.db.Where("key = ?", key).Delete(&KeyValueModel{})
		m.db.Where("key like ?", key+"%").Delete(&KeyValueModel{})
		return errors.Join(errors.New("set slice value error"), sliceErr)
	}
	return nil
}

func (m *SqliteCore) Delete(key string) error {
	if !m.IsInitialized() {
		return errors.New("sqlite core not init")
	}
	return m.db.Where("key = ?", key).Delete(&KeyValueModel{}).Error
}

func (m *SqliteCore) Have(key string) (bool, error) {
	if !m.IsInitialized() {
		return false, errors.New("sqlite core not init")
	}
	var keyValueModel KeyValueModel
	result := m.db.Where("key = ?", key).First(&keyValueModel)
	if result.Error != nil {
		return false, errors.Join(errors.New("get value error"), result.Error)
	}
	return true, nil
}

func (m *SqliteCore) GetAll() (map[string]*ValueUnit, error) {
	if !m.IsInitialized() {
		return nil, errors.New("sqlite core not init")
	}
	var keyValueModelList []KeyValueModel
	result := m.db.Find(&keyValueModelList)
	if result.Error != nil {
		return nil, errors.Join(errors.New("get all value error"), result.Error)
	}
	keyValueModelMap := make(map[string]*ValueUnit)
	for _, keyValueModel := range keyValueModelList {
		// 跳过所有含有[]的key，因为这些key是slice的成员，不是真正的key
		if strings.Contains(*keyValueModel.key, "[") || strings.Contains(*keyValueModel.key, "]") {
			continue
		}
		sliceNum, unit, err := sqliteModel2Data(keyValueModel)
		if err != nil {
			return nil, errors.Join(errors.New("sqliteModel2Data"), err)
		}

		if sliceNum != 0 {
			// slice 的内容存放在 key[0]、key[1]、key[2]...key[sliceNum-1] 列中
			_, unit, err = m.Get(*keyValueModel.key)
			if err != nil {
				return nil, errors.Join(errors.New("get slice value error"), err)
			}
		}
		keyValueModelMap[*keyValueModel.key] = unit
	}
	return keyValueModelMap, nil
}
