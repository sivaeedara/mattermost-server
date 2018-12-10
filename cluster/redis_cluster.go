package cluster

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/go-redis/redis"
	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/einterfaces"
	"github.com/mattermost/mattermost-server/model"
)

//RedisCluster redis cluster implementation
type RedisCluster struct {
	app *app.App
	einterfaces.ClusterInterface
	clusterMessageHandler map[string]einterfaces.ClusterMessageHandler
	client                *redis.Client
	pubsub                *redis.PubSub
	redisInstanceId       string
	redisTopic            string
	settings              *model.RedisSettings
	connectChan           chan int
}

func (me *RedisCluster) connectAndFetch() {
	go func() {
		defer close(me.connectChan)
		for {
			_, ok := <-me.connectChan
			if ok {
				mlog.Info("****** redis_cluster receive connect signal. Start connecting *******")
				me.subscribe()
			} else {
				mlog.Info("Exit connect. Clear your stuff now")
			}
		}
	}()
}

func init() {
	log.Printf("***** Loading redis clustering *******")
	app.RegisterClusterInterface(func(a *app.App) einterfaces.ClusterInterface {
		rdCluster := &RedisCluster{}
		rdCluster.connectChan = make(chan int, 1)
		rdCluster.app = a
		rdCluster.settings = &a.GetConfig().RedisSettings
		rdCluster.clusterMessageHandler = make(map[string]einterfaces.ClusterMessageHandler)
		if *rdCluster.settings.Enable {
			rdCluster.redisInstanceId = model.NewId()

			option := &redis.Options{Addr: *rdCluster.settings.Address, PoolSize: *rdCluster.settings.PoolSize,
				DB: *rdCluster.settings.Index, MaxRetries: 3}
			rdCluster.client = redis.NewClient(option)
			clusterName := *a.GetConfig().ClusterSettings.ClusterName

			if clusterName == "" {
				clusterName = "vincere-chat"
			}
			mlog.Info(fmt.Sprintf("******* redis_cluster cluster topic is %v ******\n", clusterName))
			rdCluster.redisTopic = clusterName
		}
		rdCluster.connectAndFetch()
		return rdCluster
	})
}

func (me *RedisCluster) StartInterNodeCommunication() {
	if !*me.settings.Enable {
		mlog.Warn("redis cluster needs redis enabled as well")
		return
	}
	//start connect
	me.connectChan <- 1
}

func (me *RedisCluster) processRedisMessage(msg *redis.Message) {
	if msg != nil {
		clusterMsg := model.ClusterMessageFromJson(bytes.NewReader([]byte(msg.Payload)))
		if me.redisInstanceId != clusterMsg.Props["instance-id"] {
			mlog.Debug(fmt.Sprintf("Received redis_event from cluster %v\n", clusterMsg))
			me.app.Go(func() {
				crm := me.clusterMessageHandler[clusterMsg.Event]
				if crm != nil {
					crm(clusterMsg)
				}
			})
		} else {
			mlog.Debug(fmt.Sprintf("Instance %v received its published message. Skip it \n",
				me.redisInstanceId))
		}
	} else {
		mlog.Warn("redis_cluster received NULL from channel")
	}
}

func (me *RedisCluster) subscribe() {
	mlog.Info("***** starting redis_clustering. Subscribing ... *****")
	me.pubsub = me.client.Subscribe(me.redisTopic)
	go func() {
		//breakLoop:
		errorCount := 0
		for {
			msg, err := me.pubsub.ReceiveMessage()
			if err != nil {
				mlog.Error(err.Error())
				time.Sleep(5 * time.Second)
				errorCount++
				if errorCount > 5 {
					mlog.Warn(fmt.Sprintf(
						"******** redis_cluster needs reconnect. Number of errors exists %v times \n", errorCount))
					me.connectChan <- 1
					break
				}
				// 	me.connectChan <- 1
			} else {
				me.processRedisMessage(msg)
			}
		}
		mlog.Info("Exiting subscribe ...")
	}()
}

func (me *RedisCluster) StopInterNodeCommunication() {
	if me.pubsub != nil {
		error := me.pubsub.Unsubscribe(me.redisTopic)
		if error != nil {
			mlog.Error(error.Error())
		}
		me.pubsub.Close()
	}
}

func (me *RedisCluster) RegisterClusterMessageHandler(event string, crm einterfaces.ClusterMessageHandler) {
	me.clusterMessageHandler[event] = crm
}
func (me *RedisCluster) GetClusterId() string {
	return ""
}
func (me *RedisCluster) IsLeader() bool {
	return false
}
func (me *RedisCluster) GetMyClusterInfo() *model.ClusterInfo {
	return nil
}
func (me *RedisCluster) GetClusterInfos() []*model.ClusterInfo {
	return nil
}
func (me *RedisCluster) SendClusterMessage(cluster *model.ClusterMessage) {
	if cluster == nil {
		mlog.Warn("Clustering Message is empty. Skipping ...")
		return
	}
	if me.client == nil {
		mlog.Warn("redis_cluster client has not set. Skipping ...")
		return
	}
	if cluster.Props == nil {
		cluster.Props = make(map[string]string)
	}
	cluster.Props["instance-id"] = me.redisInstanceId
	me.client.Publish(me.redisTopic, []byte(cluster.ToJson()))
}
func (me *RedisCluster) NotifyMsg(buf []byte) {
	log.Println("redis cluster doesnt support notifyMsg")
}

func (me *RedisCluster) GetClusterStats() ([]*model.ClusterStats, *model.AppError) {
	return nil, nil
}
func (me *RedisCluster) GetLogs(page, perPage int) ([]string, *model.AppError) {
	return nil, nil
}
func (me *RedisCluster) GetPluginStatuses() (model.PluginStatuses, *model.AppError) {
	return nil, nil
}
func (me *RedisCluster) ConfigChanged(previousConfig *model.Config, newConfig *model.Config, sendToOtherServer bool) *model.AppError {
	return nil
}
