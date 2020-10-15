package cluster

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mattermost/mattermost-server/v5/mlog"

	"github.com/go-redis/redis"
	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/model"
)

var logClusterUnsupport = []string{`we don't support log live monitoring for redis cluster. Please use external log services like logentries or ELK instead`}

//RedisCluster redis cluster implementation
type RedisCluster struct {
	server *app.Server
	einterfaces.ClusterInterface
	clusterMessageHandler map[string]einterfaces.ClusterMessageHandler
	client                *redis.Client
	pubsub                *redis.PubSub
	redisInstanceId       string
	redisTopic            string
	settings              *model.RedisSettings
	subscribeChan         chan int
	clusterInfo           *model.ClusterInfo
}

func init() {
	log.Printf("***** Loading redis clustering *******")
	app.RegisterClusterInterface(func(s *app.Server) einterfaces.ClusterInterface {
		settings := &s.Config().RedisSettings
		if !*settings.EnableRedisCluster || !*settings.Enable {
			mlog.Warn("To use redis cluster, please enable redis and redis cluster")
			return nil
		}
		rdCluster := &RedisCluster{}
		rdCluster.subscribeChan = make(chan int, 1)
		rdCluster.server = s
		rdCluster.settings = settings
		rdCluster.clusterMessageHandler = make(map[string]einterfaces.ClusterMessageHandler)
		if *rdCluster.settings.Enable {
			rdCluster.redisInstanceId = model.NewId()

			option := &redis.Options{Addr: *rdCluster.settings.Address,
				PoolSize: *rdCluster.settings.PoolSize,
				DB:       *rdCluster.settings.Index, MaxRetries: 3}
			if settings.Password != nil && *settings.Password != "" {
				option.Password = *settings.Password
			}
			rdCluster.client = redis.NewClient(option)
			clusterName := *s.Config().ClusterSettings.ClusterName

			if clusterName == "" {
				clusterName = "vincere-chat"
			}
			mlog.Info(fmt.Sprintf("******* redis_cluster cluster topic is %v ******", clusterName))
			rdCluster.redisTopic = clusterName
		}
		rdCluster.clusterInfo = &model.ClusterInfo{
			Id: rdCluster.redisInstanceId,
		}
		rdCluster.startConnectListener()
		return rdCluster
	})
}

func (me *RedisCluster) StartInterNodeCommunication() {
	if !*me.settings.Enable {
		mlog.Warn("redis cluster needs redis enabled as well")
		return
	}
	//start connect
	me.subscribeChan <- 1
}

func (me *RedisCluster) startConnectListener() {
	go func() {
		//defer close(me.subscribeChan)
		for {
			_, ok := <-me.subscribeChan
			if ok {
				mlog.Info("****** redis_cluster receive connect signal. Start connecting *******")
				me.subscribe()
			} else {
				mlog.Info("Exit connect. Clear your stuff now")
				break
			}
		}
		mlog.Info("Stopped Redis Cluster Listerner")
	}()
}

func (me *RedisCluster) processRedisMessage(msg *redis.Message) {
	if msg != nil {
		clusterMsg := model.ClusterMessageFromJson(bytes.NewReader([]byte(msg.Payload)))
		if me.redisInstanceId != clusterMsg.Props["instance-id"] {
			mlog.Debug(fmt.Sprintf("Received redis_event from cluster %v", clusterMsg))
			go func() {
				crm := me.clusterMessageHandler[clusterMsg.Event]
				if crm != nil {
					crm(clusterMsg)
				}
			}()
		} else {
			mlog.Debug(fmt.Sprintf("Instance %v received its published message. Skip it",
				me.redisInstanceId))
		}
	} else {
		mlog.Warn("redis_cluster received NULL from channel")
	}
}

func (me *RedisCluster) subscribe() {
	mlog.Info("***** starting redis_clustering. Subscribing ... *****")
	me.pubsub = me.client.Subscribe(me.redisTopic)
	func() {
		for {
			msg, ok := <-me.pubsub.Channel()
			if ok {
				me.processRedisMessage(msg)
			} else {
				mlog.Info("redis_clustering exiting subscribe(). PubSub Channel was already closed")
				break
			}
		}
	}()
}

func (me *RedisCluster) StopInterNodeCommunication() {
	defer close(me.subscribeChan)
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
	return me.clusterInfo
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
	return []*model.ClusterStats{}, nil
}
func (me *RedisCluster) GetLogs(page, perPage int) ([]string, *model.AppError) {
	return logClusterUnsupport, nil
}

func (me *RedisCluster) GetPluginStatuses() (model.PluginStatuses, *model.AppError) {
	return nil, nil
}
func (me *RedisCluster) ConfigChanged(previousConfig *model.Config, newConfig *model.Config, sendToOtherServer bool) *model.AppError {
	mlog.Info("Config changed")
	return nil
}
