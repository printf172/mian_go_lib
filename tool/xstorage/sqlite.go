package xstorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/intmian/mian_go_lib/tool/misc"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
初始化时传入数据库地址和表名
列Key用来存储键
列ValueInt用来存储整数值
列ValueString用来存储字符串值
...
请注意假如value类型为slice，会被存储于key[0]、Key[1]、Key[2]...列中，Key[0]、Key[1]、Key[2]...列的值为value的每个元素
*/

type SqliteCore struct {
	db *gorm.DB
	misc.InitTag
	rwLock sync.RWMutex
}

type EmptyLogger struct {
}

func (e EmptyLogger) LogMode(level logger.LogLevel) logger.Interface {
	return e
}

func (e EmptyLogger) Info(ctx context.Context, s string, i ...interface{}) {}

func (e EmptyLogger) Warn(ctx context.Context, s string, i ...interface{}) {}

func (e EmptyLogger) Error(ctx context.Context, s string, i ...interface{}) {}

func (e EmptyLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
}

func NewSqliteCore(DbFileAddr string) (*SqliteCore, error) {
	// 依靠外层进行日志交互，为了避免本地打印日志过多，这里不使用日志库
	db, err := gorm.Open(sqlite.Open(DbFileAddr), &gorm.Config{Logger: EmptyLogger{}})
	if err != nil {
		return nil, errors.Join(errors.New("open sqlite error"), err)
	}
	err = db.AutoMigrate(&KeyValueModel{})
	if err != nil {
		return nil, errors.Join(errors.New("auto migrate error"), err)
	}
	sqliteCore := &SqliteCore{
		db: db,
	}
	sqliteCore.SetInitialized()
	return sqliteCore, nil
}

func (m *SqliteCore) Get(key string) (bool, *ValueUnit, error) {
	return m.GetInner(key, true)
}

func (m *SqliteCore) GetInner(key string, needLock bool) (bool, *ValueUnit, error) {
	if !m.IsInitialized() {
		return false, nil, errors.New("sqlite core not init")
	}
	if needLock {
		m.rwLock.RLock()
		defer m.rwLock.RUnlock()
	}

	var keyValueModel KeyValueModel
	result := m.db.Where("Key = ?", key).First(&keyValueModel)
	// 如果没有这个

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil, nil
		} else {
			return false, nil, errors.Join(errors.New("get value error"), result.Error)
		}
	}

	sliceNum, valueUnit, err := sqliteModel2Data(keyValueModel)
	if err != nil {
		return false, valueUnit, err
	}

	if sliceNum == 0 {
		return false, valueUnit, nil
	}

	// slice 的内容存放在 Key[0]、Key[1]、Key[2]...Key[sliceNum-1] 列中
	switch ValueType(keyValueModel.ValueType) {
	case VALUE_TYPE_SLICE_INT:
		valueUnit.Data = make([]int, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_INT
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("Key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.ValueInt == nil {
				return false, nil, fmt.Errorf("slice but ValueInt is nil, Key: %s[%d]", key, i)
			}
			valueUnit.Data.([]int)[i] = *keyValueModel.ValueInt
		}
	case VALUE_TYPE_SLICE_STRING:
		valueUnit.Data = make([]string, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_STRING
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("Key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.ValueString == nil {
				return false, nil, fmt.Errorf("slice but ValueString is nil, Key: %s[%d]", key, i)
			}
			valueUnit.Data.([]string)[i] = *keyValueModel.ValueString
		}
	case VALUE_TYPE_SLICE_FLOAT:
		valueUnit.Data = make([]float32, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_FLOAT
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("Key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.ValueFloat == nil {
				return false, nil, fmt.Errorf("slice but ValueFloat is nil, Key: %s[%d]", key, i)
			}
			valueUnit.Data.([]float32)[i] = *keyValueModel.ValueFloat
		}
	case VALUE_TYPE_SLICE_BOOL:
		valueUnit.Data = make([]bool, sliceNum)
		valueUnit.Type = VALUE_TYPE_SLICE_BOOL
		for i := 0; i < sliceNum; i++ {
			result := m.db.Where("Key = ?", key+"["+strconv.Itoa(i)+"]").First(&keyValueModel)
			if result.Error != nil {
				return false, nil, errors.Join(errors.New("get value error"), result.Error)
			}
			if keyValueModel.ValueInt == nil {
				return false, nil, fmt.Errorf("slice but ValueInt is nil, Key: %s[%d]", key, i)
			}
			if *keyValueModel.ValueInt == 0 {
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
	value := &ValueUnit{}
	sliceNum := 0
	// 判断合法性
	switch ValueType(keyValueModel.ValueType) {
	case VALUE_TYPE_INT, VALUE_TYPE_BOOL:
		if keyValueModel.ValueInt == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_STRING:
		if keyValueModel.ValueString == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_FLOAT:
		if keyValueModel.ValueFloat == nil {
			return 0, nil, errors.New("value is nil")
		}
	case VALUE_TYPE_SLICE_INT, VALUE_TYPE_SLICE_STRING, VALUE_TYPE_SLICE_FLOAT, VALUE_TYPE_SLICE_BOOL:
		if keyValueModel.ValueInt == nil {
			return 0, nil, errors.New("slice but ValueInt is nil")
		}
	default:
		return 0, nil, errors.New("value type error")
	}

	// 读取值
	switch ValueType(keyValueModel.ValueType) {
	case VALUE_TYPE_INT:
		value.Data = *keyValueModel.ValueInt
		value.Type = VALUE_TYPE_INT
	case VALUE_TYPE_BOOL:
		if (*keyValueModel.ValueInt) == 0 {
			value.Data = false
		} else {
			value.Data = true
		}
		value.Type = VALUE_TYPE_BOOL
	case VALUE_TYPE_STRING:
		value.Data = *keyValueModel.ValueString
		value.Type = VALUE_TYPE_STRING
	case VALUE_TYPE_FLOAT:
		value.Data = *keyValueModel.ValueFloat
		value.Type = VALUE_TYPE_FLOAT
	case VALUE_TYPE_SLICE_INT, VALUE_TYPE_SLICE_STRING, VALUE_TYPE_SLICE_FLOAT, VALUE_TYPE_SLICE_BOOL:
		sliceNum = *keyValueModel.ValueInt
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
		return errors.New("Key can not contain []")
	}

	var needCreate []*KeyValueModel
	var needSet []*KeyValueModel
	var needRemove []*KeyValueModel // slice缩短的情况

	keyValueModels, err := sqliteData2Model(key, value)
	if err != nil {
		return errors.Join(errors.New("sqliteData2Model"), err)
	}

	exist, dbValue, err := m.GetInner(key, false)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Join(errors.New("get value error"), err)
	}

	if !exist || dbValue.Type < VALUE_TYPE_SLICE_BEGIN {
		needCreate = keyValueModels
	} else {
		sliceNum := 0
		switch dbValue.Type {
		case VALUE_TYPE_SLICE_INT:
			sliceNum = len(ToBase[[]int](dbValue))
		case VALUE_TYPE_SLICE_STRING:
			sliceNum = len(ToBase[[]string](dbValue))
		case VALUE_TYPE_SLICE_FLOAT:
			sliceNum = len(ToBase[[]float32](dbValue))
		case VALUE_TYPE_SLICE_BOOL:
			sliceNum = len(ToBase[[]bool](dbValue))
		}
		for i, keyValueModel := range keyValueModels {
			if i == 0 {
				// 长度节点
				if *keyValueModel.ValueInt != sliceNum {
					needSet = append(needSet, keyValueModel)
				}
				continue
			}
			if i > sliceNum {
				needCreate = append(needCreate, keyValueModel)
				continue
			}
			// 判断值
			switch ValueType(keyValueModel.ValueType) {
			case VALUE_TYPE_INT:
				if *keyValueModel.ValueInt != ToBase[int](dbValue) {
					needSet = append(needSet, keyValueModel)
				}
			case VALUE_TYPE_BOOL:
				b := false
				if *keyValueModel.ValueInt != 0 {
					b = true
				}
				if b != ToBase[bool](dbValue) {
					needSet = append(needSet, keyValueModel)
				}
			case VALUE_TYPE_STRING:
				if *keyValueModel.ValueString != ToBase[string](dbValue) {
					needSet = append(needSet, keyValueModel)
				}
			case VALUE_TYPE_FLOAT:
				if *keyValueModel.ValueFloat != ToBase[float32](dbValue) {
					needSet = append(needSet, keyValueModel)
				}
			}
		}
		for i := len(keyValueModels) - 1; i < sliceNum; i++ {
			needRemove = append(needRemove, &KeyValueModel{
				Key: key + "[" + strconv.Itoa(i-1) + "]",
			})
		}

	}

	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	for _, keyValueModel := range needCreate {
		result := m.db.Create(keyValueModel)
		if result.Error != nil {
			return errors.Join(errors.New("create value error"), result.Error)
		}
	}

	for _, keyValueModel := range needSet {
		result := m.db.Where("Key = ?", keyValueModel.Key).Updates(keyValueModel)
		if result.Error != nil {
			return errors.Join(errors.New("set value error"), result.Error)
		}
	}

	for _, keyValueModel := range needRemove {
		result := m.db.Where("Key = ?", keyValueModel.Key).Delete(&KeyValueModel{})
		if result.Error != nil {
			return errors.Join(errors.New("remove value error"), result.Error)
		}
	}

	return nil
}

// 将数据转换为model，如果是slice类型，会返回多个model
func sqliteData2Model(key string, value *ValueUnit) ([]*KeyValueModel, error) {
	keyValueModel := &KeyValueModel{
		Key:       key,
		ValueType: int(value.Type),
	}
	switch value.Type {
	case VALUE_TYPE_INT:
		valueInt := ToBase[int](value)
		keyValueModel.ValueInt = &valueInt
	case VALUE_TYPE_BOOL:
		valueBool := ToBase[bool](value)
		if valueBool {
			valueInt := 1
			keyValueModel.ValueInt = &valueInt
		} else {
			valueInt := 0
			keyValueModel.ValueInt = &valueInt
		}
	case VALUE_TYPE_STRING:
		valueString := ToBase[string](value)
		keyValueModel.ValueString = &valueString
	case VALUE_TYPE_FLOAT:
		valueFloat := ToBase[float32](value)
		keyValueModel.ValueFloat = &valueFloat
	case VALUE_TYPE_SLICE_INT:
		sliceNum := len(value.Data.([]int))
		if sliceNum <= 0 {
			return nil, fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
		valueInt := sliceNum
		keyValueModel.ValueInt = &valueInt
	case VALUE_TYPE_SLICE_STRING:
		sliceNum := len(value.Data.([]string))
		if sliceNum <= 0 {
			return nil, fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
		valueInt := sliceNum
		keyValueModel.ValueInt = &valueInt
	case VALUE_TYPE_SLICE_FLOAT:
		sliceNum := len(value.Data.([]float32))
		if sliceNum <= 0 {
			return nil, fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
		valueInt := sliceNum
		keyValueModel.ValueInt = &valueInt
	case VALUE_TYPE_SLICE_BOOL:
		sliceNum := len(value.Data.([]bool))
		if sliceNum <= 0 {
			return nil, fmt.Errorf("slice but sliceNum is %d", sliceNum)
		}
		valueInt := sliceNum
		keyValueModel.ValueInt = &valueInt
	}

	result := make([]*KeyValueModel, 1)
	result[0] = keyValueModel

	switch value.Type {
	case VALUE_TYPE_SLICE_INT:
		//for i, v := range ToBase[[]int](value) {
		//	sliceErr = m.Set(Key+"["+strconv.Itoa(i)+"]", &ValueUnit{
		//		Type: VALUE_TYPE_INT,
		//		Data: v,
		//	})
		//}
		for i, v := range ToBase[[]int](value) {
			newV := v
			result = append(result, &KeyValueModel{
				Key:       key + "[" + strconv.Itoa(i) + "]",
				ValueType: int(VALUE_TYPE_INT),
				ValueInt:  &newV,
			})
		}
	case VALUE_TYPE_SLICE_STRING:
		for i, v := range ToBase[[]string](value) {
			newV := v
			result = append(result, &KeyValueModel{
				Key:         key + "[" + strconv.Itoa(i) + "]",
				ValueType:   int(VALUE_TYPE_STRING),
				ValueString: &newV,
			})
		}
	case VALUE_TYPE_SLICE_FLOAT:
		for i, v := range ToBase[[]float32](value) {
			newV := v
			result = append(result, &KeyValueModel{
				Key:        key + "[" + strconv.Itoa(i) + "]",
				ValueType:  int(VALUE_TYPE_FLOAT),
				ValueFloat: &newV,
			})
		}
	case VALUE_TYPE_SLICE_BOOL:
		for i, v := range ToBase[[]bool](value) {
			ti := 0
			if v {
				ti = 1
			}
			result = append(result, &KeyValueModel{
				Key:       key + "[" + strconv.Itoa(i) + "]",
				ValueType: int(VALUE_TYPE_BOOL),
				ValueInt:  &ti,
			})
		}
	}

	return result, nil
}

func (m *SqliteCore) Delete(key string) error {
	if !m.IsInitialized() {
		return errors.New("sqlite core not init")
	}
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	return m.db.Where("Key = ?", key).Delete(&KeyValueModel{}).Error
}

func (m *SqliteCore) Have(key string) (bool, error) {
	if !m.IsInitialized() {
		return false, errors.New("sqlite core not init")
	}
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	var keyValueModel KeyValueModel
	result := m.db.Where("Key = ?", key).First(&keyValueModel)
	if result.Error != nil {
		return false, errors.Join(errors.New("get value error"), result.Error)
	}
	return true, nil
}

func (m *SqliteCore) GetAll() (map[string]*ValueUnit, error) {
	if !m.IsInitialized() {
		return nil, errors.New("sqlite core not init")
	}
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	var keyValueModelList []KeyValueModel
	result := m.db.Find(&keyValueModelList)
	if result.Error != nil {
		return nil, errors.Join(errors.New("get all value error"), result.Error)
	}
	keyValueModelMap := make(map[string]*ValueUnit)
	for _, keyValueModel := range keyValueModelList {
		// 跳过所有含有[]的key，因为这些key是slice的成员，不是真正的key
		if strings.Contains(keyValueModel.Key, "[") || strings.Contains(keyValueModel.Key, "]") {
			continue
		}
		sliceNum, unit, err := sqliteModel2Data(keyValueModel)
		if err != nil {
			return nil, errors.Join(errors.New("sqliteModel2Data"), err)
		}

		if sliceNum != 0 {
			// slice 的内容存放在 Key[0]、Key[1]、Key[2]...Key[sliceNum-1] 列中
			_, unit, err = m.Get(keyValueModel.Key)
			if err != nil {
				return nil, errors.Join(errors.New("get slice value error"), err)
			}
		}
		keyValueModelMap[keyValueModel.Key] = unit
	}
	return keyValueModelMap, nil
}
