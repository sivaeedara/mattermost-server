# Clustering Mattermost using Redis

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
- Run ``go build cmd/mattermost/main.go``, or for linux-built environment, run ``GOOS=linux go build cmd/mattermost/main.go``

### [Download Build](https://drive.google.com/open?id=14Mnveq-JcHDDnEltgcJCMCJA-OscAmPx)

## [![About Mattermost](https://user-images.githubusercontent.com/33878967/33095422-7c8aa7a4-ceb8-11e7-810a-4b261fdff6d6.png)](https://mattermost.org)

