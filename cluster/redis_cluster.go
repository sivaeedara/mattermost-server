package cluster

import (
	"bytes"
	"fmt"
	"log"

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
}

func init() {
	log.Printf("***** Loading redis clustering *******")
	app.RegisterClusterInterface(func(a *app.App) einterfaces.ClusterInterface {
		rdCluster := &RedisCluster{}
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
			rdCluster.redisTopic = clusterName
		}
		return rdCluster
	})
}

func (me *RedisCluster) StartInterNodeCommunication() {
	if !*me.settings.Enable {
		mlog.Warn("redis cluster needs redis enabled as well")
		return
	}
	mlog.Info("***** starting redis_clustering. Subscribing ... *****")
	me.pubsub = me.client.Subscribe(me.redisTopic)
	for true {
		msg, error := me.pubsub.ReceiveMessage()
		if error == nil {
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
			mlog.Error(error.Error())
		}
	}
}

func (me *RedisCluster) StopInterNodeCommunication() {
	if me.pubsub != nil {
		error := me.pubsub.Unsubscribe(me.redisTopic)
		if error != nil {
			mlog.Error(error.Error())
		}
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
