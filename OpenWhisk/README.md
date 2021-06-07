# IBM MQ Samples for OpenWhisk and IBM Cloud Functions
These MQ Samples for OpenWhisk consist of 2 actions, 1 feed, 1 sequence, 2 triggers and 2 trigger-action rules.

## Pattern
The supplied feed provides a simple *detect single message, inform all subscribers pattern*. More patterns can be integrated into the feed.

### To do
* *detect all messages, fire single trigger with array of message Ids*
* *detect all messages, fire triggers in round robin with message Ids in groups*


## Actions
The two actions are `mqfunctions/mqpost` and `mqfunctions/mqdelete`.

* **mqfunctions/mqpost** - Writes a message to a queue. If the input json contains `args.message`, then that message is written. Else a default timed message is posted.
* **mqfunctions/mqdelete** - Reads a specific message from a queue. Input json must contain `args.messageId`. It is this messageId that is used to fetch a message. Output json will contain the retrieved message, in the format:

```json
{
  'messageId' : ...,
  'message' : ...
}
```

## Feed
The feed is `mqfunctions/mqfeed`. When the feed is invoked, it will check for registered triggers. If the are registered triggers then the feed browses for the next message on the queue. If there is a message it fires the triggers sending the messageId of the next message on the queue as a parameter.


## Sequence
A single sample sequence `mq-default-sequence` is provided, which in turn invokes the `mqfunctions/mqdelete` action followed by the `mqfunctions/mqpost` action. This sequence is designed to take a message off the queue and if sucessful add another message. Thereby leaving a message on the queue for the next trigger round.


## Triggers
The two triggers are:

* **mq-feedtimer-trigger** - Which fires up every two minutes.
* **mq-feedtest-trigger** - Which registers to the feed mqfunctions/mqfeed as a subscribed trigger.


## Rules
The two rules are:

* **rule-mq-fire-trigger** - Which connects the trigger `mq-feedtimer-trigger` to the feed `mqfunctions/mqfeed`. This invokes the feed every 2 minutes. Forcing the feed to check for registered triggers and firing them. This package registers the `mq-feedtest-trigger`, so that trigger will be fired if there is a message available on the message queue.
* **rule-mq-default-sequence** - Which connects the `mq-feedtest-trigger` to the sequence
`mq-default-sequence`. It will only be fired if a message is available on the queue, and the sequence will retrieve that message and post a fresh message on the queue.


## Configurations
To interact with the Queue, the feed and actions need the following MQ configuration details:

* **QM_NAME** - Queue manager name
* **QUEUE** - Queue name
* **HOSTNAME** - Host name or IP address of your queue manager
* **PORT** - *HTTP* Listener port for your queue manager
* **MQ_USER** - User name that application uses to connect to MQ
* **MQ_PASSWORD** - Password that the application uses to connect to MQ

To persist registered triggers the feed needs the following Cloudant configuration details:

* **MQ_DB** - The name of the database created in Cloudant for the feed to use
* **MQ_DB_KEY** - The key that the feed will use to persist registred triggers
* **CLOUDANT_HOSTNAME** - Cloudant instance hostname
* **CLOUDANT_USERNAME** - Username used to connect to Cloudant
* **CLOUDANT_PASSWORD** - Password used to connect to Cloudant.


## Pre-requsites
To deploy and run these samples you will need:

* **MQ server** - You can use [IBM MQ in IBM Cloud](https://cloud.ibm.com/catalog/services/mq)
* **OpenWhisk** - You can use [IBM Cloud Functions](https://cloud.ibm.com/functions/)
* **OpenWhisk CLI** - You can use [IBM Cloud CLI plugin](https://cloud.ibm.com/functions/learn/cli)
* **Cloudant** - You can use [Cloudant on IBM Cloud](https://cloud.ibm.com/catalog/services/cloudant)


## Deploying the samples
Deploy from the `mq-package` directory.
If deploying to IBM Cloud, use the command line to `login` to IBM Cloud. Edit the `deploy.sh` script to provide configuration details, and run

````
./deploy.sh
````


## Undeploying the samples
From the `mq-package` directory run

````
./undeploy.sh
````

## Monitoring the feed and actions
You can monitor the console output from the feed and actions by running.

````
ibmcloud fn activation poll
````



---
