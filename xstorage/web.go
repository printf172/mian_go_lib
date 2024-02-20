package xstorage

import (
	"errors"
	"fmt"
	"github.com/intmian/mian_go_lib/tool/misc"
	"github.com/intmian/mian_go_lib/xlog"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebCode int // web业务码

// 可能会被外部调用，所以复杂命名
const (
	WebCodeNull WebCode = iota
	WebCodeSuc
	WebCodeFail
)

// WebPack 是xstorage的拓展之一，必须绑定xstorage使用
// 因为是小众需求所以做一下拆分
type WebPack struct {
	storageCore *XStorage
	ginEngine   *gin.Engine
	log         *xlog.XLog
	logFrom     string
	setting     WebPackSetting
	misc.InitTag
}

type WebPackSetting struct {
	LogFrom string
	Log     *xlog.XLog
	WebPort int
}

func (w *WebPack) Init(setting WebPackSetting, core *XStorage) error {
	w.setting = setting
	w.storageCore = core
	w.SetInitialized()
	return nil
}

func NewWebPack(setting WebPackSetting, core *XStorage) (*WebPack, error) {
	m := &WebPack{}
	err := m.Init(setting, core)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type WebFailReason int // web失败原因

const (
	WebFailReasonNull WebFailReason = iota
	WebFailReasonNoLegalParam
	WebFailReasonInnerError
)

func (w *WebPack) StartWeb() error {

	w.ginEngine = gin.Default()
	w.ginEngine.GET("/get", w.WebGet)
	w.ginEngine.GET("/set", w.WebSet)
	w.ginEngine.GET("/get_all", w.WebGetAll)
	addr := fmt.Sprintf("127.0.0.1:%d", w.setting.WebPort)
	err := w.ginEngine.Run(addr)
	if err != nil {
		return errors.Join(ErrGinEngineRun, err)
	}

	return nil
}
func (w *WebPack) WebGet(c *gin.Context) {
	if !w.IsInitialized() {
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonInnerError,
		})
		return
	}
	// 读取正则表达式
	//useRe := c.Query("useRe")
	//perm := c.Query("perm")
	// 从body中读取
	var body struct {
		UseRe bool   `json:"useRe"`
		Perm  string `json:"perm"`
	}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonNoLegalParam,
		})
		return
	}
	useRe := body.UseRe
	perm := body.Perm

	var results []ValueUnit
	if !useRe {
		result := &ValueUnit{}
		ok, err := w.storageCore.GetHP(perm, result)
		if err != nil {
			w.log.Error(w.setting.LogFrom, "xStorage:WebGet:get value error:"+err.Error())
			c.JSON(200, gin.H{
				"code": WebCodeFail,
				"msg":  WebFailReasonNoLegalParam,
			})
			return
		}
		if !ok {
			c.JSON(200, gin.H{
				"code": WebCodeFail,
				"msg":  WebFailReasonNoLegalParam,
			})
			return
		}
		results = append(results, *result)
	} else {
		// 遍历并且搜索正则
		all, err := w.storageCore.GetAll()
		if err != nil {
			w.log.Error(w.setting.LogFrom, "xStorage:WebGet:get all value error:"+err.Error())
			c.JSON(200, gin.H{
				"code": WebCodeFail,
				"msg":  WebFailReasonInnerError,
			})
			return
		}
		found := false
		for k, _ := range all {
			// 使用正则
			matched, err := regexp.MatchString(perm, k)
			if err != nil {
				w.log.Error(w.setting.LogFrom, "xStorage:WebGet:match string error:"+err.Error())
				c.JSON(200, gin.H{
					"code": WebCodeFail,
					"msg":  WebFailReasonInnerError,
				})
				return
			}
			if matched {
				found = true
				results = append(results, *all[k])
			}
		}
		if !found {
			c.JSON(200, gin.H{
				"code": WebCodeFail,
				"msg":  WebFailReasonNoLegalParam,
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"code":   WebCodeSuc,
		"result": results,
	})
}

func (w *WebPack) WebSet(c *gin.Context) {
	if !w.IsInitialized() {
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonInnerError,
		})
		return
	}
	key := c.Query("key")
	valueType := c.Query("value_type")
	valueTypeInt, err := strconv.Atoi(valueType)
	if err != nil {
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonNoLegalParam,
		})
		return
	}
	value := c.Query("value")

	err = w.storageCore.Set(key, StringToUnit(value, ValueType(valueTypeInt)))
	if err != nil {
		w.log.Error(w.setting.LogFrom, "xStorage:WebSet:set value error:"+err.Error())
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonInnerError,
		})
		return
	}
	c.JSON(200, gin.H{
		"code": WebCodeSuc,
	})
}

func (w *WebPack) WebGetAll(c *gin.Context) {
	if !w.IsInitialized() {
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonInnerError,
		})
		return
	}
	all, err := w.storageCore.GetAll()
	if err != nil {
		w.log.Error(w.setting.LogFrom, "xStorage:WebGet:get all value error:"+err.Error())
		c.JSON(200, gin.H{
			"code": WebCodeFail,
			"msg":  WebFailReasonInnerError,
		})
		return
	}
	c.JSON(200, gin.H{
		"code":   WebCodeSuc,
		"result": all,
	})
}
