package api

import (
	"encoding/json"
	"fmt"
	"github.com/Bnei-Baruch/uptime-mon/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

var MQTT mqtt.Client
var STATUS map[string]string

type MqttPayload struct {
	Action  string      `json:"action,omitempty"`
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Source  string      `json:"src,omitempty"`
	Error   error       `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Result  string      `json:"result,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type JanusStatus struct {
	Online bool `json:"online,omitempty"`
}

type PahoLogAdapter struct {
	level log.Level
}

func NewPahoLogAdapter(level log.Level) *PahoLogAdapter {
	return &PahoLogAdapter{level: level}
}

func (a *PahoLogAdapter) Println(v ...interface{}) {
	log.Infof("MQTT: %s", fmt.Sprint(v...))
}

func (a *PahoLogAdapter) Printf(format string, v ...interface{}) {
	log.Infof("MQTT: %s", fmt.Sprintf(format, v...))
}

func InitMQTT() error {
	log.Info("MQTT: Init")
	STATUS = make(map[string]string)
	//mqtt.DEBUG = NewPahoLogAdapter(log.DebugLevel)
	//mqtt.WARN = NewPahoLogAdapter(log.WarnLevel)
	//mqtt.CRITICAL = NewPahoLogAdapter(log.PanicLevel)
	//mqtt.ERROR = NewPahoLogAdapter(log.ErrorLevel)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(viper.GetString("mqtt.url"))
	opts.SetClientID(viper.GetString("mqtt.client_id"))
	opts.SetUsername(viper.GetString("mqtt.user"))
	opts.SetPassword(viper.GetString("mqtt.password"))
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(SubMQTT)
	opts.SetConnectionLostHandler(LostMQTT)
	opts.SetBinaryWill(viper.GetString("mqtt.status_topic"), []byte("Offline"), byte(2), true)
	MQTT = mqtt.NewClient(opts)
	if token := MQTT.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func SubMQTT(c mqtt.Client) {
	if token := MQTT.Publish(viper.GetString("mqtt.status_topic"), byte(1), true, []byte("Online")); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: notify status error: %s", token.Error())
	} else {
		log.Infof("MQTT: notify status to: %s", viper.GetString("mqtt.status_topic"))
	}

	execStatus := viper.GetString("mqtt.exec_status")
	if token := MQTT.Subscribe(execStatus, byte(1), gotMessage); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: Subscribe error: %s", token.Error())
	} else {
		log.Infof("MQTT: Subscribed to: %s", execStatus)
	}

	wfStatus := viper.GetString("mqtt.wf_status")
	if token := MQTT.Subscribe(wfStatus, byte(1), gotMessage); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: Subscribe error: %s", token.Error())
	} else {
		log.Infof("MQTT: Subscribed to: %s", wfStatus)
	}

	janusStatus := viper.GetString("mqtt.janus_status")
	if token := MQTT.Subscribe(janusStatus, byte(1), gotMessage); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: Subscribe error: %s", token.Error())
	} else {
		log.Infof("MQTT: Subscribed to: %s", janusStatus)
	}
}

func LostMQTT(c mqtt.Client, err error) {
	log.Errorf("MQTT: Lost connection: %s", err)
}

func gotMessage(c mqtt.Client, m mqtt.Message) {
	log.Debugf("MQTT: Received message %s from topic: %s\n", m.Topic(), string(m.Payload()))
	var serviceName string
	var serviceStatus string

	s := strings.Split(m.Topic(), "/")
	if s[2] == "status" {
		serviceName = s[1]
		js := &JanusStatus{}
		st := "Offline"
		err := json.Unmarshal(m.Payload(), js)
		if err != nil {
			log.Errorf("MQTT: Message parsing error: %s", err)
		}
		if js.Online {
			st = "Online"
		}
		serviceStatus = st
		STATUS[s[1]] = st
	} else {
		serviceName = s[2]
		serviceStatus = string(m.Payload())
		STATUS[s[2]] = string(m.Payload())
	}

	STATUS[serviceName] = serviceStatus

	if serviceStatus == "Offline" {
		log.Infof("MQTT: Service %s is: %s\n", serviceName, serviceStatus)
		//utils.SendEmail(s[2], "Offline")
		utils.SendSlack(serviceName + " - Offline")
	}
}

func SendRespond(id string, m *MqttPayload) {
	var topic string

	if id == "false" {
		topic = viper.GetString("mqtt.srv_topic")
	} else {
		topic = viper.GetString("mqtt.srv_topic") + "/" + id
	}
	message, err := json.Marshal(m)
	if err != nil {
		log.Errorf("MQTT: Message parsing error: %s", err)
	}

	if token := MQTT.Publish(topic, byte(1), false, message); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: Publish error: %s, reason: %s", topic, token.Error())
	}

	log.Debugf("MQTT: Send message: %s to topic: %s\n", string(message), topic)
}

func SendMessage(id string) {
	var topic string
	var m interface{}
	//date := time.Now().Format("2006-01-02")

	message, err := json.Marshal(m)
	if err != nil {
		log.Errorf("MQTT: Message parsing error: %s", err)
	}

	if token := MQTT.Publish(topic, byte(1), true, message); token.Wait() && token.Error() != nil {
		log.Errorf("MQTT: Publish error: %s, reason: %s", topic, token.Error())
	}

	log.Debugf("MQTT: Send message from: %s to topic: %s\n", id, topic)
}
