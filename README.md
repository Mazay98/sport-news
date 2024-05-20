# Sport News Parser

This task is based on a common business task we have within the backend team. We have a number of microservices that process data from external feed providers. This allows us to transform the data into a consistent and desirable format that the app developers can consume. It also provides stability so that, as is often the case, when the external provider has issues we can still provide data to the apps, albeit stale data.

The project must meet the following requirements:
Be written in Golang - https://golang.org/doc/ -
Use MongoDB to store news articles - (https://www.mongodb.com/download-center/community)
At regular intervals, poll the endpoint for new news articles
Transform the XML feeds of news articles into appropriate model(s) and save them in the database
Provide two REST endpoints which return JSON:
Retrieve a list of all articles
Retrieve a single article by ID
Comments where appropriate
Send over README file listing what works and what doesn't - as well as instructions (if any) on how to run

External Feeds
There are two feeds provided by our external provider for Huddersfield Town AFC club news:
Get a list of the latest n news articles
https://www.htafc.com/api/incrowd/getnewlistinformation?count=50
Get additional details of a news article by ID
https://www.htafc.com/api/incrowd/getnewsarticleinformation?id=XXXX

Example Output
Below are examples of output for the two REST endpoints. Ideally, your project should produce an output in a similar structure.
Retrieve a list of all articles
http://feeds.incrowdsports.com/provider/realise/v1/teams/t94/news
Retrieve a single article by ID
http://feeds.incrowdsports.com/provider/realise/v1/teams/t94/news/f48cf122-57f1-5512-bd80-b567e7f8c402

Additional Information
You may find the following links helpful.

Go also has a number of useful guides: https://golang.org/doc/
Useful Libraries:
https://github.com/gorilla/mux
https://github.com/mongodb/mongo-go-driver
https://github.com/jasonlvhit/gocron / https://godoc.org/github.com/robfig/cron

API Design Docs
https://pages.apigee.com/rs/apigee/images/api-design-ebook-2012-03.pdf
http://www.restapitutorial.com/resources.html
https://labs.omniti.com/labs/jsend



### For start 
```shell
docker-compose --profile dev up -d
```

### Endpoints 

**GET /v1/teams/{team}/news** - for get all news

**GET /v1/teams/{team}/news/{id}** - for get single news 

### Test

**For run e2e test you need up docker container and external server**
```shell
PARSER_ENABLE=0 MONGO_C_NAME=test docker-compose --profile dev up -d && \
go run e2e/cmd/external/server.go
```

**For run e2e tests you need run**
```shell
go test ./e2e
```

**For run unit tests you need run**
```shell
go test ./internal/repository
```
