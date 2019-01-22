package match

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tesujiro/smf3/data/db"
)

type Matcher struct {
	cancelFunc func()
}

func NewMatcher(ctx context.Context) *Matcher {
	return &Matcher{}
}

func (m *Matcher) Run(ctx context.Context) error {
	var _cancelOnce sync.Once
	var _cancel func()
	ctx, _cancel = context.WithCancel(ctx)
	cancel := func() {
		_cancelOnce.Do(func() {
			fmt.Printf("Matcher cancel called.\n")
			_cancel()
		})
	}
	m.cancelFunc = cancel

	go m.loop(ctx, cancel)

	//<-ctx.Done()
	return nil
}

func (m *Matcher) loop(ctx context.Context, cancel func()) {
	defer cancel()
	tick := time.NewTicker(time.Millisecond * time.Duration(500)).C
	//signal_chan := make(chan os.Signal, 1)
	//signal.Notify(signal_chan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			err := m.match()
			if err != nil {
				log.Printf("match error:%v\n", err)
				log.Printf("Please Restart\n")
				cancel()
				break mainloop //??
			}
			//fmt.Printf("Matcher tick\n")
			//case s := <-signal_chan:
			//fmt.Printf("signal:%v\n", s)
			//break mainloop
		}
	}

}

func (m *Matcher) match() error {
	now := time.Now().Unix()
	flyers, err := db.ScanValidFlyers(now)
	if err != nil {
		return err
	}

	for _, f := range flyers {
		//fmt.Printf("flyer:%#v\n", f)
		// TODO: YURUFUA -> struct
		/*
			lat := f.(map[string]interface{})["geometry"].(map[string]interface{})["coordinates"].([]interface{})[1].(float64)
			lon := f.(map[string]interface{})["geometry"].(map[string]interface{})["coordinates"].([]interface{})[0].(float64)
			distance := f.(map[string]interface{})["properties"].(map[string]interface{})["distance"].(float64)
			flyerID := f.(map[string]interface{})["properties"].(map[string]interface{})["id"].(float64)          //TODO
			stocked := f.(map[string]interface{})["properties"].(map[string]interface{})["stocked"].(float64)     //TODO
			delivered := f.(map[string]interface{})["properties"].(map[string]interface{})["delivered"].(float64) //TODO
			//fmt.Printf("flyer:{lat:%v,lon:%v,distance:%v}\n", lat, lon, distance)
		*/
		lat := f.Lat
		lon := f.Lon
		distance := f.Distance
		flyerID := f.ID
		stocked := f.Stocked
		delivered := f.Delivered

		if stocked <= 0 {
			fmt.Printf("No flyer stock. flyerID:%v stocked:%v delivered:%v\n", flyerID, stocked, delivered)
			continue
		}

		locations, err := db.LocationWithinCircle(lat, lon, distance)
		if err != nil {
			return err
		}
		for _, loc := range locations {
			//fmt.Printf("location:%#v\n", loc)
			lat := loc.Geometry.Coordinates[1]
			lon := loc.Geometry.Coordinates[0]
			userID := loc.Properties["id"].(float64)
			//fmt.Printf("location:{userID:%v,lat:%v,lon:%v}\n", userID, lat, lon)
			now := time.Now().Unix()
			n := &db.Notification{
				//ID:           flyerID*100 + int64(userID), //TODO:
				ID:           db.NewNotificationID(),
				FlyerID:      int64(flyerID),
				UserID:       int64(userID),
				Lat:          lat,
				Lon:          lon,
				DeliveryTime: now,
			}

			if prev, err := db.GetNotification(fmt.Sprintf("%v", n.ID)); err != nil {
				return err
			} else if prev == nil && stocked > 0 {
				err := n.Set()
				if err != nil {
					return err
				}

				stocked--
				delivered++
				err = f.Jset("properties.stocked", stocked)
				if err != nil {
					return err
				}
				err = f.Jset("properties.delivered", delivered)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
