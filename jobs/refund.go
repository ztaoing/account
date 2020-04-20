package jobs

import (
	"fmt"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tietang/go-utils"
	"github.com/ztaoing/infra"
	"time"
)

//退款
type RefundExpiredStarter struct {
	infra.BaseStarter
	ticker *time.Ticker
	mutex  *redsync.Mutex
}

func (r *RefundExpiredStarter) Init(ctx infra.StarterContext) {
	//时间间隔
	d := ctx.Props().GetDurationDefault("jobs.refund.interval", time.Minute)
	r.ticker = time.NewTicker(d)
	//构建redis连接池
	pools := make([]redsync.Pool, 0)
	maxIdle := ctx.Props().GetIntDefault("redis.maxIdle", 2)
	maxActive := ctx.Props().GetIntDefault("redis.maxActive", 5)
	timeout := ctx.Props().GetDurationDefault("redis.timeout", 20*time.Second)
	addr := ctx.Props().GetDefault("redis.addr", "127.0.0.1:6379")
	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive, //连接池的最大活跃链接数
		IdleTimeout: timeout,
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp", addr)

		},
	}
	pools = append(pools, pool)
	//根据提供的redis连接池，创建一个redsync实例
	rsysnc := redsync.New(pools)
	//获得ip地址
	ip, err := utils.GetExternalIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	//通过实例创建互斥锁

	r.mutex = rsysnc.NewMutex("lock:RefundExpired",
		redsync.SetExpiry(50*time.Second),
		redsync.SetRetryDelay(3),
		redsync.SetGenValueFunc(func() (string, error) {
			now := time.Now()
			logrus.Infof("节点%s正在执行过期红包的退款任务", ip)
			return fmt.Sprintf("%d:%s", now.Unix(), ip), nil
		}),
	)

}

func (r *RefundExpiredStarter) Start(ctx infra.StarterContext) {
	//使用goroutine来执行定时
	go func() {
		for {
			c := <-r.ticker.C
			//释放锁
			defer r.mutex.Unlock()
			//尝试取锁
			if err := r.mutex.Lock(); err == nil {
				//拿到锁
				log.Debug("触发红包退款...", c)
				//红包退款
				domain := envelopes.ExpiredEnvelopeDomain{}
				domain.Expired()
			} else {
				//没有拿到锁
				log.Debug("获取锁失败")
			}

		}
	}()
}

//停止的时候，定时任务也需要停止
func (r *RefundExpiredStarter) Stop(ctx infra.StarterContext) {
	//停止定时触发红包退款
	r.ticker.Stop()
}
