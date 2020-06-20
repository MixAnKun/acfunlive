// 弹幕下载相关
package main

import (
	"context"
	"time"

	"github.com/orzogc/acfundanmu"
)

var subConfigs = map[int]acfundanmu.SubConfig{
	0:    {PlayResX: 720, PlayResY: 1280, FontSize: 20},
	540:  {PlayResX: 960, PlayResY: 540, FontSize: 30},
	720:  {PlayResX: 1280, PlayResY: 720, FontSize: 40},
	1080: {PlayResX: 1920, PlayResY: 1080, FontSize: 60},
}

// 设置自动下载指定主播的直播弹幕
func addDanmu(uid int) bool {
	isExist := false
	streamers.mu.Lock()
	if s, ok := streamers.crt[uid]; ok {
		isExist = true
		if s.Danmu {
			lPrintln("已经设置过自动下载" + s.Name + "的直播弹幕")
		} else {
			s.Danmu = true
			sets(s)
			lPrintln("成功设置自动下载" + s.Name + "的直播弹幕")
		}
	}
	streamers.mu.Unlock()

	if !isExist {
		name := getName(uid)
		if name == "" {
			lPrintln("不存在uid为" + itoa(uid) + "的用户")
			return false
		}

		newStreamer := streamer{UID: uid, Name: name, Notify: false, Record: false, Danmu: true}
		streamers.mu.Lock()
		sets(newStreamer)
		streamers.mu.Unlock()
		lPrintln("成功设置自动下载" + name + "的直播弹幕")
	}

	saveConfig()
	return true
}

// 取消自动下载指定主播的直播弹幕
func delDanmu(uid int) bool {
	streamers.mu.Lock()
	if s, ok := streamers.crt[uid]; ok {
		if s.Notify || s.Record {
			s.Danmu = false
			sets(s)
		} else {
			deleteStreamer(uid)
		}
		lPrintln("成功取消订阅" + s.Name + "的开播提醒")
	} else {
		lPrintln("没有订阅过uid为" + itoa(uid) + "的主播的开播提醒")
	}
	streamers.mu.Unlock()

	saveConfig()
	return true
}

// 下载直播弹幕
func (s streamer) getDanmu(ctx context.Context, cfg acfundanmu.SubConfig, filename string) {
	dctx, dcancel := context.WithCancel(ctx)
	defer dcancel()
	msgMap.mu.Lock()
	if m, ok := msgMap.msg[s.UID]; ok {
		m.danmuCancel = dcancel
		msgMap.msg[s.UID] = m
	} else {
		msgMap.msg[s.UID] = sMsg{danmuCancel: dcancel}
	}
	msgMap.mu.Unlock()

	assFile, ok := transFilename(filename)
	if !ok {
		return
	}
	assFile = assFile + ".ass"

	lPrintln("开始下载" + s.longID() + "的直播弹幕")
	lPrintln("本次下载的ass文件保存在" + assFile)
	if *isListen {
		lPrintln("如果想提前结束下载直播弹幕，运行stopdanmu " + s.itoa())
	}
	desktopNotify("开始下载" + s.Name + "的直播弹幕")
	q := acfundanmu.Start(dctx, s.UID)
	cfg.Title = filename
	cfg.StartTime = time.Now().UnixNano()
	q.WriteASS(dctx, cfg, assFile)

	msgMap.mu.Lock()
	m := msgMap.msg[s.UID]
	m.danmuCancel = nil
	msgMap.msg[s.UID] = m
	msgMap.mu.Unlock()
	deleteMsg(s.UID)

	lPrintln(s.longID() + "的直播弹幕下载已经结束")
	desktopNotify(s.Name + "的直播弹幕下载已经结束")
}

// 临时下载指定主播的直播弹幕
func startDanmu(uid int) bool {
	name := getName(uid)
	if name == "" {
		lPrintln("不存在uid为" + itoa(uid) + "的用户")
		return false
	}
	s := streamer{UID: uid, Name: name}

	if !s.isLiveOn() {
		lPrintln(s.longID() + "不在直播，取消下载弹幕")
		return false
	}

	_, _, cfg := s.getStreamURL()
	filename := getTime() + " " + s.Name + " " + s.getTitle()

	// 查看程序是否处于监听状态
	if *isListen {
		// goroutine是为了快速返回
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			s.getDanmu(ctx, cfg, filename)
		}()
	} else {
		// 程序只在单独下载一个直播，不用goroutine，防止程序提前结束运行
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		s.getDanmu(ctx, cfg, filename)
	}
	return true
}

// 停止下载指定主播的直播弹幕
func stopDanmu(uid int) bool {
	msgMap.mu.Lock()
	defer msgMap.mu.Unlock()
	if m, ok := msgMap.msg[uid]; ok {
		s := streamer{UID: uid, Name: getName(uid)}
		if m.danmuCancel != nil {
			lPrintln("开始停止下载" + s.longID() + "的直播弹幕")
			m.danmuCancel()
		} else {
			lPrintln("没有在下载" + s.longID() + "的直播弹幕")
		}
	} else {
		lPrintln("没有在下载uid为" + itoa(uid) + "的主播的直播弹幕")
	}
	return true
}
