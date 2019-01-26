# Clustering Mattermost (version 5.6.0) using Redis

## For those who are using Mattermost Team Edition finding a solution of clustering Mattermost. This project would help you.

This project is to make Mattermost running in cluster mode using Redis Pubsub. 

How to make it run: 

1. Update configuration 
- In config.json, enable clustering mode ClusterSettings > Enable
- Adding redis settings. 
  ```json 
   "RedisSettings": {
        "Enable": true,
        "EnableRedisCluster": true,
        "Address": "localhost:6379",
        "PoolSize": 100, 
        "Index": 2
    } 
    ```
2. Build mattermost-webapp for redis clustering (This version added Redis settings under advanced tab/ console admin)
- Checkout mattermost-webapp for redis clustering version here:https://github.com/mrbkiter/mattermost-webapp.git
- Place the project under same parent folder of mattermost-server (for example, you placed mattermost-server under src/ folder, now you should place mattermost-webapp under same one)
- Run ``make build`` to build webapp project. You need to install node. 

3. Run mattermost-server (this step require you have basic knowledge of Golang and how to setup Golang environment)
- Go to mattermost-server
- Run ``go run cmd/mattermost/main.go`` if you like to use go command.
- Or run ``make run-server`` to use make cmd. Please note that this command would automatically install docker dependent images such as DB, .... 
- Or run ``make run-server-no-docker`` if you like to run without docker dependencies. For this case, you would need to create an empty database, and update JDBC URL in config/config.json

4. If you need to build mattermost-server.
- Go to mattermost-server
- Run ``go build -o mattermost cmd/mattermost/main.go``, or for linux-built environment, run ``GOOS=linux go build -o mattermost cmd/mattermost/main.go``, or for Windows run ``GOOS=windows GOARCH=amd64 go build -o mattermost cmd/mattermost/main.go``

## Updates from original 5.6.0

- Open public for preview file APIs. (now you don't need session to preview images)
- Add API that get file infos for one channel. 

## Mattermost additional APIs
- GET /api/v4/files/channel/<channel_id>?page=0&per_page=10

Return file infos related to a channel. 

Response: 

```
[
    {
        "id": "jk15apoxuir5zk884f97jw79my",
        "user_id": "mcd6rjz8j7fcmm7f6u6nm669qc",
        "post_id": "c74rczih1bdgurgsh3emzsgfhe",
        "channel_id": "pcr3ackxh7bk7xn8x1y61x6bwy",
        "create_at": 1548475503393,
        "update_at": 1548475503393,
        "delete_at": 0,
        "name": "handset_round-2-512.png",
        "extension": "png",
        "size": 28339,
        "mime_type": "image/png",
        "width": 512,
        "height": 512,
        "has_preview_image": true
    }
]
```

### [Download Build](https://drive.google.com/open?id=14Mnveq-JcHDDnEltgcJCMCJA-OscAmPx)

## [About Mattermost on Github](https://github.com/mattermost/mattermost-server)

