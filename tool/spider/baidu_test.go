package spider

import (
	"testing"
)

func TestGetBaiduNews(t *testing.T) {
	params := []string{
		"gta6",
		"iphone16 ",
		"ps5pro ",
		"apple glass",
		"iphone16",
		"特斯拉",
	}
	keywords := []string{}
	newss := [][]BaiduNew{}
	for _, v := range params {
		news, err, _ := GetTodayBaiduNews(v)
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("%s: %d", v, len(news))
		keywords = append(keywords, v)
		newss = append(newss, news)
	}
	//for _, news := range newss {
	//	for _, baiduNew := range news {
	//		t.Log(baiduNew.title)
	//		t.Log(baiduNew.content)
	//		t.Log(baiduNew.source)
	//		t.Log(baiduNew.time)
	//		t.Log(baiduNew.same)
	//		//println(baiduNew.title)
	//		//println(baiduNew.content)
	//		//println(baiduNew.source)
	//		//println(baiduNew.time)
	//		//println(baiduNew.same)
	//	}
	//}
	_ = ParseNewToMarkdown(keywords, newss)
	//f, _ := os.Create("baidu.md")
	//f.WriteString(s)
	//p := xpush.Mgr{}
	//p.SetTag("auto")
	//p.SetPushDeerToken("PDU10120Tp8PByEPFdrKiStSvMWeOdeFtwY7GuOmQ")
	//p.PushPushDeer("新闻", s, true)
}

func TestGetBaiduNewsNew(t *testing.T) {
	param := "iphone16"
	results, newLink, err, _ := GetBaiduNewsNew(param, "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("no news")
	}
	if results[0].href != newLink {
		t.Fatal("newLink error")
	}
	if len(results) < 10 {
		t.Fatal("results too less")
	}
	results2, _, err2, _ := GetBaiduNewsNew(param, results[5].href, 1)
	if err2 != nil {
		t.Fatal(err2)
	}
	if len(results2) != 5 {
		t.Fatal("results2 error")
	}
	for i := 0; i < 5; i++ {
		if results2[i].title != results[i].title {
			t.Fatal("results2 error")
		}
	}
}
