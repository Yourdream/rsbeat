package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/liugaohua/rsbeat/config"
	"github.com/garyburd/redigo/redis"
	"strings"
)

type Rsbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
	lastIndexTime time.Time //test
	poolList map[string]*redis.Pool
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	logp.Info("config: %v", config.Redis )
	var poolList = make(map[string]*redis.Pool)
	var redisList []redisInstance
	var redisInstance redisInstance
	for _, ipPort := range  config.Redis {
		//poolList = append( poolList , poolInit(ipPort ,""))
		poolList[ipPort] = poolInit( ipPort,"")
		logp.Info("redis: %s", ipPort)

		var ipPortArr []string
		ipPortArr = strings.Split( ipPort , ":")
		redisInstance.ip = ipPortArr[0]
		redisInstance.port = ipPortArr[1]
		redisList = append(redisList, redisInstance)
	}
	fmt.Printf("%q\n", redisList )
	fmt.Printf("%q\n", poolList )

	bt := &Rsbeat{
		done: make(chan struct{}),
		config: config,
		poolList: poolList,
	}

	return bt, nil
}

type redisInstance struct {
	ip string
	port string
}
func (bt *Rsbeat) Run(b *beat.Beat) error {
	logp.Info("rsbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

/*		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       b.Name,
			"counter":    counter,
		}
		bt.client.PublishEvent(event)
		*/
		for ipPort, pool := range bt.poolList {
			logp.Info("Event sent instance:%s",ipPort )
			go bt.redisc( b.Name , true , pool.Get() , ipPort )
		}
		bt.lastIndexTime = time.Now()
		logp.Info("Event sent")
		counter++
	}
}

func (bt *Rsbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

type itemLog struct {
	slowId int
	timestamp int64
	duration int
	cmd string
	key string
	args []string
}

func (bt *Rsbeat ) redisc( beatname string, init bool, c redis.Conn , ipPort string ){
/*	c, err := redis.Dial("tcp", "192.168.33.10:6379")
	if err != nil {
		fmt.Println(err)
		return
	}*/
	defer c.Close()
	logp.Info("conn:%v", c )

	//c.Do("CONFIG", "SET", "slowlog-log-slower-than", "10")
	//reply , err := redis.Values(c.Do("slowlog", "get", 30 ))
	c.Send("SLOWLOG", "GET", "30")
	c.Send("SLOWLOG", "RESET")
	logp.Info("redis: slowlog get. slowlog rest");

	c.Flush()
	reply, err := redis.Values( c.Receive() ) // reply from GET
	c.Receive() // reply from SET


	now := time.Now()
	for _, i := range reply {
		rp, _ := redis.Values(i, err)
		var itemLog itemLog
		var args []string
		redis.Scan(rp,&itemLog.slowId,&itemLog.timestamp,&itemLog.duration,&args)
		argsLen := len( args )
		if argsLen >= 1{
			itemLog.cmd = args[0]
		}
		if argsLen >= 2{
			itemLog.key = args[1]
		}
		if argsLen >= 3{
			itemLog.args = args[2:]
		}
		t := time.Unix(itemLog.timestamp, 0).UTC()
		extraTime := t.Format("2006-01-02T15:04:05Z07:00")
		//t := time.Date( 0, 0, 0, 0, 0, itemLog.timestamp, 0, time.UTC)
		fmt.Println(itemLog.slowId,t,itemLog.duration,itemLog.cmd, itemLog.key, itemLog.args )
		event := common.MapStr{
			//"_id":       itemLog.slowId,
			"type":       beatname,
			"@timestamp": common.Time(t),
			"slowId": itemLog.slowId,
			"duration":   itemLog.duration,
			"cmd":   itemLog.cmd,
			"key":	itemLog.key,
			"args":  itemLog.args,
			"extraTime": extraTime,
			"ipPort":ipPort,
		}
		if init {
			// index all files and directories on init
			bt.client.PublishEvent(event)
		} else {
			// Index only changed files since last run.
			if now.After(bt.lastIndexTime) {
				bt.client.PublishEvent(event)
			}
		}
	}
}

func poolInit(server string , password string ) (*redis.Pool) {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func () (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			c.Do("CONFIG", "SET", "slowlog-log-slower-than", "10")
			logp.Info("redis: config set");
/*			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}*/
			//if _, err := c.Do("SELECT",1); err != nil {
			// c.Close()
			// return nil, err
			//}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			logp.Info("redis: PING");
			return err
		},
	}
}
