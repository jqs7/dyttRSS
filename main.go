package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/feeds"
	"github.com/levigross/grequests"
	"github.com/tuotoo/biu"
	"github.com/tuotoo/biu/box"
	"github.com/tuotoo/biu/log"
	"github.com/tuotoo/biu/opt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type DYTT struct{}

func (ctl DYTT) WebService(ws biu.WS) {
	ws.Route(ws.GET("rss.xml"),
		opt.RouteID("dytt.rss"),
		opt.RouteTo(ctl.rss),
		opt.RouteErrors(map[int]string{
			100: "请求电影页面失败",
			101: "获取页面结构失败",
			102: "RSS 输出失败",
		}),
	)
}

func (ctl DYTT) rss(ctx box.Ctx) {
	host := "https://www.dy2018.com"
	resp, err := grequests.Get(host+"/html/gndy/dyzz/index.html", nil)
	ctx.Must(err, 100)
	defer func() {
		err := resp.Close()
		if err != nil {
			log.Info().Err(err).Msg("close resp")
		}
	}()

	gbkResp := transform.NewReader(resp, simplifiedchinese.GBK.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(gbkResp)
	ctx.Must(err, 101)

	tables := doc.Find(".co_content8").First().Find("ul").Children()

	feed := &feeds.Feed{
		Title:  "电影天堂",
		Link:   &feeds.Link{Href: host},
		Author: &feeds.Author{Name: "Jqs7", Email: "7@jqs7.com"},
	}
	items := make([]*feeds.Item, tables.Size())
	tables.Each(func(i int, table *goquery.Selection) {
		trs := table.Find("tr")
		a := trs.Eq(1).Find("a").First()
		td := trs.Eq(3).Find("td").First()
		items[i] = &feeds.Item{
			Title:       a.Text(),
			Link:        &feeds.Link{Href: host + a.AttrOr("href", "/")},
			Description: td.Text(),
		}
	})
	feed.Items = items
	rss, err := feed.ToRss()

	ctx.Must(err, 102)
	ctx.Write([]byte(rss))
}

func main() {
	log.UseColorLogger()
	restful.Filter(biu.LogFilter())
	biu.AddServices("/v1", nil,
		biu.NS{
			NameSpace:  "dytt",
			Controller: DYTT{},
			Desc:       "电影天堂",
		},
	)
	swaggerService := biu.NewSwaggerService(biu.SwaggerInfo{
		Version:     "1.0.0",
		RoutePrefix: "/v1",
	})
	restful.Add(swaggerService)
	biu.Run(":7096")
}
