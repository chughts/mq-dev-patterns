/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the 'License');
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"mq-golang/ibmmq"
	"mqsamputils"
	"os"
	"strings"
	"time"
)

var logger = log.New(os.Stdout, "MQ Response: ", log.LstdFlags)

type message struct {
	Greeting  string `json:"greeting"`
	Message   string `json:"message"`
	Value     int    `json:"value"`
	MSGCorrID string `json:"correlationID"`
}

// Main Entry to Put application
// Creates Connection to Queue
func main() {

	logger.Println("Application is Starting")

	logSettings()

	mqsamputils.EnvSettings.LogSettings()

	qMgr, err := mqsamputils.CreateConnection()
	if err != nil {
		logger.Fatalln("Unable to Establish Connection to server")
		os.Exit(1)
	}
	defer qMgr.Disc()

	qObject, err := mqsamputils.OpenQueue(qMgr, mqsamputils.Get)
	if err != nil {
		logger.Fatalln("Unable to Open Queue")
		os.Exit(1)
	}
	defer qObject.Close(0)

	getMessages(qMgr, qObject)

	logger.Println("Application is Ending")
}

// Output Basic Authentication values to verify that they have
// been read from the envrionment settings
func logSettings() {
	logger.Printf("Username is (%s)\n", mqsamputils.EnvSettings.User)
	logger.Printf("Password is (%s)\n", mqsamputils.EnvSettings.Password)
}

func logError(err error) {
	logger.Println(err)
	logger.Printf("Error Code %v", err.(*ibmmq.MQReturn).MQCC)
}

func getMessages(qMgr ibmmq.MQQueueManager, qObject ibmmq.MQObject) {
	logger.Println("Getting Message from Queue")
	var err error
	msgAvail := true

	for msgAvail == true && err == nil {
		var datalen int

		// The PUT requires control structures, the Message Descriptor (MQMD)
		// and Put Options (MQPMO). Create those with default values.
		getmqmd := ibmmq.NewMQMD()
		gmo := ibmmq.NewMQGMO()

		// The default options are OK, but it's always
		// a good idea to be explicit about transactional boundaries as
		// not all platforms behave the same way.
		gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT | ibmmq.MQGMO_WAIT | ibmmq.MQGMO_FAIL_IF_QUIESCING

		// Set options to wait for a maximum of 3 seconds for any new message to arrive

		gmo.WaitInterval = 3 * 1000 // The WaitInterval is in milliseconds

		// Create a buffer for the message data. This one is large enough
		// for the messages put by the amqsput sample.
		buffer := make([]byte, 1024)

		// Now we can try to get the message
		datalen, err = qObject.Get(getmqmd, gmo, buffer)

		if err != nil {
			msgAvail = false
			logger.Println(err)
			mqret := err.(*ibmmq.MQReturn)
			logger.Printf("return code %d, expected %d,", mqret.MQRC, ibmmq.MQRC_NO_MSG_AVAILABLE)
			if mqret.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
				// If there's no message available, then I won't treat that as a real error as
				// it's an expected situation
				msgAvail = true
				err = nil
			}
		} else {
			// Assume the message is a printable string, which it will be
			// if it's been created by the amqsput program
			logger.Printf("Got message of length %d: ", datalen)
			logger.Println(string(buffer[:datalen]))
			/*      if getmqmd.Format == "MQSTR" {
			          replyToMsg(string(buffer[:datalen]))
			        } else {
			          logger.Println("Message in an unexpected format")
			        } */

			qObject, err := mqsamputils.OpenDynamicQueue(qMgr, getmqmd.ReplyToQ)
			if err != nil {
				logger.Fatalln("Unable to Open Queue")
				os.Exit(1)
			}
			defer qObject.Close(0)

			replyToMsg(qObject, string(buffer[:datalen]), getmqmd)
		}
	}
}

/*
RFH
?MQSTR
? <mcd>
<Msd>jms_text</Msd>
</mcd>
?<jms><Dst>queue:///DEV.QUEUE.1</Dst>
<Rto>queue://QM1/AMQ.5C6AAC7D23035202
?persistence=1</Rto>
<Tms>1550771520920</Tms>
<Cid>f261990d-b631-4b69-9e9c-ee239580859d</Cid>
<Dlv>2</Dlv>
</jms>
{"message":"The number is:","value":3}
*/

func replyToMsg(qObject ibmmq.MQObject, msg string, getmqmd *ibmmq.MQMD) {
	logger.Println("About to reply to request ", msg)
	var messageObject message

	/*  if p := strings.Index(msg, "</jms>"); p > 0 {
	    logger.Println("Whole JMS message found ", p)
	    msg = msg[p+6:]
	  } */
	logger.Println("Pruned message ", msg)

	json.Unmarshal([]byte(msg), &messageObject)
	logger.Println("Found message", messageObject.Greeting)
	logger.Println("Found message", messageObject.Message)
	logger.Println("Found message", messageObject.Value)

	msgData := &message{
		Greeting: "Reply from Go is " + time.Now().Format(time.RFC3339),
		Value:    messageObject.Value * messageObject.Value}
	data, err := json.Marshal(msgData)
	if err != nil {
		logger.Println("Unexpected error marshalling data to send")
		logError(err)
		return
	}

	putmqmd := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()

	emptyByteArray := make([]byte, 24)

	if bytes.Equal(getmqmd.CorrelId, emptyByteArray) || bytes.Contains(getmqmd.CorrelId, emptyByteArray) {
		logger.Println("CorrelId is empty")
		putmqmd.CorrelId = getmqmd.MsgId
	} else {
		logger.Println("Correl ID found on request")

		putmqmd.CorrelId = getmqmd.CorrelId
	}

	putmqmd.MsgId = getmqmd.MsgId

	// Tell MQ what the message body format is.
	// In this case, a text string
	putmqmd.Format = ibmmq.MQFMT_STRING

	logger.Println("Looking for match on Correl ID CorrelID:" + hex.EncodeToString(putmqmd.CorrelId))

	logger.Println("Looking for match on Correl ID CorrelID:" + string(putmqmd.CorrelId))

	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT

	// Now put the message to the queue
	logger.Printf("Sending message %s", data)
	err = qObject.Put(putmqmd, pmo, data)

	if err != nil {
		logError(err)
	} else {
		logger.Println("Put message to", strings.TrimSpace(qObject.Name))
		// Print the MsgId so it can be used as a parameter to amqsget
		logger.Println("MsgId:" + hex.EncodeToString(putmqmd.MsgId))
	}
}
